package driver

import (
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	"reflect"
)

type ConfigMap map[string]string

func GetDriverConfig() ConfigMap {
	return device.DriverConfigs()
}

func GetConfig(name string, v interface{}) error {
	return GetDriverConfig().Get(name, v)
}

func (cm ConfigMap) Get(name string, v interface{}) error {
	cv := reflect.ValueOf(v)
	if !cv.IsValid() || cv.IsNil() || cv.Kind() != reflect.Ptr {
		panic(fmt.Errorf("invalid destination for configuration %q", name))
	}

	cv = reflect.Indirect(cv)
	if !cv.CanSet() {
		panic(fmt.Errorf("cannot set value for configuration %q", name))
	}

	strVal, ok := cm[name]
	if !ok {
		return fmt.Errorf("missing configuration for %q", name)
	}

	if strVal == "" {
		return fmt.Errorf("configuration for %q is empty", name)
	}

	switch cv.Kind() {
	case reflect.String:
		cv.SetString(strVal)
	default:
		return fmt.Errorf("unsupported property kind %v for field %q", cv.Kind(), name)
	}
	return nil
}
