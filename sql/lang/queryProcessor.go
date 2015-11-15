package query

type query struct {
	selects []selectTerm
	wheres  []whereTerm
}

type selectTerm struct {
}

type whereTerm struct {
}
