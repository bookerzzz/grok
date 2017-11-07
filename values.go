package grok

import (
	"fmt"
	"reflect"
)

type Values []reflect.Value

// Len implements sort.Interface
func (s Values) Len() int {
	return len(s)
}

// Swap implements sort.Interface
func (s Values) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less implements sort.Interface
func (s Values) Less(i, j int) bool {
	return fmt.Sprintf("%v", s[i]) < fmt.Sprintf("%v", s[j])
}
