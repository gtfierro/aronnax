%{
package query

import (
    "github.com/taylorchu/toki"
    "bufio"
    "fmt"
    "strconv"
    _time "time"
)
%}

%union{
    str string
    selectTerm  SelectTerm
    selectTermList  []SelectTerm
    whereTerm  WhereTerm
    whereClause  WhereClause
    time _time.Time
    timediff _time.Duration
}

%token <str> SELECT DISTINCT WHERE
%token <str> LVALUE QSTRING LIKE HAS
%token <str> NOW SET IBEFORE BEFORE IAFTER AFTER AND AS TO OR IN NOT FOR
%token <str> LPAREN RPAREN NEWLINE
%token NUMBER
%token SEMICOLON

%token <str> EQ NEQ COMMA ALL

%type <selectTermList> selectTermList selectClause
%type <selectTerm> selectTerm
%type <whereClause> whereClause
%type <whereTerm> whereTerm
%type <time> timeref abstime
%type <timediff> reltime
%type <str> NUMBER

%right EQ

%%

query   :   SELECT selectClause WHERE whereClause SEMICOLON
        {
            Querylex.(*QueryLex).Query.Selects = $2
            Querylex.(*QueryLex).Query.Wheres = $4
            fmt.Printf("Select: %v\n", $1)
            fmt.Printf("Where: %v\n", $2)
        }
        |   SELECT selectClause SEMICOLON
        {
            Querylex.(*QueryLex).Query.Selects = $2
        }
        ;

selectClause    :   selectTermList
                {
                    fmt.Println("select")
                    $$ = $1
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
            |   ALL
            {
                $$ = SelectTerm{Tag: $1}
            }
            ;


whereClause :   whereTerm
            {
                $$ = WhereClause{SQL: $1.SQL}
            }
            |   whereTerm timeTerm
            {
                $$ = WhereClause{SQL: $1.SQL}
            }
            |   whereTerm OR whereClause
            {
                $$ = WhereClause{SQL: fmt.Sprintf(`(%s) or (%s)`, $1.SQL, $3.SQL)}
            }
            |   whereTerm AND whereClause
            {
                $$ = WhereClause{SQL: fmt.Sprintf(`(%s) and (%s)`, $1.SQL, $3.SQL)}
            }
            |   NOT whereClause
            {
                $$ = WhereClause{SQL: fmt.Sprintf(`not (%s)`, $2.SQL)}
            }
            ;


whereTerm   : LVALUE LIKE QSTRING
            {
                if $1 == "uuid" {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid LIKE %s`, $3)}
                } else {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval LIKE %s`, $1, $3)}
                }
            }
            | LVALUE EQ QSTRING
            {
                if $1 == "uuid" {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid = %s`, $3)}
                } else {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval = %s`, $1, $3)}
                }
            }
            | LVALUE NEQ QSTRING
            {
                if $1 == "uuid" {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid != %s`, $3)}
                } else {
                    $$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval != %s`, $1, $3)}
                }
            }
            | HAS LVALUE
            {
                if $2 == "uuid" {
                    $$ = WhereTerm{Key: $1, Op: $1, SQL: `data.uuid is not null`}
                } else {
                    $$ = WhereTerm{Key: $2, Op: $1, SQL: fmt.Sprintf(`data.dkey = "%s"`, $2)}
                }
            }
            | LPAREN whereClause RPAREN
            {
                $$ = WhereTerm{SQL: fmt.Sprintf(`(%s)`, $2.SQL)}
            }
            ;

timeTerm    :   IN timerange
            |   FOR timerange
            |   BEFORE timeref
            |   IBEFORE timeref
            |   AFTER timeref
            |   IAFTER timeref
            ;

timerange   :   LPAREN RPAREN
            ;

timeref     : abstime
            {
                $$ = $1
            }
            | abstime reltime
            {
                $$ = $1.Add($2)
            }
            ;

abstime     : NUMBER LVALUE
            {
                foundtime, err := parseAbsTime($1, $2)
                if err != nil {
                    Querylex.(*QueryLex).Error(fmt.Sprintf("Could not parse time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
                $$ = foundtime
            }
            | NUMBER
            {
                num, err := strconv.ParseInt($1, 10, 64)
                if err != nil {
                    Querylex.(*QueryLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $1, err.Error()))
                }
                $$ = _time.Unix(num, 0)
            }
            | QSTRING
            {
                found := false
                for _, format := range supported_formats {
                    t, err := _time.Parse(format, $1)
                    if err != nil {
                        continue
                    }
                    $$ = t
                    found = true
                    break
                }
                if !found {
                    Querylex.(*QueryLex).Error(fmt.Sprintf("No time format matching \"%v\" found", $1))
                }
            }
            | NOW
            {
                $$ = _time.Now()
            }
            ;

reltime     : NUMBER LVALUE
            {
                var err error
                $$, err = parseReltime($1, $2)
                if err != nil {
                    Querylex.(*QueryLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
            }
            | NUMBER LVALUE reltime
            {
                newDuration, err := parseReltime($1, $2)
                if err != nil {
                    Querylex.(*QueryLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
                $$ = addDurations(newDuration, $3)
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
            {Token: IBEFORE, Pattern: "ibefore"},
            {Token: AFTER, Pattern: "after"},
            {Token: IAFTER, Pattern: "iafter"},
            {Token: COMMA, Pattern: ","},
            {Token: AND, Pattern: "and"},
            {Token: AS, Pattern: "as"},
            {Token: TO, Pattern: "to"},
            {Token: FOR, Pattern: "for"},
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
