package ref

import "path/filepath"

// A Spec defines the reference documentation sources for an OpenGL API and version.
type Spec struct {
	Name    string            // The name of the specification (i.e. "gl4", "es3.0", etc).
	Entries map[string]*Entry // Documentation entries defined in the specification.
}

// LoadSpec returns a new Spec loaded from the specified base and child directory name.
func LoadSpec(base, name string) (spec Spec, err error) {
	var glob []string
	var page *Entry

	glob, err = filepath.Glob(filepath.Join(base, name, "gl*.xml"))
	if err != nil {
		return
	}

	spec.Name = name
	spec.Entries = make(map[string]*Entry)
	for _, match := range glob {
		page, err = LoadEntry(match)
		if err != nil {
			return
		}
		spec.Entries[page.Name] = page
		for _, fn := range page.Funcs {
			spec.Entries[fn.Name] = page
		}
	}

	return
}
