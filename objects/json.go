package objects

import (
	"encoding/json"
	"errors"
	"github.com/PaesslerAG/jsonpath"
	"strings"
)

type JsonObject struct {
	object interface{}
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

func (s *JsonObject) ToInt64() (int64, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int64(v), nil
}

func (s *JsonObject) ToInt32() (int32, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int32(v), nil
}

func (s *JsonObject) ToInt() (int, error) {
	v, ok := s.object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int(v), nil
}

func (s *JsonObject) ToString() (string, error) {
	v, ok := s.object.(string)
	if !ok {
		return "", errors.New("settings value type conversion")
	}

	return v, nil
}

func (s *JsonObject) Int64(name string) (int64, error) {
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

func (s *JsonObject) Int32(name string) (int32, error) {
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

func (s *JsonObject) Int(name string) (int, error) {
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

func (s *JsonObject) StringAt(name string) (string, error) {
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

func (s *JsonObject) BoolAt(name string) (bool, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return false, err
	}

	v, ok := o.(bool)
	if !ok {
		return false, errors.New("settings value type conversion")
	}

	return v, nil
}

func (s *JsonObject) Get(name string) (*JsonObject, error) {
	name = at(name)
	o, err := jsonpath.Get(name, s.object)
	if err != nil {
		return nil, err
	}
	return &JsonObject{object: o}, nil
}

func (s *JsonObject) GetObject() interface{} {
	return s.object
}

func (s *JsonObject) String() string {
	data, _ := json.Marshal(s.object)
	return string(data)
}

func (s *JsonObject) Marshal() ([]byte, error) {
	return json.Marshal(s.object)
}
