package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ListFlag []string

func (f *ListFlag) String() string {
	b, _ := json.Marshal(*f)
	return string(b)
}

func (f *ListFlag) Set(value string) error {
	for _, str := range strings.Split(value, ",") {
		*f = append(*f, str)
	}
	return nil
}

type MapFlag map[string]string

func (f *MapFlag) String() string {
	b, _ := json.Marshal(*f)
	return string(b)
}

func (f *MapFlag) Set(value string) error {
	flagVal := *f
	for _, str := range strings.Split(value, ",") {
		param := strings.Split(str, "=")
		if len(param) == 2 {
			flagVal[param[0]] = param[1]
		} else {
			return fmt.Errorf("invalid argument format for Map")
		}
	}
	*f = flagVal
	return nil
}
