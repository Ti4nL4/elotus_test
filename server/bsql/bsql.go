package bsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
}

// NewDB creates a new DB wrapper
func NewDB(db *sql.DB) *DB {
	return &DB{DB: db}
}

// Open opens a database connection with the given parameters
func Open(username, password, host, port, dbname string, maxIdleConnection, maxOpenConnection int) *DB {
	db := openSQL(username, password, host, port, dbname)
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetMaxOpenConns(maxOpenConnection)

	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	return NewDB(db)
}

// OpenDefault opens a database with default connection settings
func OpenDefault(host, username, password, dbname string) *DB {
	return Open(username, password, host, "", dbname, 40, 80)
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

// Insert inserts a row and returns the new ID
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

// Update updates rows matching conditionDict with values from dict
func (db *DB) Update(tableName string, conditionDict map[string]interface{}, dict map[string]interface{}) (err error) {
	var updateBuffer bytes.Buffer
	updateBuffer.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
	values := []interface{}{}
	var counter int
	var counterData int

	sortedList := sortDict(dict)
	for _, entry := range sortedList {
		key := entry.key
		value := entry.value
		updateBuffer.WriteString(fmt.Sprintf("%s = $%d", key, counter+1))
		if counterData != len(dict)-1 {
			updateBuffer.WriteString(", ")
		} else {
			updateBuffer.WriteString(" ")
		}
		values = append(values, value)
		counter++
		counterData++
	}

	if len(conditionDict) > 0 {
		updateBuffer.WriteString("WHERE ")
		var counterCond int
		sortedList := sortDict(conditionDict)
		for _, entry := range sortedList {
			key := entry.key
			value := entry.value
			updateBuffer.WriteString(fmt.Sprintf("%s = $%d", key, counter+1))
			if counterCond != len(sortedList)-1 {
				updateBuffer.WriteString(" AND ")
			}
			values = append(values, value)
			counter++
			counterCond++
		}
	}

	_, err = db.Exec(updateBuffer.String(), values...)
	return err
}

// QueryInt queries for a single integer value
func (db *DB) QueryInt(queryString string, args ...interface{}) (value int, err error) {
	var sqlValue sql.NullInt64
	row := db.QueryRow(queryString, args...)
	err = row.Scan(&sqlValue)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return int(sqlValue.Int64), err
}

// QueryInt64 queries for a single int64 value
func (db *DB) QueryInt64(queryString string, args ...interface{}) (value int64, err error) {
	var sqlValue sql.NullInt64
	row := db.QueryRow(queryString, args...)
	err = row.Scan(&sqlValue)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return sqlValue.Int64, err
}

// QueryString queries for a single string value
func (db *DB) QueryString(queryString string, args ...interface{}) (value string, err error) {
	var sqlValue sql.NullString
	row := db.QueryRow(queryString, args...)
	err = row.Scan(&sqlValue)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return sqlValue.String, err
}

// QueryBool queries for a single boolean value
func (db *DB) QueryBool(queryString string, args ...interface{}) (value bool, err error) {
	var sqlValue sql.NullBool
	row := db.QueryRow(queryString, args...)
	err = row.Scan(&sqlValue)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return sqlValue.Bool, err
}

// Entry represents a key-value pair for sorting
type Entry struct {
	key   string
	value interface{}
}

func sortDict(dict map[string]interface{}) []*Entry {
	attrs := []string{}
	for key := range dict {
		attrs = append(attrs, key)
	}
	sort.Strings(attrs)
	entries := []*Entry{}
	for _, key := range attrs {
		entries = append(entries, &Entry{
			key:   key,
			value: dict[key],
		})
	}
	return entries
}
