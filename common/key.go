package common

import (
	"crypto/rand"
	"io/ioutil"
	"os"
)

func LoadOrGenerateKey(filename string, size int) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		return data, nil
	}

	key := make([]byte, size)
	_, err = rand.Read(key)
	if err != nil {
		return nil, err
	}

	_ = ioutil.WriteFile(filename, key, os.ModePerm)
	return key, nil
}
