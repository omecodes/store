package oms

import (
	"encoding/json"
	"errors"
	"github.com/PaesslerAG/jsonpath"
	"strings"
)

type JSON struct {
	object interface{}
}

func NewJSON(content string) (*JSON, error) {
	j := &JSON{}
	return j, json.Unmarshal([]byte(content), &j.object)
}

func at(jp string) string {
	jp = strings.Replace(jp, "/", ".", -1)
	if strings.HasPrefix(jp, "$.") {
		return jp
	}
	if strings.HasPrefix(jp, ".") {
		return "$" + jp
	}
	return "$." + jp
}

func (s *JSON) ToInt64() (int64, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int64(v), nil
}

func (s *JSON) ToInt32() (int32, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int32(v), nil
}

func (s *JSON) ToInt() (int, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int(v), nil
}

func (s *JSON) ToString() (string, error) {
	v, ok := s.object.(string)
	if !ok {
		return "", errors.New("settings value type conversion")
	}

	return v, nil
}

func (s *JSON) Int64(name string) (int64, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return 0, err
	}

	v, ok := o.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int64(v), nil
}

func (s *JSON) Int32(name string) (int32, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return 0, err
	}

	v, ok := o.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int32(v), nil
}

func (s *JSON) Int(name string) (int, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return 0, err
	}

	v, ok := o.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int(v), nil
}

func (s *JSON) StringAt(name string) (string, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return "", err
	}

	v, ok := o.(string)
	if !ok {
		return "", errors.New("settings value type conversion")
	}

	return v, nil
}

func (s *JSON) Get(name string) (*JSON, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return nil, err
	}
	return &JSON{object: o}, nil
}

func (s *JSON) String() string {
	data, _ := json.Marshal(s.object)
	return string(data)
}
