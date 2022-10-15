package internal

import (
	"net/url"
	"strconv"
)

type Values struct {
	url.Values
}

func NewValues(values url.Values) *Values {
	if values == nil {
		values = url.Values{}
	}
	return &Values{Values: values}
}

func (v Values) SetPtr(key string, s *string) {
	if s != nil {
		v.Set(key, *s)
	}
}

func (v Values) SetBool(key string, b bool) {
	v.Set(key, strconv.FormatBool(b))
}

func (v Values) SetBoolPtr(key string, b *bool) {
	if b != nil {
		v.SetBool(key, *b)
	}
}
func (v Values) SetInt(key string, i int) {
	v.Set(key, strconv.Itoa(i))
}

func (v Values) SetIntPtr(key string, i *int) {
	if i != nil {
		v.SetInt(key, *i)
	}
}

func (v Values) SetInt64(key string, i int64) {
	v.Set(key, strconv.FormatInt(i, 10))
}

func (v Values) SetInt64Ptr(key string, i *int64) {
	if i != nil {
		v.SetInt64(key, *i)
	}
}
