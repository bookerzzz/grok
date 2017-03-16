package grok

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	colourReset  = "\x1B[0m"
	colourRed    = "\x1B[38;5;124m"
	colourYellow = "\x1B[38;5;208m"
	colourBlue   = "\x1B[38;5;33m"
	colourGrey   = "\x1B[38;5;144m"
	colourGreen  = "\x1B[38;5;34m"
	colourGold   = "\x1B[38;5;3m"
	defaults = Conf{
		depth:   0,
		out:     os.Stdout,
		tabstop: 4,
		colour: true,
		maxDepth: 10,
		maxLength: 100,
	}
)

type Option func(c *Conf)

type Conf struct {
	depth   int
	out     io.Writer
	tabstop int
	colour bool
	maxDepth int
	maxLength int
}


func outer(c Conf) func(string, ...interface{}) {
	return func(format string, params ...interface{}) {
		c.out.Write([]byte(fmt.Sprintf(format, params...)))
	}
}

func indent(c Conf) string {
	return strings.Repeat(" ", c.depth*c.tabstop)
}

func colourer(c Conf) func(str string, colour string) string {
	return func (str string, colour string) string {
		if c.colour {
			return colour + str + colourReset
		}
		return str
	}
}

// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option {
	return func(c *Conf) {
		c.out = w
	}
}

// WithoutColours disables colouring of output from debug functions. Defaults to `true`
func WithoutColours() Option {
	return func(c *Conf) {
		c.colour = false
	}
}

// WithMaxDepth sets the maximum recursion depth from debug functions. Defaults to `10`, use `0` for unlimited
func WithMaxDepth(depth int) Option {
	return func(c *Conf) {
		c.maxDepth = depth
	}
}

// WithMaxLength sets the maximum length of string values. Default is `100`, use `0` for unlimited
func WithMaxLength(chars int) Option {
	return func(c *Conf) {
		c.maxLength = chars
	}
}

// WithTabStop sets the width of a tabstop to the given char count. Defaults to `4`
func WithTabStop(chars int) Option {
	return func(c *Conf) {
		c.tabstop = chars
	}
}
