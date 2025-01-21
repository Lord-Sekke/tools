package debug

import (
	"runtime"
	"time"
)

type Context struct {  // Changed 'context' to 'Context'
	os      string
	started time.Time
	debug   bool
	email   string
}

func NewContext(debug bool) (*Context, error) {  // Update function to return *Context
	return &Context{
		started: time.Now(),
		os:      runtime.GOOS,
		debug:   debug,
	}, nil
}

func (d *Context) Action(text string) {
	if d.debug {
		// Implement debug action logic here
	}
}

func (d *Context) Error(text string, err error) {
	if d.debug {
		// Implement error handling logic here
	}
}
