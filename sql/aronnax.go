package main

import (
	query "./lang"
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

var ZERO_TIME = time.Time{}

type mysqlBackend struct {
	db *sql.DB
}

var tableCreate = `
CREATE TABLE data
(
    uuid CHAR(37) NOT NULL,
    dkey VARCHAR(128) NOT NULL,
    dval VARCHAR(128) NULL,
    timestamp TIMESTAMP(6) NOT NULL
);
`

var whereTemplate = `
select second.uuid, second.dkey, second.dval, second.timestamp
from (
   select data.uuid, data.dkey, data.dval, data.timestamp
   from data
   inner join
   (
        select distinct uuid, dkey, max(timestamp) as maxtime from data group by dkey, uuid order by timestamp desc
   ) sorted
   on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
   where data.dval is not null
) as second
right join
(
    %s
) internal
on internal.uuid = second.uuid;
`

var showQuery = flag.Bool("debug", false, "Show generated MySQL queries")
var httpPort = flag.Int("port", 2000, "Serve query interface on HTTP port")

func newBackend(user, password, database string) *mysqlBackend {
	var (
		db     *sql.DB
		err    error
		tables *sql.Rows
	)
	if db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?parseTime=true", user, password, database)); err != nil {
		log.Fatal(err)
	}

	// check for liveliness
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// check if table is created
	if tables, err = db.Query("show tables;"); err != nil {
		log.Fatal(err)
	}

	foundTable := false
	for tables.Next() && !foundTable {
		var name string
		if err := tables.Scan(&name); err != nil {
			log.Fatal(err)
		}
		foundTable = (name == "data")
	}

	// if table not found, create it!
	if !foundTable {
		if _, err = db.Exec(tableCreate); err != nil {
			log.Fatal(err)
		}
	}

	return &mysqlBackend{
		db: db,
	}
}

// remove data from table
func (mbd *mysqlBackend) RemoveData() error {
	_, err := mbd.db.Exec("DELETE FROM data;")
	return err
}

func (mbd *mysqlBackend) Insert(doc *Document) error {
	_, err := mbd.db.Exec(doc.GenerateInsertStatement(true))
	return err
}

func (mbd *mysqlBackend) InsertWithTimestamp(doc *Document, timestamp time.Time) error {
	_, err := mbd.db.Exec(doc.GenerateInsertStatementWithTimestamp(timestamp))
	return err
}

// passes through the error if it is nil
func (mbd *mysqlBackend) Eval(q *query.Query) ([]*Document, error) {
	var (
		docs    = []*Document{}
		err     error
		evalErr error
		rows    *sql.Rows
		tosend  string
	)
	// build SQL string using WHERE clause
	if q.Wheres.SQL != "" {
		tosend = fmt.Sprintf(whereTemplate, q.Wheres.SQL)
	}
	// print generated query if flag is set
	if *showQuery {
		fmt.Println(tosend)
	}
	// evaluate WHERE clause against the backend
	if rows, evalErr = mbd.db.Query(tosend); evalErr != nil {
		return docs, evalErr
	}

	// transform the returned rows into documents so they are easier to work with
	if docs, err = DocsFromRows(rows, q.Now); err != nil {
		return docs, err
	}

	// apply the select clause
	for idx, doc := range docs {
		// for each document, for each term in the select clause, pull out
		// which keys in the document match
		fmt.Println(idx, doc)
		doc.ApplySelect(q.Selects)
	}
	fmt.Println(q.Selects)
	return docs, err
}

func (mbd *mysqlBackend) Parse(querystring string) (*query.Query, error) {
	var parseErr error
	lex := query.NewQueryLexer(querystring)
	query.QueryParse(lex)
	if lex.Err != nil {
		parseErr = fmt.Errorf("ERROR %s %s", lex.Err, querystring)
	}
	return lex.Query, parseErr
}

// passes through the error if it is nil
func (mbd *mysqlBackend) EvalWhere(q *query.Query, err error) (*sql.Rows, time.Time, error) {
	var tosend string
	if q.Wheres.SQL != "" {
		tosend = fmt.Sprintf(whereTemplate, q.Wheres.SQL)
	}
	if *showQuery {
		fmt.Println(tosend)
	}
	if err == nil {
		rows, evalErr := mbd.db.Query(tosend)
		return rows, q.Now, evalErr
	} else {
		return nil, q.Now, err
	}
}

func (mbd *mysqlBackend) StartInteractive() {
	fi := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("aronnax> ")
		s, err := fi.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		parsed, parseErr := mbd.Parse(s)
		if parseErr != nil {
			log.Print("Error parse: ", parseErr)
			continue
		}
		docs, evalErr := mbd.Eval(parsed)
		if evalErr != nil {
			log.Print("Error eval: ", evalErr)
			continue
		}
		for _, doc := range docs {
			fmt.Println(doc.PrettyString())
		}
	}
}

func main() {
	flag.Parse()
	user := os.Getenv("ARONNAXUSER")
	pass := os.Getenv("ARONNAXPASS")
	dbname := os.Getenv("ARONNAXDB")
	backend := newBackend(user, pass, dbname)

	// setup HTTP server
	go backend.StartInteractive()
	StartHTTPServer(backend, *httpPort)
}
