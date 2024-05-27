package cfgmgr

import (
	"sync/atomic"
	"unsafe"
)

type Config[T any] struct {
	ptr unsafe.Pointer // *T
}

func (c *Config[T]) Get() *T {
	return (*T)(atomic.LoadPointer(&c.ptr))
}

func (c *Config[T]) Set(t *T) {
	atomic.StorePointer(&c.ptr, unsafe.Pointer(t))
}

type unmarshaler func(in []byte, out any) (err error)

// xxconfig := Config[T]{}
// ConfigManager.Watch("config.json", AtomicLoad(&xxconfig, json.Unmarshal))
// cfg := xxconfig.Get()
func AtomicLoad[T any](dest *Config[T], unmarshalFn unmarshaler) LoadFunc {
	var dummy T
	dest.Set(&dummy)

	return func(buf []byte) error {

		var value T

		err := unmarshalFn(buf, &value)
		if err != nil {
			return err
		}

		dest.Set(&value)

		return nil
	}
}
