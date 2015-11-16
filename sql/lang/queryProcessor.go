//go:generate go tool yacc -o query.go -p Query query.y
package query

import (
	"fmt"
)

type Query struct {
	Selects []SelectTerm
	Wheres  WhereClause
	SQL     string
}

type SelectTerm struct {
	Tag string
}

type WhereTerm struct {
	Key string
	Op  string
	Val string
	SQL string
}

type WhereClause struct {
	SQL string
}

func (wt WhereTerm) ToSQL() string {
	var s string
	switch wt.Op {
	case "=":
		s = fmt.Sprintf(`data.dkey = "%s" and data.dval = %s`, wt.Key, wt.Val)
	case "!=":
		s = fmt.Sprintf(`data.dkey = "%s" and data.dval != %s`, wt.Key, wt.Val)
	case "has":
		s = fmt.Sprintf(`data.dkey = "%s"`, wt.Key)
	case "like":
		s = fmt.Sprintf(`data.dkey = "%s" and data.dval LIKE %s`, wt.Key, wt.Val)
	}
	return s
}
