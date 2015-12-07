package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"time"
)

type Document struct {
	UUID uuid.UUID
	Tags map[string]string
}

// Generates a batch INSERT statement
func (doc *Document) GenerateinsertStatement() string {
	var s = "INSERT INTO data (uuid, dkey, dval) VALUES "
	for key, val := range doc.Tags {
		if len(val) == 0 {
			val = "NULL"
		} else {
			val = `"` + val + `"`
		}
		s += fmt.Sprintf(`("%s", "%s", %s),`, doc.UUID.String(), key, val)
	}
	s = s[:len(s)-1]
	return s + ";"
}

func (doc *Document) GenerateinsertStatementWithTimestamp(timestamp time.Time) string {
	var s = "INSERT INTO data (uuid, dkey, dval, timestamp) VALUES "
	for key, val := range doc.Tags {
		if len(val) == 0 {
			val = "NULL"
		} else {
			val = `"` + val + `"`
		}
		s += fmt.Sprintf(`("%s", "%s", %s, "%s"),`, doc.UUID.String(), key, val, timestamp.Format(time.RFC3339))
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
		if len(val) == 0 {
			val = "NULL"
		} else {
			val = `"` + val + `"`
		}
		ret = append(ret, fmt.Sprintf(`("%s", "%s", %s)`, doc.UUID.String(), key, val))
	}
	return ret
}

func (doc *Document) PrettyString() string {
	if b, err := json.MarshalIndent(doc, "", "  "); err != nil {
		return fmt.Sprintf("ERROR FORMATTING (%v) %v", err, doc)
	} else {
		return string(b)
	}
}

// Generate a list of documents from the results of a SQL query
func DocsFromRows(rows *sql.Rows) ([]*Document, error) {
	var (
		uniqueDocs = map[string]*Document{}
		docs       = []*Document{}
	)
	if rows == nil {
		return docs, fmt.Errorf("No rows returned")
	}
	for rows.Next() {
		var (
			duuid string
			dkey  string
			dval  string
		)
		if err := rows.Scan(&duuid, &dkey, &dval); err != nil {
			return docs, err
		}
		if doc, found := uniqueDocs[duuid]; found {
			doc.Tags[dkey] = dval
		} else {
			parsedUUID, err := uuid.FromString(duuid)
			if err != nil {
				return docs, err
			}
			doc = &Document{UUID: parsedUUID, Tags: map[string]string{dkey: dval}}
			uniqueDocs[duuid] = doc
			docs = append(docs, doc)
		}
	}
	return docs, nil
}
