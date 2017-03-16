package grok

import (
	"reflect"
	"fmt"
	"sort"
	"sync"
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

var (
	dumpMutex = sync.Mutex{}
)

// V aliases Value
func V(value interface{}, options ...Option) {
	Value(value, options...)
}

// Value prints a more human readable representation of any value
func Value(value interface{}, options ...Option) {
	c := defaults
	for _, o := range options {
		o(&c)
	}

	// only dump one value at a time to avoid overlap
	dumpMutex.Lock()
	defer dumpMutex.Unlock()

	// dump it like you mean it
	dump("value", reflect.ValueOf(value), c)
}

func dump(name string, v reflect.Value, c Conf) {
	out := outer(c)
	colour := colourer(c)

	if !v.IsValid() {
		out(indent(c) + "<invalid>\n")
		return
	}
	t := v.Type()

	tn := ""
	switch t.Kind() {
	case reflect.Interface:
		name := t.Name()
		tn = colour("interface{}", colourBlue) + colour(name, colourBlue)
		if !v.IsNil() {
			v = v.Elem()
			t = v.Type()
		}
		if t.Kind() == reflect.Ptr {
			v = v.Elem()
			t = v.Type()
		}
		if t.Name() != name {
			tn = colour(t.Name(), colourBlue) + colour(" as ", colourRed) + tn
		}
	case reflect.Ptr:
		v = v.Elem()
		t = t.Elem()
		tn = colour("*", colourRed) + colour(t.Name(), colourBlue)
	case reflect.Slice:
		tn = colour("[]", colourRed)
		switch t.Elem().Kind() {
		case reflect.Interface:
			tn = tn + colour(t.Elem().Name(), colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Elem().Elem().Name(), colourBlue)
		default:
			tn = tn + colour(t.Elem().Name(), colourBlue)
		}
	case reflect.Map:
		tn = colour("map[", colourRed)
		switch t.Key().Kind() {
		case reflect.Interface:
			tn = tn + colour(t.Key().Name(), colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Key().Elem().Name(), colourBlue)
		default:
			tn = tn + colour(t.Key().Name(), colourBlue)
		}
		tn = tn + colour("]", colourRed)
		switch t.Elem().Kind() {
		case reflect.Interface:
			tn = tn + colour("interface{}", colourBlue)
		case reflect.Ptr:
			tn = tn + colour("*", colourRed)
			tn = tn + colour(t.Elem().Elem().Name(), colourBlue)
		default:
			tn = tn + colour(t.Elem().Name(), colourBlue)
		}
	case reflect.Chan:
		tn = colour(t.ChanDir().String(), colourRed)
		tn = tn + " " + colour(t.Elem().Name(), colourBlue)
	case reflect.Func:
		tn = colour("func", colourRed)
	case reflect.UnsafePointer:
		tn = colour("unsafe*", colourRed) + colour(t.Name(), colourBlue)
	default:
		tn = colour(t.Name(), colourBlue)
	}

	if len(name) > 0 {
		out(indent(c) + "%s %s = ", colour(name, colourYellow), tn)
	} else {
		out(indent(c))
	}

	switch v.Kind() {
	case reflect.Bool:
		out(colour("%v", colourGreen), v.Bool())
	case reflect.Uintptr:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		out(colour("%v", colourGreen), v)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Len() == 0 {
			out("[]\n")
			return
		}
		out("[\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			out(indent(c) + colour("... max depth reached\n", colourGrey))
		} else {
			for i := 0; i < v.Len(); i++ {
				dump(colour(fmt.Sprintf("%d", i), colourRed), v.Index(i), c)
			}
		}
		c.depth = c.depth - 1
		out(indent(c) + "]")
	case reflect.Chan:
		if v.IsNil() {
			out(colour("<nil>", colourGrey))
		} else {
			out(colour("%v", colourGreen), v)
		}
	case reflect.Func:
		if v.IsNil() {
			out(colour("<nil>", colourGrey))
		} else {
			out(colour("%v", colourGreen), v)
		}
	case reflect.Map:
		if !v.IsValid() {
			out(colour("<nil>", colourGrey))
		} else {
			if v.Len() == 0 {
				out("[]\n")
				return
			}
			out("[\n")
			c.depth = c.depth + 1
			if c.maxDepth > 0 && c.depth >= c.maxDepth {
				out(indent(c) + colour("... max depth reached\n", colourGrey))
			} else {
				keys := v.MapKeys();
				sort.Sort(Values(keys))
				for _, k := range v.MapKeys() {
					dump(fmt.Sprintf("%v", k), v.MapIndex(k), c)
				}
			}
			c.depth = c.depth - 1
			out(indent(c) + "]")
		}
	case reflect.String:
		s := v.String()
		slen := len(s)
		if c.maxLength > 0 && slen > c.maxLength {
			s = fmt.Sprintf("%s...", string([]byte(s)[0:c.maxLength]))
		}
		out(colour("%q ", colourGreen), s)
		out(colour("%d", colourGrey), slen)
	case reflect.Struct:
		if v.NumField() == 0 {
			out("{}\n")
			return
		}
		out("{\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			out(indent(c) + colour("... max depth reached\n", colourGrey))
		} else {
			for i := 0; i < v.NumField(); i++ {
				dump(t.Field(i).Name, v.Field(i), c)
			}
		}
		c.depth = c.depth - 1
		out(indent(c) + "}")
	case reflect.UnsafePointer:
		out(colour("%v", colourGreen), v)
	case reflect.Invalid:
		out(colour("<nil>", colourGrey))
	case reflect.Interface:
		if v.IsNil() {
			out(colour("<nil>", colourGrey))
		} else {
			out(colour("%v", colourGreen), v)
		}
	default:
		out(colour("??? %s", colourRed), v.Kind().String())
	}
	out("\n")
}