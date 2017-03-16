# Dump it like you mean it!

Tired of debugging using the same old `fmt.Println` and `fmt.Printf("%+v", var)`. Enter `dump`! A new package to make your vars look pretty again.

## Usage:

```go
import "github.com/leisurespecials/lighthouse/lib/debug/dump"

var := "Whatever you want me to be!"

dump.Value(var)

// or for customised output

dump.Value(var, ...dump.Option)
```

The dump package comes with the following customisation options baked in:
 
```go
// WithWriter redirects output from debug functions to the given io.Writer
func WithWriter(w io.Writer) Option
```
```go
// WithoutColours disables colouring of output from debug functions. Defaults to `true`
func WithoutColours() Option
```
```go
// WithMaxDepth sets the maximum recursion depth from debug functions. Defaults to `10`, use `0` for unlimited
func WithMaxDepth(depth int) Option 
```
```go
// WithMaxLength sets the maximum length of string values. Default is `100`, use `0` for unlimited
func WithMaxLength(chars int) Option
```
```go
// WithTabStop sets the width of a tabstop to the given char count. Defaults to `4`
func WithTabStop(chars int) Option
```

## Got 99 problems and this code is one?

Talk to `Merten van Gerven`. He wrote this package so he can figure it out, right?!