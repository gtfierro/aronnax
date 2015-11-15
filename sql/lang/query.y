%{
package query

import (
	"github.com/taylorchu/toki"
	"bufio"
	"fmt"
)
%}

%union{
	str string
    selectTerm  selectTerm
    selectTermList  []selectTerm
    whereTerm  whereTerm
    whereTermList  []whereTerm
}

%token <str> SELECT DISTINCT WHERE
%token <str> LVALUE QSTRING LIKE HAS
%token <str> NOW SET BEFORE AFTER AND AS TO OR IN NOT
%token <str> LPAREN RPAREN NEWLINE
%token NUMBER
%token SEMICOLON

%token <str> EQ NEQ COMMA ALL

%type <selectTermList> selectTermList selectClause
%type <selectTerm> selectTerm
%type <whereTermList> whereClause
%type <whereTerm> whereTerm

%right EQ

%%

query   :   selectClause    whereClause SEMICOLON
        {
            //queryLex.(*queryLex).query.select
            fmt.Printf("Select: %v\n", $1)
            fmt.Printf("Where: %v\n", $2)
        }
        ;

selectClause    :   SELECT  selectTermList
                {
                    $$ = $2
                }
                ;

selectTermList  :   selectTerm
                {
                    $$ = []selectTerm{$1}
                }
                |   selectTerm  COMMA selectTermList
                {
                    $$ = append([]selectTerm{$1}, $3...)
                }
                ;

selectTerm  :   LVALUE
            {
                $$ = selectTerm{}
            }
            ;


whereClause :   WHERE   whereTerm
            {
                $$ = []whereTerm{}
            }
            ;

whereTerm   : LVALUE LIKE QSTRING
            {
                $$ = whereTerm{}
            }
            | LVALUE EQ QSTRING
            {
                $$ = whereTerm{}
            }
            | LVALUE NEQ QSTRING
            {
                $$ = whereTerm{}
            }
            | HAS LVALUE
            {
                $$ = whereTerm{}
            }
            ;

%%

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
    query   *query
    querystring   string
	scanner *toki.Scanner
    lasttoken string
    tokens  []string
    error   error
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
