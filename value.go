package grok

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

var (
	dumpMutex = sync.Mutex{}
)

type value struct {
	Name        string
	IsValid     bool
	IsPointer   bool
	IsInterface bool
	Type        string
	RValue      reflect.Value
	RType       reflect.Type
	Elem        interface{}
	Children    []value
}

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

	// create a buffer for our output to make it concurrency safe.
	var b bytes.Buffer

	// dump it like you mean it
	dump("value", reflect.ValueOf(value), writer(&b), colourizer(c.colour), indenter(c.tabstop))

	// Write the contents of the buffer to the configured writer.
	c.writer.Write(b.Bytes())
}

func dump(name string, v reflect.Value, write Writer, colour Colourizer, indent Indenter) {
	val := value{
		Name:   name,
		RValue: v,
	}

	if !v.IsValid() {
		val.Elem = "<invalid>"
		write(indent("<invalid>\n", c.depth))
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
		write(indent("%s %s = ", c.depth), colour(name, colourYellow), tn)
	} else {
		write(indent("", c.depth))
	}

	switch v.Kind() {
	case reflect.Bool:
		write(colour("%v", colourGreen), v.Bool())
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
		write(colour("%v", colourGreen), v)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Len() == 0 {
			write("[]\n")
			return
		}
		write("[\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), c.depth))
		} else {
			for i := 0; i < v.Len(); i++ {
				dump(colour(fmt.Sprintf("%d", i), colourRed), v.Index(i), write, colour, indent)
			}
		}
		c.depth = c.depth - 1
		write(indent("]", c.depth))
	case reflect.Chan:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	case reflect.Func:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	case reflect.Map:
		if !v.IsValid() {
			write(colour("<nil>", colourGrey))
		} else {
			if v.Len() == 0 {
				write("[]\n")
				return
			}
			write("[\n")
			c.depth = c.depth + 1
			if c.maxDepth > 0 && c.depth >= c.maxDepth {
				write(indent(colour("... max depth reached\n", colourGrey), c.depth))
			} else {
				keys := v.MapKeys()
				sort.Sort(Values(keys))
				for _, k := range v.MapKeys() {
					dump(fmt.Sprintf("%v", k), v.MapIndex(k), write, colour, indent)
				}
			}
			c.depth = c.depth - 1
			write(indent("]", c.depth))
		}
	case reflect.String:
		s := v.String()
		slen := len(s)
		if c.maxLength > 0 && slen > c.maxLength {
			s = fmt.Sprintf("%s...", string([]byte(s)[0:c.maxLength]))
		}
		write(colour("%q ", colourGreen), s)
		write(colour("%d", colourGrey), slen)
	case reflect.Struct:
		if v.NumField() == 0 {
			write("{}\n")
			return
		}
		write("{\n")
		c.depth = c.depth + 1
		if c.maxDepth > 0 && c.depth >= c.maxDepth {
			write(indent(colour("... max depth reached\n", colourGrey), c.depth))
		} else {
			for i := 0; i < v.NumField(); i++ {
				dump(t.Field(i).Name, v.Field(i), write, colour, indent)
			}
		}
		c.depth = c.depth - 1
		write(indent("}", c.depth))
	case reflect.UnsafePointer:
		write(colour("%v", colourGreen), v)
	case reflect.Invalid:
		write(colour("<nil>", colourGrey))
	case reflect.Interface:
		if v.IsNil() {
			write(colour("<nil>", colourGrey))
		} else {
			write(colour("%v", colourGreen), v)
		}
	default:
		write(colour("??? %s", colourRed), v.Kind().String())
	}
	write("\n")
}
