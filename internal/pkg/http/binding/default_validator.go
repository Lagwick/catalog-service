package binding

import (
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

type defaultValidator struct {
	validate *validator.Validate
	once     sync.Once
}

var _ StructValidator = &defaultValidator{}

func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
	})
}

func (v *defaultValidator) Engine() any {
	v.lazyInit()
	return v.validate
}

func (v *defaultValidator) ValidateStruct(obj any) error {
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		v.lazyInit()
		return v.validate.Struct(val.Interface())
	}

	return nil
}
