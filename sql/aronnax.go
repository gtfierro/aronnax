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
    uuid CHAR(16) NOT NULL,
    dkey VARCHAR(20) NOT NULL,
    dval VARCHAR(20) NULL,
    timestamp TIMESTAMP NOT NULL
);
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

	} else {
		log.Println("Found table!")
	}

	fmt.Println(tables.Columns())
	fmt.Println(tables.Next())
	fmt.Println(tables.Err())

	return &mysqlBackend{
		db: db,
	}
}

func main() {
	user := os.Getenv("ARONNAXUSER")
	pass := os.Getenv("ARONNAXPASS")
	dbname := os.Getenv("ARONNAXDB")
	backend := newBackend(user, pass, dbname)
	fmt.Println(backend)

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
	}
}
