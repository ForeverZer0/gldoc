package ref

import (
	"fmt"
	"strings"
)

// A Function describes the prototype and parameter names/types of an OpenGL function.
type Function struct {
	Name string   `xml:"funcdef>function" json:"name"`
	Args []string `xml:"paramdef>parameter" json:"args"`
}

// String implements the [fmt.Stringer] interface.
func (fn *Function) String() string {
	return fmt.Sprintf("Name: %s Args: [%s]", fn.Name, strings.Join(fn.Args, ", "))
}
