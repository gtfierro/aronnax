package main

import (
	"fmt"
	"github.com/satori/go.uuid"
)

type Document struct {
	UUID uuid.UUID
	Tags map[string]string
}

// Generates a batch INSERT statement
func (doc *Document) GenerateinsertStatement() string {
	var s = "INSERT INTO data (uuid, dkey, dval) VALUES "
	for key, val := range doc.Tags {
		s += fmt.Sprintf(`("%s", "%s", "%s"),`, doc.UUID.String(), key, val)
	}
	s = s[:len(s)-1]
	return s + ";"
}

// Generate the VALUES input ("uuid","key","val") for the purposes
// of checking w/ unit tests. This is because we don't get guaranteed
// iteration order w/ maps (our tags field)
func (doc *Document) GenerateValues() []string {
	var ret []string
	for key, val := range doc.Tags {
		ret = append(ret, fmt.Sprintf(`("%s", "%s", "%s")`, doc.UUID.String(), key, val))
	}
	return ret
}
