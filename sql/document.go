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
