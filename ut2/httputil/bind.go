package httputil

import (
	"encoding"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type (
	HTTPError struct {
		Code    int
		Message interface{}
	}
)

func (he *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", he.Code, he.Message)
}
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

func HttpBindQuery(i interface{}, req *http.Request) (err error) {
	if err = BindData(i, req.URL.Query(), "query"); err != nil {
		return NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return
}

// Bind implements the `Binder#Bind` function.
func HttpBind(i interface{}, req *http.Request) (err error) {
	if req.ContentLength == 0 {
		if req.Method == "GET" {
			if err = BindData(i, req.URL.Query(), "query"); err != nil {
				return NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return
		}
		return NewHTTPError(http.StatusBadRequest, "Request body can't be empty")
	}
	ctype := req.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ctype, "application/json"):
		if err = json.NewDecoder(req.Body).Decode(i); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, offset=%v", ute.Type, ute.Value, ute.Offset))
			} else if se, ok := err.(*json.SyntaxError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()))
			} else {
				return NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}
	case strings.HasPrefix(ctype, "application/xml"), strings.HasPrefix(ctype, "text/xml"):
		if err = xml.NewDecoder(req.Body).Decode(i); err != nil {
			if ute, ok := err.(*xml.UnsupportedTypeError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unsupported type error: type=%v, error=%v", ute.Type, ute.Error()))
			} else if se, ok := err.(*xml.SyntaxError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: line=%v, error=%v", se.Line, se.Error()))
			} else {
				return NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}
	case strings.HasPrefix(ctype, "application/x-www-form-urlencoded"), strings.HasPrefix(ctype, MIMEMultipartForm):
		params, err := formParams(req)
		if err != nil {
			return NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err = BindData(i, params, "form"); err != nil {
			return NewHTTPError(http.StatusBadRequest, err.Error())
		}
	default:
		return NewHTTPError(http.StatusUnsupportedMediaType)
	}
	return
}

const MIMEMultipartForm = "multipart/form-data"
const defaultMemory = 32 << 20 // 32 MB
func formParams(req *http.Request) (url.Values, error) {
	if strings.HasPrefix(req.Header.Get("Content-Type"), MIMEMultipartForm) {
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	return req.Form, nil
}

func UnmarshalQuery(destination any, queryString, tag string) error {
	query, err := url.ParseQuery(queryString)
	if err != nil {
		return err
	}

	return BindData(destination, query, tag)
}

// bindData will bind data ONLY fields in destination struct that have EXPLICIT tag
func BindData(destination interface{}, data map[string][]string, tag string) error {
	if destination == nil || len(data) == 0 {
		return nil
	}
	typ := reflect.TypeOf(destination).Elem()
	val := reflect.ValueOf(destination).Elem()

	// Map
	if typ.Kind() == reflect.Map {
		for k, v := range data {
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}
		return nil
	}

	// !struct
	if typ.Kind() != reflect.Struct {
		if tag == "param" || tag == "query" || tag == "header" {
			// incompatible type, data is probably to be found in the body
			return nil
		}
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if typeField.Anonymous {
			if structField.Kind() == reflect.Ptr {
				structField = structField.Elem()
			}
		}
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)
		if typeField.Anonymous && structField.Kind() == reflect.Struct && inputFieldName != "" {
			// if anonymous struct with query/param/form tags, report an error
			return errors.New("query/param/form tags are not allowed with anonymous struct field")
		}

		// if inputFieldName == "" {
		// 	// If tag is nil, we inspect if the field is a not BindUnmarshaler struct and try to bind data into it (might contains fields with tags).
		// 	// structs that implement BindUnmarshaler are binded only when they have explicit tag
		// 	// if _, ok := structField.Addr().Interface().(BindUnmarshaler); !ok && structFieldKind == reflect.Struct {
		// 	// 	if err := BindData(structField.Addr().Interface(), data, tag); err != nil {
		// 	// 		return err
		// 	// 	}
		// 	// }
		// 	// does not have explicit tag and is not an ordinary struct - so move to next field
		// 	continue
		// }

		if inputFieldName == "" {
			inputFieldName = typeField.Name
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			// Go json.Unmarshal supports case insensitive binding.  However the
			// url params are bound case sensitive which is inconsistent.  To
			// fix this we must check all of the map values in a
			// case-insensitive search.
			for k, v := range data {
				if strings.EqualFold(k, inputFieldName) {
					inputValue = v
					exists = true
					break
				}
			}
		}

		if !exists {
			continue
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(typeField.Type.Kind(), inputValue[0], structField); ok {
			if err != nil {
				return err
			}
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
			return err

		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	// But also call it here, in case we're dealing with an array of BindUnmarshalers
	if ok, err := unmarshalField(valueKind, val, structField); ok {
		return err
	}

	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	fieldIValue := field.Addr().Interface()

	if unmarshaler, ok := fieldIValue.(encoding.TextUnmarshaler); ok {
		return true, unmarshaler.UnmarshalText([]byte(value))
	}

	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		// Initialize the pointer to a nil value
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
