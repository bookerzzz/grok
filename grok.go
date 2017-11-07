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
	defaults     = Conf{
		depth:     0,
		writer:    os.Stdout,
		tabstop:   4,
		colour:    true,
		maxDepth:  10,
		maxLength: 100,
	}
)

type Writer func(string, ...interface{})

type Indenter func(string, int) string

type Colourizer func(str string, colour string) string

type Option func(c *Conf)

type Conf struct {
	depth     int
	writer    io.Writer
	tabstop   int
	colour    bool
	maxDepth  int
	maxLength int
}

func writer(w io.Writer) Writer {
	return Writer(func(format string, params ...interface{}) {
		w.Write([]byte(fmt.Sprintf(format, params...)))
	})
}

func indenter(tabstop int) Indenter {
	return Indenter(func(v string, depth int) string {
		return strings.Repeat(" ", depth*tabstop) + v
	})
}

func colourizer(colourize bool) Colourizer {
	return Colourizer(func(str string, colour string) string {
		if colourize {
			return colour + str + colourReset
		}
		return str
	})
}

// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option {
	return func(c *Conf) {
		c.writer = w
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

// WithMaxLength sets the maximum length of string byValue. Default is `100`, use `0` for unlimited
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
