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
	Key         string
	Op          string
	Val         string
	SQL         string
	Letter      string
	IsPredicate bool
}

func (wt WhereTerm) GetClause() WhereClause {
	if wt.IsPredicate {
		return WrapTermInSelect(wt.SQL, wt.Letter)
	} else {
		return WhereClause{SQL: wt.SQL, Letter: wt.Letter}
	}
}

func (wt WhereTerm) GetClauseWithTime(inner string) WhereClause {
	if wt.IsPredicate {
		return WrapTermInSelectWithTime(wt.SQL, wt.Letter, inner)
	} else {
		return WhereClause{SQL: wt.SQL, Letter: wt.Letter}
	}
}

type WhereClause struct {
	SQL    string
	Letter string
}

func WrapTermInSelect(where, letter string) WhereClause {
	sql := fmt.Sprintf(`
    (
    select distinct data.uuid
    from data
    inner join
    (
        select distinct uuid, dkey, max(timestamp) as maxtime from data
        group by dkey, uuid order by timestamp desc
    ) sorted
    on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
    where data.dval is not null and
    %s)`, where)
	return WhereClause{SQL: sql, Letter: letter}
}

func WrapTermInSelectWithTime(where, letter, inner string) WhereClause {
	sql := fmt.Sprintf(`
    (
    select distinct data.uuid
    from data
    inner join
    (
        %s
    ) sorted
    on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
    where data.dval is not null and
    %s)`, inner, where)
	return WhereClause{SQL: sql, Letter: letter}
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
