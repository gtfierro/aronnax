package main

import (
	query "./lang"
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

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
select data.uuid, data.dkey, data.dval
from data
inner join
(
    select distinct uuid, dkey, max(timestamp) as maxtime from data group by dkey, uuid order by timestamp desc
) sorted
on data.uuid = sorted.uuid and data.dkey = sorted.dkey and data.timestamp = sorted.maxtime
where data.dval is not null
`

func newBackend(user, password, database string) *mysqlBackend {
	var (
		db     *sql.DB
		err    error
		tables *sql.Rows
	)
	if db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", user, password, database)); err != nil {
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
	result, err := mbd.db.Exec(doc.GenerateinsertStatement())
	fmt.Println(result)
	return err
}

func (mbd *mysqlBackend) Query(q *query.Query) *sql.Rows {
	tosend := whereTemplate
	tosend += "and " + q.Wheres.SQL
	fmt.Println(tosend)
	rows, err := mbd.db.Query(tosend)
	if err != nil {
		log.Fatal(err)
	}
	return rows
}

func (mbd *mysqlBackend) StartInteractive() {
	fi := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("aronnax> ")
		s, err := fi.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		lex := query.NewQueryLexer(s)
		query.QueryParse(lex)
		fmt.Println("ERROR", lex.Err)

		rows := mbd.Query(lex.Query)
		if cols, err := rows.Columns(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("cols: ", cols)
		}

		for rows.Next() {
			var (
				uuid string
				key  string
				val  string
			)
			if err := rows.Scan(&uuid, &key, &val); err != nil {
				log.Fatal(err)
			} else {
				fmt.Printf("-> %s %s %s\n", uuid, key, val)
			}
		}

	}
}

func main() {
	user := os.Getenv("ARONNAXUSER")
	pass := os.Getenv("ARONNAXPASS")
	dbname := os.Getenv("ARONNAXDB")
	backend := newBackend(user, pass, dbname)
	backend.StartInteractive()
}
