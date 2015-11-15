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
    selectTerm  SelectTerm
    selectTermList  []SelectTerm
    whereTerm  WhereTerm
    whereTermList  []WhereTerm
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
%type <whereTermList> whereTermList whereClause
%type <whereTerm> whereTerm

%right EQ

%%

query   :   selectClause    whereClause SEMICOLON
        {
            //QueryLex.(*QueryLex).query.select
            Querylex.(*QueryLex).Query.Selects = $1
            Querylex.(*QueryLex).Query.Wheres = $2
            fmt.Printf("Select: %v\n", $1)
            fmt.Printf("Where: %v\n", $2)
        }
        ;

selectClause    :   SELECT  selectTermList
                {
                    fmt.Println("select")
                    $$ = $2
                }
                ;

selectTermList  :   selectTerm
                {
                    $$ = []SelectTerm{$1}
                }
                |   selectTerm  COMMA selectTermList
                {
                    $$ = append([]SelectTerm{$1}, $3...)
                }
                ;

selectTerm  :   LVALUE
            {
                $$ = SelectTerm{Tag: $1}
            }
            ;


whereClause :   WHERE   whereTermList
            {
                $$ = $2
            }
            ;

whereTermList : whereTerm
                {
                    $$ = []WhereTerm{$1}
                }
              ;

//whereTermList : whereTermList AND whereTermList
//              | whereTermList OR whereTermList
//              | NOT whereTermList
//              | LPAREN whereTermList RPAREN
//              | whereTerm
//              ;

whereTerm   : LVALUE LIKE QSTRING
            {
                $$ = WhereTerm{Key: $1, Op: $2, Val: $3}
            }
            | LVALUE EQ QSTRING
            {
                $$ = WhereTerm{Key: $1, Op: $2, Val: $3}
            }
            | LVALUE NEQ QSTRING
            {
                $$ = WhereTerm{Key: $1, Op: $2, Val: $3}
            }
            | HAS LVALUE
            {
                $$ = WhereTerm{Key: $2, Op: $1}
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
type QueryLex struct {
    Query   *Query
    querystring   string
	scanner *toki.Scanner
    lasttoken string
    tokens  []string
    Err   error
}

func NewQueryLexer(s string) *QueryLex {
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
	return &QueryLex{Query: &Query{}, querystring: s, scanner: scanner, Err: nil, lasttoken: "", tokens: []string{}}
}

func (lex *QueryLex) Lex(lval *QuerySymType) int {
	r := lex.scanner.Next()
    lex.lasttoken = r.String()
	if r.Pos.Line == 2 || len(r.Value) == 0 {
		return eof
	}
	lval.str = string(r.Value)
    lex.tokens = append(lex.tokens, lval.str)
	return int(r.Token)
}

func (lex *QueryLex) Error(s string) {
    lex.Err = fmt.Errorf(s)
}

func readline(fi *bufio.Reader) (string, bool) {
	fmt.Printf("aronnax> ")
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}
