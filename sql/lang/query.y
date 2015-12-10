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
%token <str> NOW SET AT BEFORE AFTER AND AS TO OR IN NOT FOR HAPPENS
%token <str> LPAREN RPAREN NEWLINE
%token <str> FIRST LAST IAFTER IBEFORE BETWEEN
%token NUMBER
%token SEMICOLON

%token <str> EQ NEQ COMMA ALL

%type <selectTermList> selectTermList selectClause
%type <selectTerm> selectTerm selectTermValue
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

selectTerm	:	FIRST selectTermValue
			{
				$2.Filter = FIRST
				$$ = $2
			}
			|	LAST selectTermValue
			{
				$2.Filter = LAST
				$$ = $2
			}
			|	ALL selectTermValue
			{
				$2.Filter = ALL
				$$ = $2
			}
			|	selectTermValue AT timeref
			{
				$1.Filter = AT
				$1.StartTime = $3
				$$ = $1
			}
			|	selectTermValue IAFTER timeref
			{
				$1.Filter = IAFTER
				$1.StartTime = $3
				$$ = $1
			}
			|	selectTermValue IBEFORE timeref
			{
				$1.Filter = IBEFORE
				$1.StartTime = $3
				$$ = $1
			}
			|	selectTermValue AFTER timeref
			{
				$1.Filter = AFTER
				$1.StartTime = $3
				$$ = $1
			}
			|	selectTermValue BEFORE timeref
			{
				$1.Filter = BEFORE
				$1.StartTime = $3
				$$ = $1
			}
			|	selectTermValue IN LPAREN timeref COMMA timeref RPAREN
			{
				$1.Filter = BETWEEN
				$1.StartTime = $4
				$1.EndTime = $6
				$$ = $1
			}
			;

selectTermValue	:	LVALUE
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

timeTerm	:	HAPPENS IN LPAREN timeref COMMA timeref RPAREN
			{
				template := `select uuid, dkey, timestamp as maxtime from data
				where timestamp >= "%s" and timestamp < "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $4.Format(_time.RFC3339), $6.Format(_time.RFC3339))
			}
			|	HAPPENS BEFORE timeref
			{
				template := `select uuid, dkey, timestamp as maxtime from data
				where timestamp <  "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $3.Format(_time.RFC3339))
			}
			|	AT timeref
			{
				template := `select distinct uuid, dkey, max(timestamp) as maxtime from data
				where timestamp <= "%s"
				group by dkey, uuid order by timestamp desc`
				$$ = fmt.Sprintf(template, $2.Format(_time.RFC3339))
			}
			|	HAPPENS AFTER timeref
			{
				template := `select uuid, dkey, timestamp as maxtime from data
				where timestamp >= "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $3.Format(_time.RFC3339))
			}
			|	FOR LPAREN timeref COMMA timeref RPAREN
			{
				template := `select uuid, dkey, timestamp as maxtime from data
				where timestamp >= "%s" and timestamp < "%s"
				order by timestamp desc`
				$$ = fmt.Sprintf(template, $3.Format(_time.RFC3339), $5.Format(_time.RFC3339))
			}
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
				now := Querylex.(*QueryLex).Now
				Querylex.(*QueryLex).Query.Now = now
				$$ = now
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

type SelectPredicate uint32
const (
	t_FIRST	SelectPredicate = FIRST
	t_LAST SelectPredicate = LAST
	t_ALL SelectPredicate = ALL
	t_AT SelectPredicate = AT
	t_IAFTER SelectPredicate = IAFTER
	t_AFTER SelectPredicate = AFTER
	t_IBEFORE SelectPredicate = IBEFORE
	t_BEFORE SelectPredicate = BEFORE
	t_BETWEEN SelectPredicate = BETWEEN
)

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
	Now		_time.Time
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
			{Token: FIRST, Pattern: "first"},
			{Token: LAST, Pattern: "last"},
			{Token: IBEFORE, Pattern: "ibefore"},
			{Token: BETWEEN, Pattern: "between"},
			{Token: HAPPENS, Pattern: "happens"},
			{Token: AT, Pattern: "at"},
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
	lex := &QueryLex{Query: &Query{}, Now: _time.Now(), querystring: s, scanner: scanner, Err: nil, lasttoken: "", tokens: []string{}}
	//lex.Rewrite(s)
	return lex
}

func (lex *QueryLex) Rewrite(s string) {
	var input = []toki.Token{WHERE}
	for lex.scanner.Peek() != nil {
		res := lex.scanner.Next()
		if res.Token == input[0] {
			pos := res.Pos.Column
			fmt.Println("match:", s[pos-1:pos-1+len("WHERE")])
			fmt.Println("WHERE")
			break
		}
	}
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
