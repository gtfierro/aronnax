//line query.y:2
package query

import __yyfmt__ "fmt"

//line query.y:2
import (
	"bufio"
	"fmt"
	"github.com/taylorchu/toki"
)

//line query.y:11
type querySymType struct {
	yys            int
	str            string
	selectTerm     selectTerm
	selectTermList []selectTerm
	whereTerm      whereTerm
	whereTermList  []whereTerm
}

const SELECT = 57346
const DISTINCT = 57347
const WHERE = 57348
const LVALUE = 57349
const QSTRING = 57350
const LIKE = 57351
const HAS = 57352
const NOW = 57353
const SET = 57354
const BEFORE = 57355
const AFTER = 57356
const AND = 57357
const AS = 57358
const TO = 57359
const OR = 57360
const IN = 57361
const NOT = 57362
const LPAREN = 57363
const RPAREN = 57364
const NEWLINE = 57365
const NUMBER = 57366
const SEMICOLON = 57367
const EQ = 57368
const NEQ = 57369
const COMMA = 57370
const ALL = 57371

var queryToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"SELECT",
	"DISTINCT",
	"WHERE",
	"LVALUE",
	"QSTRING",
	"LIKE",
	"HAS",
	"NOW",
	"SET",
	"BEFORE",
	"AFTER",
	"AND",
	"AS",
	"TO",
	"OR",
	"IN",
	"NOT",
	"LPAREN",
	"RPAREN",
	"NEWLINE",
	"NUMBER",
	"SEMICOLON",
	"EQ",
	"NEQ",
	"COMMA",
	"ALL",
}
var queryStatenames = [...]string{}

const queryEofCode = 1
const queryErrCode = 2
const queryMaxDepth = 200

//line query.y:92
const eof = 0

var supported_formats = []string{"1/2/2006",
	"1-2-2006",
	"1/2/2006 03:04:05 PM MST",
	"1-2-2006 03:04:05 PM MST",
	"1/2/2006 15:04:05 MST",
	"1-2-2006 15:04:05 MST",
	"2006-1-2 15:04:05 MST"}

type List []string

type queryLex struct {
	query       *query
	querystring string
	scanner     *toki.Scanner
	lasttoken   string
	tokens      []string
	error       error
}

func newQueryLexer(s string) *queryLex {
	scanner := toki.NewScanner(
		[]toki.Def{
			{Token: WHERE, Pattern: "where"},
			{Token: SELECT, Pattern: "select"},
			{Token: DISTINCT, Pattern: "distinct"},
			{Token: ALL, Pattern: "\\*"},
			{Token: NOW, Pattern: "now"},
			{Token: SET, Pattern: "set"},
			{Token: BEFORE, Pattern: "before"},
			{Token: AFTER, Pattern: "after"},
			{Token: COMMA, Pattern: ","},
			{Token: AND, Pattern: "and"},
			{Token: AS, Pattern: "as"},
			{Token: TO, Pattern: "to"},
			{Token: OR, Pattern: "or"},
			{Token: IN, Pattern: "in"},
			{Token: HAS, Pattern: "has"},
			{Token: NOT, Pattern: "not"},
			{Token: NEQ, Pattern: "!="},
			{Token: EQ, Pattern: "="},
			{Token: LPAREN, Pattern: "\\("},
			{Token: RPAREN, Pattern: "\\)"},
			{Token: SEMICOLON, Pattern: ";"},
			{Token: NEWLINE, Pattern: "\n"},
			{Token: LIKE, Pattern: "(like)|~"},
			{Token: NUMBER, Pattern: "([+-]?([0-9]*\\.)?[0-9]+)"},
			{Token: LVALUE, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%_\\-]*"},
			{Token: QSTRING, Pattern: "(\"[^\"\\\\]*?(\\.[^\"\\\\]*?)*?\")|('[^'\\\\]*?(\\.[^'\\\\]*?)*?')"},
		})
	scanner.SetInput(s)
	return &queryLex{querystring: s, scanner: scanner, error: nil, lasttoken: "", tokens: []string{}}
}

func (lex *queryLex) Lex(lval *querySymType) int {
	r := lex.scanner.Next()
	lex.lasttoken = r.String()
	if r.Pos.Line == 2 || len(r.Value) == 0 {
		return eof
	}
	lval.str = string(r.Value)
	lex.tokens = append(lex.tokens, lval.str)
	return int(r.Token)
}

func (lex *queryLex) Error(s string) {
	lex.error = fmt.Errorf(s)
}

func readline(fi *bufio.Reader) (string, bool) {
	fmt.Printf("smap> ")
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}

//go:generate go tool yacc -o query.go -p query query.y

//line yacctab:1
var queryExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const queryNprod = 11
const queryPrivate = 57344

var queryTokenNames []string
var queryStates []string

const queryLast = 21

var queryAct = [...]int{

	14, 13, 9, 11, 8, 6, 12, 21, 20, 19,
	17, 5, 3, 1, 10, 4, 7, 15, 16, 18,
	2,
}
var queryPact = [...]int{

	8, -1000, 5, -3, -23, -4, -1000, -27, -1000, -1000,
	-1000, -9, 3, -3, 1, 0, -1, -1000, -1000, -1000,
	-1000, -1000,
}
var queryPgo = [...]int{

	0, 5, 20, 16, 15, 14, 13,
}
var queryR1 = [...]int{

	0, 6, 2, 1, 1, 3, 4, 5, 5, 5,
	5,
}
var queryR2 = [...]int{

	0, 3, 2, 1, 3, 1, 2, 3, 3, 3,
	2,
}
var queryChk = [...]int{

	-1000, -6, -2, 4, -4, 6, -1, -3, 7, 25,
	-5, 7, 10, 28, 9, 26, 27, 7, -1, 8,
	8, 8,
}
var queryDef = [...]int{

	0, -2, 0, 0, 0, 0, 2, 3, 5, 1,
	6, 0, 0, 0, 0, 0, 0, 10, 4, 7,
	8, 9,
}
var queryTok1 = [...]int{

	1,
}
var queryTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29,
}
var queryTok3 = [...]int{
	0,
}

var queryErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	queryDebug        = 0
	queryErrorVerbose = false
)

type queryLexer interface {
	Lex(lval *querySymType) int
	Error(s string)
}

type queryParser interface {
	Parse(queryLexer) int
	Lookahead() int
}

type queryParserImpl struct {
	lookahead func() int
}

func (p *queryParserImpl) Lookahead() int {
	return p.lookahead()
}

func queryNewParser() queryParser {
	p := &queryParserImpl{
		lookahead: func() int { return -1 },
	}
	return p
}

const queryFlag = -1000

func queryTokname(c int) string {
	if c >= 1 && c-1 < len(queryToknames) {
		if queryToknames[c-1] != "" {
			return queryToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func queryStatname(s int) string {
	if s >= 0 && s < len(queryStatenames) {
		if queryStatenames[s] != "" {
			return queryStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func queryErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !queryErrorVerbose {
		return "syntax error"
	}

	for _, e := range queryErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + queryTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := queryPact[state]
	for tok := TOKSTART; tok-1 < len(queryToknames); tok++ {
		if n := base + tok; n >= 0 && n < queryLast && queryChk[queryAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if queryDef[state] == -2 {
		i := 0
		for queryExca[i] != -1 || queryExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; queryExca[i] >= 0; i += 2 {
			tok := queryExca[i]
			if tok < TOKSTART || queryExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if queryExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += queryTokname(tok)
	}
	return res
}

func querylex1(lex queryLexer, lval *querySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = queryTok1[0]
		goto out
	}
	if char < len(queryTok1) {
		token = queryTok1[char]
		goto out
	}
	if char >= queryPrivate {
		if char < queryPrivate+len(queryTok2) {
			token = queryTok2[char-queryPrivate]
			goto out
		}
	}
	for i := 0; i < len(queryTok3); i += 2 {
		token = queryTok3[i+0]
		if token == char {
			token = queryTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = queryTok2[1] /* unknown char */
	}
	if queryDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", queryTokname(token), uint(char))
	}
	return char, token
}

func queryParse(querylex queryLexer) int {
	return queryNewParser().Parse(querylex)
}

func (queryrcvr *queryParserImpl) Parse(querylex queryLexer) int {
	var queryn int
	var querylval querySymType
	var queryVAL querySymType
	var queryDollar []querySymType
	_ = queryDollar // silence set and not used
	queryS := make([]querySymType, queryMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	querystate := 0
	querychar := -1
	querytoken := -1 // querychar translated into internal numbering
	queryrcvr.lookahead = func() int { return querychar }
	defer func() {
		// Make sure we report no lookahead when not parsing.
		querystate = -1
		querychar = -1
		querytoken = -1
	}()
	queryp := -1
	goto querystack

ret0:
	return 0

ret1:
	return 1

querystack:
	/* put a state and value onto the stack */
	if queryDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", queryTokname(querytoken), queryStatname(querystate))
	}

	queryp++
	if queryp >= len(queryS) {
		nyys := make([]querySymType, len(queryS)*2)
		copy(nyys, queryS)
		queryS = nyys
	}
	queryS[queryp] = queryVAL
	queryS[queryp].yys = querystate

querynewstate:
	queryn = queryPact[querystate]
	if queryn <= queryFlag {
		goto querydefault /* simple state */
	}
	if querychar < 0 {
		querychar, querytoken = querylex1(querylex, &querylval)
	}
	queryn += querytoken
	if queryn < 0 || queryn >= queryLast {
		goto querydefault
	}
	queryn = queryAct[queryn]
	if queryChk[queryn] == querytoken { /* valid shift */
		querychar = -1
		querytoken = -1
		queryVAL = querylval
		querystate = queryn
		if Errflag > 0 {
			Errflag--
		}
		goto querystack
	}

querydefault:
	/* default state action */
	queryn = queryDef[querystate]
	if queryn == -2 {
		if querychar < 0 {
			querychar, querytoken = querylex1(querylex, &querylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if queryExca[xi+0] == -1 && queryExca[xi+1] == querystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			queryn = queryExca[xi+0]
			if queryn < 0 || queryn == querytoken {
				break
			}
		}
		queryn = queryExca[xi+1]
		if queryn < 0 {
			goto ret0
		}
	}
	if queryn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			querylex.Error(queryErrorMessage(querystate, querytoken))
			Nerrs++
			if queryDebug >= 1 {
				__yyfmt__.Printf("%s", queryStatname(querystate))
				__yyfmt__.Printf(" saw %s\n", queryTokname(querytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for queryp >= 0 {
				queryn = queryPact[queryS[queryp].yys] + queryErrCode
				if queryn >= 0 && queryn < queryLast {
					querystate = queryAct[queryn] /* simulate a shift of "error" */
					if queryChk[querystate] == queryErrCode {
						goto querystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if queryDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", queryS[queryp].yys)
				}
				queryp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if queryDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", queryTokname(querytoken))
			}
			if querytoken == queryEofCode {
				goto ret1
			}
			querychar = -1
			querytoken = -1
			goto querynewstate /* try again in the same state */
		}
	}

	/* reduction by production queryn */
	if queryDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", queryn, queryStatname(querystate))
	}

	querynt := queryn
	querypt := queryp
	_ = querypt // guard against "declared and not used"

	queryp -= queryR2[queryn]
	// queryp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if queryp+1 >= len(queryS) {
		nyys := make([]querySymType, len(queryS)*2)
		copy(nyys, queryS)
		queryS = nyys
	}
	queryVAL = queryS[queryp+1]

	/* consult goto table to find next state */
	queryn = queryR1[queryn]
	queryg := queryPgo[queryn]
	queryj := queryg + queryS[queryp].yys + 1

	if queryj >= queryLast {
		querystate = queryAct[queryg]
	} else {
		querystate = queryAct[queryj]
		if queryChk[querystate] != -queryn {
			querystate = queryAct[queryg]
		}
	}
	// dummy call; replaced with literal code
	switch querynt {

	case 1:
		queryDollar = queryS[querypt-3 : querypt+1]
		//line query.y:38
		{
			//queryLex.(*queryLex).query.select
			fmt.Printf("Select: %v\n", queryDollar[1].selectTermList)
			fmt.Printf("Where: %v\n", queryDollar[2].whereTermList)
		}
	case 2:
		queryDollar = queryS[querypt-2 : querypt+1]
		//line query.y:46
		{
			queryVAL.selectTermList = queryDollar[2].selectTermList
		}
	case 3:
		queryDollar = queryS[querypt-1 : querypt+1]
		//line query.y:52
		{
			queryVAL.selectTermList = []selectTerm{queryDollar[1].selectTerm}
		}
	case 4:
		queryDollar = queryS[querypt-3 : querypt+1]
		//line query.y:56
		{
			queryVAL.selectTermList = append([]selectTerm{queryDollar[1].selectTerm}, queryDollar[3].selectTermList...)
		}
	case 5:
		queryDollar = queryS[querypt-1 : querypt+1]
		//line query.y:62
		{
			queryVAL.selectTerm = selectTerm{}
		}
	case 6:
		queryDollar = queryS[querypt-2 : querypt+1]
		//line query.y:69
		{
			queryVAL.whereTermList = []whereTerm{}
		}
	case 7:
		queryDollar = queryS[querypt-3 : querypt+1]
		//line query.y:75
		{
			queryVAL.whereTerm = whereTerm{}
		}
	case 8:
		queryDollar = queryS[querypt-3 : querypt+1]
		//line query.y:79
		{
			queryVAL.whereTerm = whereTerm{}
		}
	case 9:
		queryDollar = queryS[querypt-3 : querypt+1]
		//line query.y:83
		{
			queryVAL.whereTerm = whereTerm{}
		}
	case 10:
		queryDollar = queryS[querypt-2 : querypt+1]
		//line query.y:87
		{
			queryVAL.whereTerm = whereTerm{}
		}
	}
	goto querystack /* stack new state and value */
}
