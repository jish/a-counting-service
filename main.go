package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"os"
	"regexp"
)

func main() {
	port := os.Getenv("PORT")
	database_url := os.Getenv("DATABASE_URL")

	if port == "" {
		port = "3000"
	}

	if database_url == "" {
		panic("Must set DATABASE_URL (e.g. postgres://pqgotest:password@localhost:5432)")
	}

	db, err := sql.Open("postgres", database_url)
	if err != nil {
		panic(err)
	}

	createCounterTable(db)
	db.Close()

	fmt.Printf("Starting server on port %s\n", port)
	http.HandleFunc("/", makeDbHandler(database_url, rootRoute))
	http.ListenAndServe(":"+port, nil)
}

type Handler func(*sql.DB, http.ResponseWriter, *http.Request)

func makeDbHandler(databaseUrl string, fn Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			panic(err)
		}

		defer db.Close()
		fn(db, w, r)
	}
}

func rootRoute(db *sql.DB, response http.ResponseWriter, request *http.Request) {
	path := request.URL.Path

  fmt.Printf("\n%s\n", path)

	readCount, _ := regexp.Compile("^/count/([a-z]+)$")
	if readCount.MatchString(path) {
		key := readCount.FindStringSubmatch(path)[1]
		fmt.Printf("key: %s\n", key)
		count := getCountForKey(db, key)
		fmt.Printf("count: %d\n", count)
		io.WriteString(response, fmt.Sprintf("{ \"count\": %d }\n", count))
		return
	}

	incrementCount, _ := regexp.Compile("^/count/([a-z]+)/increment$")
	if incrementCount.MatchString(path) {
		key := incrementCount.FindStringSubmatch(path)[1]
		fmt.Printf("key: %s\n", key)

		count := getCountForKey(db, key)
		fmt.Printf("count: %d\n", count)
    count = count + 1
    fmt.Printf("new count: %d\n", count)

		if count == 1 {
			insertInitialCountForKey(db, key)
		} else {
		  setCountForKey(db, key, count)
		}

		io.WriteString(response, fmt.Sprintf("{ \"count\": %d }\n", count))
		return
	}

	io.WriteString(response, "<h1>Hello Human, Welcome to A Counting Service</h1>\n")
}

func getCountForKey(db *sql.DB, key string) int {
	var value int
	rows, err := db.Query("SELECT value FROM counters WHERE key = $1", key)

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		rows.Scan(&value)
	}
	rows.Close()

	return value
}

func insertInitialCountForKey(db *sql.DB, key string) {
  _, err := db.Exec("INSERT INTO counters VALUES($1, 1)", key)
  if err != nil {
    panic(err)
  }
}

func setCountForKey(db *sql.DB, key string, count int) {
  _, err := db.Exec("UPDATE counters SET value = $1 WHERE key = $2", count, key)
  if err != nil {
    panic(err)
  }
}

func createCounterTable(db *sql.DB) {
	tableExists := false

	// [schemaname tablename tableowner tablespace hasindexes hasrules hastriggers]
	rows, err := db.Query("SELECT tablename from pg_catalog.pg_tables WHERE tablename = 'counters'")
	if err != nil {
		panic(err)
	}

	// columns, _ := rows.Columns()
	// fmt.Println(rows)
	// fmt.Println(columns)

	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)

		// fmt.Println(tableName)

		if tableName == "counters" {
			fmt.Println("Found counters table")
			tableExists = true
		}
	}
	rows.Close()

	if !tableExists {
		fmt.Println("counters table not found, creating...")
		_, err := db.Exec("CREATE TABLE counters (key text, value bigint)")
		if err != nil {
			panic(err)
		}
	}
}
