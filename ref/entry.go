package ref

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// A Entry describes a simplified reference page of the OpenGL spec, intended for providing brief inline documentation.
type Entry struct {
	// The shorthand name of the function(s), without any suffix, etc.
	//
	// For example, the following functions would all be defined within the "glUniform" entry:
	//
	//  - glUniform2d
	//  - glUniform3fv
	//  - glUniformMatrix4fv
	Name string `json:"name"`
	// Desc is the description/summary for all functions covered within this entry. This is typically a brief single
	// sentence, often without a subject and/or ending punctuation.
	//
	// For example, the entry for "glGet" reads verbatim as: "return the value or values of a selected parameter"
	Desc string `json:"desc"`
	// Description of each OpenGL function this entry pertains to. These functions describe the full names of actual
	// functions as they are defined in the OpenGL specification.
	Funcs []Function `json:"functions"`
	// Documentation for function parameters, using name as key and summary as value.
	//
	// Unlike the Desc field, these are typically full sentences. Any inner XML or links are flattened to only provide
	// the text value. Newlines and tabs used for formatting on a webpage are stripped.
	Params map[string]string `json:"params"`
	// Contains the names of other related functions.
	SeeAlso []string `json:"seealso"`
	// Possible error value constants OpenGL will emit when using this function.
	Errors []string `json:"errors"`
}

// NewEntry returns a new Entry from the given reader.
func NewEntry(r io.Reader) (*Entry, error) {
	var entry Entry
	d := xml.NewDecoder(r)
	d.Strict = false
	if err := d.Decode(&entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

// LoadEntry returns a new Entry loaded from the XML file at the given path.
func LoadEntry(path string) (*Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	entry, err := NewEntry(file)
	if err == nil && len(entry.Name) == 0 {
		base := filepath.Base(path)
		entry.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return entry, err
}

// Func returns the function definition with the given name and flag indicating if it was found.
func (entry *Entry) Func(name string) (fn Function, ok bool) {
	for _, fn = range entry.Funcs {
		if fn.Name == name {
			ok = true
			return
		}
	}
	return
}

// UnmarshalXML implements the [xml.Unmarshaler] interface.
func (entry *Entry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	entry.Params = make(map[string]string)

	token, err := d.Token()
	for token != start.End() {
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		for _, attr := range start.Attr {
			if attr.Name.Local == "id" {
				entry.Name = attr.Value
			}
		}

		if child, ok := token.(xml.StartElement); ok {
			switch child.Name.Local {
			case "refnamediv":
				err = entry.parseDesc(d, child)
			case "refsynopsisdiv":
				err = entry.parseFuncs(d, child)
			case "refsect1":
				for _, attr := range child.Attr {
					if attr.Name.Local != "id" {
						continue
					}
					switch attr.Value {
					case "parameters", "parameters2", "parameters3":
						err = entry.parseParams(d, child)
					case "seealso":
						err = entry.parseSeeAlso(d, child)
					case "errors":
						err = entry.parseErrors(d, child)
						// case "description", "description2":
						// case "notes":
						// case "example", "examples":
						// case "associatedgets":
						// case "versions":
						// case "Copyright":
					}
					break
				}
			}
			if err != nil {
				return err
			}
		}

		token, err = d.Token()
	}

	return nil
}

func Sanitize[T ~string](text T) string {
	var sb strings.Builder
	for _, line := range strings.Split(string(text), "\n") {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strings.TrimSpace(line))
	}

	return sb.String()
}

func (entry *Entry) parseDesc(d *xml.Decoder, start xml.StartElement) error {
	var name struct {
		Name string `xml:"refname"`
		Desc string `xml:"refpurpose"`
	}
	if err := d.DecodeElement(&name, &start); err != nil {
		return err
	}
	if len(entry.Name) == 0 {
		entry.Name = name.Name
	}
	entry.Desc = Sanitize(name.Desc)
	return nil
}

func (entry *Entry) parseFuncs(d *xml.Decoder, start xml.StartElement) error {
	var funcs struct {
		Values []Function `xml:"funcsynopsis>funcprototype"`
	}
	if err := d.DecodeElement(&funcs, &start); err != nil {
		return err
	}
	entry.Funcs = append(entry.Funcs, funcs.Values...)
	return nil
}

func (entry *Entry) parseParams(d *xml.Decoder, start xml.StartElement) error {
	var values struct {
		Items []struct {
			Names []string `xml:"term>parameter"`
			Text  Text     `xml:"listitem"`
		} `xml:"variablelist>varlistentry"`
	}

	if err := d.DecodeElement(&values, &start); err != nil {
		return err
	}

	for _, item := range values.Items {
		for _, name := range item.Names {
			entry.Params[name] = strings.TrimSpace(string(item.Text))
		}
	}
	return nil
}

func (entry *Entry) parseSeeAlso(d *xml.Decoder, start xml.StartElement) error {
	var values struct {
		Names []string `xml:"para>citerefentry>refentrytitle"`
	}
	if err := d.DecodeElement(&values, &start); err != nil {
		return err
	}
	entry.SeeAlso = append(entry.SeeAlso, values.Names...)
	return nil
}

func (entry *Entry) parseErrors(d *xml.Decoder, start xml.StartElement) error {
	var values struct {
		Names []string `xml:"para>constant"`
	}
	if err := d.DecodeElement(&values, &start); err != nil {
		return err
	}

	for _, errName := range values.Names {
		if _, ok := knownErrs[errName]; ok && !slices.Contains(entry.Errors, errName) {
			entry.Errors = append(entry.Errors, errName)
		}
	}
	return nil
}

// A map where each key is the name of a known OpenGL error.
var knownErrs = map[string]bool{
	"GL_OUT_OF_MEMORY":                 true,
	"GL_INVALID_ENUM":                  true,
	"GL_INVALID_VALUE":                 true,
	"GL_INVALID_OPERATION":             true,
	"GL_STACK_OVERFLOW":                true,
	"GL_STACK_UNDERFLOW":               true,
	"GL_INVALID_FRAMEBUFFER_OPERATION": true,
	"GL_CONTEXT_LOST":                  true,
	"GL_TABLE_TOO_LARGE":               true,
}
