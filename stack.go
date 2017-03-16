package grok

import (
	"runtime/debug"
	"bytes"
	"bufio"
	"os"
	"strings"
	"io/ioutil"
	"io"
	"sync"
	"fmt"
	"regexp"
	"strconv"
	"os/exec"
	"unsafe"
	"reflect"
)

const (
	newline = '\n'
	tab = '\t'
	help = `Interactive stack:

> help (h)          prints this help menu
> exit (x)          exits the interactive stack
> show (?)          repeats the current position
> quick (q)         summary of the entire stack
> next (n)          next item in the stack
> prev (p)          previous item in the stack
> top (t)           go to the top of the stack
> jump (j) {index}  jumps to the entry at the given index
> dump (d) {index}  prints a dump of the value corresponding to {index} of the current stack entry

`
)
var (
	stackMutex = sync.Mutex{}
	rxPointA   = regexp.MustCompile(`^(?P<Package>.+)\.(\(?P<Receiver>([^\)]+)\)\.)(?P<Func>[^\(]+)\((?P<Args>[^\)]*)\)\s*\t(?P<File>[^:]+):(?P<Line>[0-9]+)\s(?P<Identifier>\+0x[0-9a-f]+)$`)
	rxPointB   = regexp.MustCompile(`^(?P<Package>.+)\.(?P<Func>[^\(]+)\((?P<Args>[^\)]*)\)\s*\t(?P<File>[^:]+):(?P<Line>[0-9]+) (?P<Identifier>\+0x[0-9a-f]+)$`)
	rxPointC   = regexp.MustCompile(`^(?P<Package>.+)\.(\(?P<Receiver>([^\)]+)\)\.)(?P<Func>[^\(]+)\s*\t(?P<File>[^:]+):(?P<Line>[0-9]+)\s(?P<Identifier>\+0x[0-9a-f]+)$`)
	rxPointD   = regexp.MustCompile(`^(?P<Package>.+)\.(?P<Func>[^\(]+)\s*\t(?P<File>[^:]+):(?P<Line>[0-9]+)\s(?P<Identifier>\+0x[0-9a-f]+)$`)
)

type point struct {
	Position   string
	Package    string
	Receiver   string
	Func       string
	File       string
	Line       string
	Args       []arg
	Identifier string
}

type arg struct {
	Hex string
	Pointer uintptr
	Value interface{}
	Err error
}

func Stack(options ...Option) {
	c := defaults
	for _, o := range options {
		o(&c)
	}

	// only one stack at a time
	stackMutex.Lock()
	defer stackMutex.Unlock()

	colour := colourer(c)
	out := outer(c)
	points := stack(c)

	for _, p := range points {
		args := ""
		for i, a := range p.Args {
			if i > 0 {
				args = args + ", "
			}
			args = args + colour(a.Hex, colourRed)
		}
		out("%s(%s) %s:%s\n", colour(p.Func, colourGreen), args, colour(p.File, colourBlue), colour(p.Line, colourRed))
	}
}

func stack(c Conf) []point {
	points := []point{}
	stack := debug.Stack()
	r := bytes.NewReader(stack)

	lines := [][]byte{}
	line := []byte{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			lines = append(lines, line)
			break;
		}
		if b == newline {
			lines = append(lines, line)
			line = []byte{}
			continue;
		}
		line = append(line, b)
	}

	for i := 1; i < len(lines); i = i+2 {
		// Skip stack creation lines
		if i > 0 && i <= 6 {
			continue
		}
		// skip empty lines
		if len(lines[i]) == 0 {
			continue
		}
		p := point{
			Position: fmt.Sprintf("%d", len(points)),
		}
		l := append(lines[i], lines[i+1]...)
		for _, rx := range []*regexp.Regexp{rxPointA, rxPointB, rxPointC, rxPointD} {
			if rx.Match(l) {
				matches := rx.FindSubmatch(l)
				names := rx.SubexpNames()
				for k, m := range matches {
					n := names[k]
					switch n {
					case "Package":
						p.Package = string(m)
					case "Receiver":
						p.Receiver = string(m)
					case "Func":
						p.Func = string(m)
					case "Args":
						p.Args = []arg{}
						args := bytes.Split(m, []byte(", "))
						for _, h := range args {
							if len(h) == 0 {
								continue
							}
							a := arg{
								Hex: string(h),
							}
							i, err := strconv.ParseUint(string(h[2:]), 16, 64)
							if err != nil {
								a.Err = err
								p.Args = append(p.Args, a)
								continue
							}
							a.Pointer = uintptr(i)
							//v := reflect.Indirect(reflect.ValueOf((*interface{})(unsafe.Pointer(a.Pointer))))
							v := unsafe.Pointer(a.Pointer)
							if v != nil {
								a.Value = v
							}
							p.Args = append(p.Args, a)
						}
					case "File":
						p.File = string(m)
					case "Line":
						p.Line = string(m)
					case "Identifier":
						p.Identifier = string(m)
					}
				}
			}
		}
		points = append(points, p)
	}
	return points
}

func hijackStdOutput() (original io.Writer, done func()) {
	stdout := *os.Stdout
	stderr := *os.Stderr
	r, w, _ := os.Pipe()
	*os.Stdout = *w
	*os.Stderr = *w

	return &stdout, func() {
		w.Close()
		captured, _ := ioutil.ReadAll(r)
		*os.Stdout = stdout
		*os.Stderr = stderr
		os.Stdout.Write(captured)
	}
}

func InteractiveStack(options ...Option) {
	stackMutex.Lock()
	defer stackMutex.Unlock()

	stdout, done := hijackStdOutput()
	defer done()

	c := defaults
	WithWriter(stdout)(&c)
	for _, o := range options {
		o(&c)
	}
	colour := colourer(c)
	out := outer(c)
	index := 0

	points := stack(c)
	reader := bufio.NewReader(os.Stdin)
	out(colour(help, colourGreen))
	points[index].out(c)
	for {
		out(colour("> ", colourRed))
		input, err := reader.ReadString(newline)
		if err != nil {
			out(colour("Sorry, couldn't pick up what you just put down.\n", colourRed))
			continue
		}

		command := ""
		parts := []string{}
		for _, p := range strings.Split(input, " ") {
			p = strings.TrimSpace(p)
			if len(p) == 0 {
				continue
			}
			parts = append(parts, p)
		}
		if len(parts) > 0 {
			command = parts[0]
		}

		if command == "exit" || command == "x" {
			out(colour("Thank you. Goodbye.\n", colourGreen))
			break;
		}

		switch command {
		case "show", "?":
			points[index].out(c)
		case "top", "t":
			index = 0
			points[index].out(c)
		case "next", "n":
			if index == len(points) - 1 {
				out(colour("You've reached the end!\n", colourRed))
			} else {
				index = index + 1
			}
			points[index].out(c)
		case "prev", "p":
			if index == 0 {
				out(colour("You've reached the top!\n", colourRed))
			} else {
				index = index - 1
			}
			points[index].out(c)
		case "dump", "d":
			if len(parts) <= 0 {
				out(colour("Not quite sure which arg to dump?!\n", colourRed))
				out(colour(help, colourGreen))
				continue
			}
			i, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				out(colour("I don't understand %q. It's not a number.\n", colourRed), i)
				continue
			}
			if int(i) < 0 || int(i) > len(points[index].Args) - 1 {
				out(colour("I can't dump it. It's out of range.\n", colourRed))
				continue
			}
			a := points[index].Args[int(i)]
			dump(a.Hex, reflect.ValueOf(a.Value), c)
		case "quick", "q":
			for pi, p := range points {
				args := ""
				for i, a := range p.Args {
					if i > 0 {
						args = args + ", "
					}
					args = args + colour(a.Hex, colourGreen)
				}
				out("%s %s(%s) %s%s\n",colour(fmt.Sprintf("[%d]", pi), colourRed), colour(p.Func, colourBlue), args, colour(p.File, colourBlue), colour(":" + p.Line, colourRed))
			}
		case "jump", "j":
			if len(parts) <= 0 {
				out(colour("Not quite sure what I need to jump to?!\n", colourRed))
				out(colour(help, colourGreen))
				continue
			}
			i, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				out(colour("I don't understand %q. It's not a number.\n", colourRed), i)
				continue
			}
			if int(i) < 0 || int(i) > len(points) - 1 {
				out(colour("I can't jump that far. It's out of range.\n", colourRed))
				continue
			}
			index = int(i)
			points[index].out(c)
		case "open", "o":
			editor := os.Getenv("EDITOR")
			cmd := exec.Command(editor, points[index].File, "+" + points[index].Line)
			cmd.Stdout = stdout
			cmd.Stderr = stdout
			cmd.Stdin = os.Stdin
			err := cmd.Run()
			if err != nil {
				out(colour(err.Error()+"\n", colourRed))
			}
		default:
			out(colour(help, colourGreen))
		}
	}
}

func (p *point) out(c Conf) {
	colour := colourer(c)
	out := outer(c)

	out("position: %s\n", colour(p.Position, colourRed))
	out("identifier: %s\n", colour(p.Identifier, colourRed))
	out("package: %s\n", colour(p.Package, colourGreen))
	out("receiver: %s\n", colour(p.Receiver, colourGreen))
	out("function: %s\n", colour(p.Func, colourGreen))
	out("args: \n")
	for i, a := range p.Args {
		out(colour("    [%d] %s %+v\n", colourRed), i, colour(a.Hex, colourGreen), a.Value)
	}
	out("file: %s%s\n", colour(p.File, colourBlue), colour(":" + p.Line, colourRed))
}
