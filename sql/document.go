package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"time"
)

type Document struct {
	// the unique document identifier
	UUID uuid.UUID
	// Key->Value pairs this document contains
	Tags map[string]string
	// When the keys were applied
	TagTimes map[string]time.Time
	// the time at which this document is valid (the max of the tag times)
	ValidTime time.Time
}

// Generates a batch INSERT statement. If ignoreTagTimes is true, then the tags are
// applied at time "now" as determined by the MySQL database
func (doc *Document) GenerateInsertStatement(ignoreTagTimes bool) string {
	var s = "INSERT INTO data (uuid, dkey, dval) VALUES "
	for key, val := range doc.Tags {
		if len(val) == 0 {
			val = "NULL"
		} else {
			val = `"` + val + `"`
		}
		if ignoreTagTimes {
			s += fmt.Sprintf(`("%s", "%s", %s),`, doc.UUID.String(), key, val)
		} else {
			s += fmt.Sprintf(`("%s", "%s", %s, "%s"),`, doc.UUID.String(), key, val, doc.TagTimes[key].Format(time.RFC3339))
		}
	}
	s = s[:len(s)-1]
	return s + ";"
}

func (doc *Document) GenerateInsertStatementWithTimestamp(timestamp time.Time) string {
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

// finds the most recent tag.Time from its list of tags
// and sets that to doc.ValidTime
func (doc *Document) CalcMaxTagTime() {
	latest := time.Time{} // earliest time
	for _, tagtime := range doc.TagTimes {
		if tagtime.After(latest) {
			latest = tagtime
		}
	}
	doc.ValidTime = latest
}

func (doc *Document) PrettyString() string {
	if b, err := json.MarshalIndent(doc, "", "  "); err != nil {
		return fmt.Sprintf("ERROR FORMATTING (%v) %v", err, doc)
	} else {
		return string(b)
	}
}

// Generate a list of documents from the results of a SQL query
func DocsFromRows(rows *sql.Rows, now time.Time) ([]*Document, error) {
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
			dval  sql.NullString
			dtime time.Time
		)
		if err := rows.Scan(&duuid, &dkey, &dval, &dtime); err != nil {
			return docs, err
		}
		if doc, found := uniqueDocs[duuid]; found {
			if dval.Valid { // value can be null
				doc.Tags[dkey] = dval.String
			} // but we still want to keep track of the time
			doc.TagTimes[dkey] = dtime
		} else {
			parsedUUID, err := uuid.FromString(duuid)
			if err != nil {
				return docs, err
			}
			doc = &Document{UUID: parsedUUID, Tags: map[string]string{}, TagTimes: map[string]time.Time{dkey: dtime}}
			if dval.Valid {
				doc.Tags[dkey] = dval.String
			}
			uniqueDocs[duuid] = doc
			docs = append(docs, doc)
		}
	}
	// add in the valid times for all documents
	for _, doc := range docs {
		if now == ZERO_TIME {
			doc.CalcMaxTagTime()
		} else {
			doc.ValidTime = now
		}
	}
	return docs, nil
}
