//go:generate go tool yacc -o query.go -p Query query.y
package query

type Query struct {
	Selects []SelectTerm
	Wheres  []WhereTerm
}

type SelectTerm struct {
	Tag string
}

type WhereTerm struct {
	Key string
	Op  string
	Val string
}
