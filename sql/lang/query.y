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
	selectTerm	SelectTerm
	selectTermList	[]SelectTerm
	whereTerm  WhereTerm
	whereClause  WhereClause
	time _time.Time
	timediff _time.Duration
}

%token <str> SELECT DISTINCT WHERE
%token <str> LVALUE QSTRING LIKE HAS
%token <str> NOW SET AT BEFORE AFTER AND AS TO OR IN NOT FOR
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
%type <str> NUMBER timeTerm
// timeTerm is a string because it returns a SQL clause

%right EQ

%%

query	:	SELECT selectClause WHERE whereClause SEMICOLON
		{
			Querylex.(*QueryLex).Query.Selects = $2
			Querylex.(*QueryLex).Query.Wheres = $4
		}
		|	SELECT selectClause SEMICOLON
		{
			Querylex.(*QueryLex).Query.Selects = $2
		}
		;

selectClause	:	selectTermList
				{
					$$ = $1
				}
				;

selectTermList	:	selectTerm
				{
					$$ = []SelectTerm{$1}
				}
				|	selectTerm	COMMA selectTermList
				{
					$$ = append([]SelectTerm{$1}, $3...)
				}
				|	DISTINCT LVALUE
				{
					$$ = []SelectTerm{{Tag: $2}}
				}
				;

selectTerm	:	LVALUE
			{
				$$ = SelectTerm{Tag: $1}
			}
			|	ALL
			{
				$$ = SelectTerm{Tag: $1}
			}
			;


whereClause :	whereTerm
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				if ($1.IsPredicate) {
					$$ = WrapTermInSelect($1.SQL, $1.Letter)
				} else { // have a full select clause
					$$ = WhereClause{SQL: $1.SQL, Letter: $1.Letter}
				}
			}
			|	whereTerm timeTerm
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				if ($1.IsPredicate) {
					$$ = WrapTermInSelectWithTime($1.SQL, $1.Letter, $2)
				} else { // have a full select clause
					$$ = WhereClause{SQL: $1.SQL, Letter: $1.Letter}
				}
			}
			|	whereTerm OR whereClause
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				var firstTerm = $1.GetClause()
				sql := fmt.Sprintf(`
select distinct uuid
from
%s as %s
union
%s`, firstTerm.SQL, firstTerm.Letter, $3.SQL)
				$$ = WhereClause{SQL: sql, Letter: firstTerm.Letter}
			}
			|	whereTerm timeTerm OR whereClause
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				var firstTerm = $1.GetClauseWithTime($2)
				sql := fmt.Sprintf(`
select distinct uuid
from
%s as %s
union
%s`, firstTerm.SQL, firstTerm.Letter, $4.SQL)
				$$ = WhereClause{SQL: sql, Letter: firstTerm.Letter}
			}
			|	whereTerm AND whereClause
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				var firstTerm = $1.GetClause()
				sql := fmt.Sprintf(`
select distinct %s.uuid
from
%s as %s
inner join
(%s) as %s
on %s.uuid = %s.uuid`, firstTerm.Letter, firstTerm.SQL, firstTerm.Letter, $3.SQL, $3.Letter, firstTerm.Letter, $3.Letter)
				$$ = WhereClause{SQL: sql, Letter: firstTerm.Letter}
			}
			|	whereTerm timeTerm AND whereClause
			{
				letter := Querylex.(*QueryLex).NextLetter()
				$1.Letter = letter
				var firstTerm = $1.GetClauseWithTime($2)
				sql := fmt.Sprintf(`
select distinct %s.uuid
from
%s as %s
inner join
(%s) as %s
on %s.uuid = %s.uuid`, firstTerm.Letter, firstTerm.SQL, firstTerm.Letter, $4.SQL, $4.Letter, firstTerm.Letter, $4.Letter)
				$$ = WhereClause{SQL: sql, Letter: firstTerm.Letter}
			}
			|	NOT whereClause
			{
				sql := fmt.Sprintf(`
select distinct data.uuid
from
data
where data.uuid not in (%s)`, $2.SQL)
				$$ = WhereClause{SQL: sql, Letter: $2.Letter}
			}
			;


whereTerm	: LVALUE LIKE QSTRING
			{
				if $1 == "uuid" {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid LIKE %s`, $3), IsPredicate: true}
				} else {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval LIKE %s`, $1, $3), IsPredicate: true}
				}
			}
			| LVALUE EQ QSTRING
			{
				if $1 == "uuid" {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid = %s`, $3), IsPredicate: true}
				} else {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval = %s`, $1, $3), IsPredicate: true}
				}
			}
			| LVALUE NEQ QSTRING
			{
				if $1 == "uuid" {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.uuid != %s`, $3), IsPredicate: true}
				} else {
					$$ = WhereTerm{Key: $1, Op: $2, Val: $3, SQL: fmt.Sprintf(`data.dkey = "%s" and data.dval != %s`, $1, $3), IsPredicate: true}
				}
			}
			| HAS LVALUE
			{
				if $2 == "uuid" {
					$$ = WhereTerm{Key: $1, Op: $1, SQL: `data.uuid is not null`, IsPredicate: true}
				} else {
					$$ = WhereTerm{Key: $2, Op: $1, SQL: fmt.Sprintf(`data.dkey = "%s"`, $2), IsPredicate: true}
				}
			}
			| LPAREN whereClause RPAREN
			{
				$$ = WhereTerm{SQL: fmt.Sprintf(`(%s)`, $2.SQL), IsPredicate: false}
			}
			;

timeTerm	:	IN LPAREN timeref COMMA timeref RPAREN
			{
				//TODO: fix!
				$$ = fmt.Sprintf(timePredicateRange, ">=", _time.Now())
				template := `select distinct uuid, dkey, timestamp as maxtime from data
				where timestamp >= "%s" and timestamp < "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $3.Format(_time.RFC3339), $5.Format(_time.RFC3339))
			}
			|	BEFORE timeref
			{
				template := `select distinct uuid, dkey, timestamp as maxtime from data
				where timestamp <= "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $2.Format(_time.RFC3339))
			}
			|	AT timeref
			{
				template := `select distinct uuid, dkey, max(timestamp) as maxtime from data
				where timestamp <= "%s"
				group by dkey, uuid order by timestamp desc`
				$$ = fmt.Sprintf(template, $2.Format(_time.RFC3339))
			}
			|	AFTER timeref
			{
				template := `select distinct uuid, dkey, timestamp as maxtime from data
				where timestamp >= "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $2.Format(_time.RFC3339))
			}
//			  |   FOR timeRange
			;

timeref		: abstime
			{
				$$ = $1
			}
			| abstime reltime
			{
				$$ = $1.Add($2)
			}
			;

abstime		: NUMBER LVALUE
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

reltime		: NUMBER LVALUE
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
	Query	*Query
	querystring   string
	scanner *toki.Scanner
	lasttoken string
	tokens	[]string
	innertable	int
	Err   error
}

func (ql *QueryLex) NextLetter() string {
	var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ql.innertable += 1
	return string(alphabet[ql.innertable-1])
}

var termTemplate = `
	(
		select distinct data.uuid
		from data
		inner join
		(
			select distinct uuid, dkey, max(timestamp) as maxtime from data
			group by dkey, uuid order by timestamp desc
		) sorted
		on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
		where data.dval is not null
			%s
	) as %s
`

var termTemplateUnion = `
	(
		select distinct data.uuid
		from data
		inner join
		(
			select distinct uuid, dkey, max(timestamp) as maxtime from data
			group by dkey, uuid order by timestamp desc
		) sorted
		on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
		where data.dval is not null
			%s
	)
`

var timePredicateSingle = `
	select distinct uuid, dkey, max(timestamp) as maxtime from data
	where timestamp %s "%s"
	group by dkey, uuid order by timestamp desc
`

var timePredicateRange = `
	select distinct uuid, dkey, max(timestamp) as maxtime from data
	where timestamp %s "%s" and timestamp %s "%s"
	group by dkey, uuid order by timestamp desc
`

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
			{Token: AT, Pattern: "at"},
			{Token: AFTER, Pattern: "after"},
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
