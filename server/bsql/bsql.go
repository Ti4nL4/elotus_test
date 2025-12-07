package bsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{DB: db}
}

func Open(username, password, host, port, dbname string, maxIdleConnection, maxOpenConnection int) *DB {
	db := openSQL(username, password, host, port, dbname)
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetMaxOpenConns(maxOpenConnection)

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	return NewDB(db)
}

func openSQL(username, password, host, port, dbname string) *sql.DB {
	connectionStrTokens := []string{
		"sslmode=disable",
		"binary_parameters=yes",
	}

	if username != "" {
		connectionStrTokens = append(connectionStrTokens, fmt.Sprintf("user=%s", username))
	}

	if password != "" {
		connectionStrTokens = append(connectionStrTokens, fmt.Sprintf("password=%s", password))
	}

	if host != "" {
		connectionStrTokens = append(connectionStrTokens, fmt.Sprintf("host=%s", host))
	}

	if port != "" {
		connectionStrTokens = append(connectionStrTokens, fmt.Sprintf("port=%s", port))
	}

	if dbname != "" {
		connectionStrTokens = append(connectionStrTokens, fmt.Sprintf("dbname=%s", dbname))
	}

	connectionStr := strings.Join(connectionStrTokens, " ")
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}
	return db
}

func (db *DB) Insert(tableName string, dict map[string]interface{}) (id int64, err error) {
	var keyBuffer bytes.Buffer
	var valueBuffer bytes.Buffer
	keyBuffer.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
	valueBuffer.WriteString(") VALUES (")
	values := []interface{}{}
	var counter int

	sortedList := sortDict(dict)
	for _, entry := range sortedList {
		key := entry.key
		value := entry.value
		keyBuffer.WriteString(key)
		valueBuffer.WriteString(fmt.Sprintf("$%d", counter+1))
		if counter != len(dict)-1 {
			keyBuffer.WriteString(", ")
			valueBuffer.WriteString(", ")
		}
		values = append(values, value)
		counter++
	}
	valueBuffer.WriteString(") RETURNING id;")
	keyBuffer.WriteString(valueBuffer.String())

	err = db.QueryRow(keyBuffer.String(), values...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, err
}

type entry struct {
	key   string
	value interface{}
}

func sortDict(dict map[string]interface{}) []*entry {
	attrs := []string{}
	for key := range dict {
		attrs = append(attrs, key)
	}
	sort.Strings(attrs)
	entries := []*entry{}
	for _, key := range attrs {
		entries = append(entries, &entry{
			key:   key,
			value: dict[key],
		})
	}
	return entries
}
