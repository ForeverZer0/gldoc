package ref

import (
	"fmt"
	"strings"
)

// A Function describes the prototype and parameter names/types of an OpenGL function.
type Function struct {
	// The OpenGL name of the function.
	Name string `xml:"funcdef>function" json:"name"`
	// The ordered list of names for the function arguments.
	Args []string `xml:"paramdef>parameter" json:"args"`
}

// An Arg describes the name and description of a function argument.
type Arg struct {
	Name string // The name of the argument.
	Desc string // A brief description of the argument.
}

// String implements the [fmt.Stringer] interface.
func (fn *Function) String() string {
	return fmt.Sprintf("Name: %s Args: [%s]", fn.Name, strings.Join(fn.Args, ", "))
}
