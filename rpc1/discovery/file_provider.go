package discovery

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/dogsays/mo/logger"
	"github.com/dogsays/mo/ut2"
)

type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

type FilePortProvider struct {
	config ut2.IMap[string, string]
}

func NewFilePortProvider(filename string) *FilePortProvider {
	fd := &FilePortProvider{}
	fd.config = ut2.NewSyncMap[string, string]()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		logger.Err(err)
	}
	err = fd.loadConfig(bytes)
	if err != nil {
		logger.Err(err)
	}
	return fd
}

func (am *FilePortProvider) Close() {
}

func (am *FilePortProvider) GetPort(service string) (int, error) {
	arr, _ := am.GetAddr(service)
	if len(arr) != 0 {
		addr := arr[0]
		idx := strings.IndexByte(addr, ':')
		if idx != -1 {
			return strconv.Atoi(addr[idx+1:])
		}
	}

	return 0, errors.New("no config for" + service)
}

func (am *FilePortProvider) GetAddr(service string) ([]string, error) {

	addr, ok := am.config.Load(service)
	if ok {
		return []string{addr}, nil
	}
	return nil, nil
}

func (f *FilePortProvider) loadConfig(buf []byte) error {
	tmp := map[string]any{}
	err := json.Unmarshal(buf, &tmp)
	if err != nil {
		return err
	}

	f.config.Clear()
	for k, v := range tmp {
		switch dest := v.(type) {
		case string:
			f.config.Store(k, dest)
		case []any:
			for _, one := range dest {
				f.config.Store(k, one.(string))
			}
		default:
			logger.Info("类型错误", dest)
		}
	}

	return nil
}
