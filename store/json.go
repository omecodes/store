package store

import (
	"errors"
	"github.com/PaesslerAG/jsonpath"
	"strings"
)

type JSON struct {
	Object interface{}
}

func at(jsonPath string) string {
	if strings.HasPrefix(jsonPath, "$.") {
		return jsonPath
	}
	if strings.HasPrefix(jsonPath, ".") {
		return "$" + jsonPath
	}
	return "$." + jsonPath
}

func (s *JSON) ToInt64() (int64, error) {
	v, ok := s.Object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int64(v), nil
}

func (s *JSON) ToInt32() (int32, error) {
	v, ok := s.Object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int32(v), nil
}

func (s *JSON) ToInt() (int, error) {
	v, ok := s.Object.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int(v), nil
}

func (s *JSON) ToString() (string, error) {
	v, ok := s.Object.(string)
	if !ok {
		return "", errors.New("settings value type conversion")
	}

	return v, nil
}

func (s *JSON) Int64(name string) (int64, error) {
	name = at(strings.Replace(name, "/", ".", -1))
	o, err := jsonpath.Get(name, s.Object)
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
	name = at(strings.Replace(name, "/", ".", -1))
	o, err := jsonpath.Get(name, s.Object)
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
	name = at(strings.Replace(name, "/", ".", -1))
	o, err := jsonpath.Get(name, s.Object)
	if err != nil {
		return 0, err
	}

	v, ok := o.(float64)
	if !ok {
		return 0, errors.New("settings value type conversion")
	}

	return int(v), nil
}

func (s *JSON) String(name string) (string, error) {
	name = at(strings.Replace(name, "/", ".", -1))
	o, err := jsonpath.Get(name, s.Object)
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
	name = at(strings.Replace(name, "/", ".", -1))
	o, err := jsonpath.Get(name, s.Object)
	if err != nil {
		return nil, err
	}
	return &JSON{Object: o}, nil
}
