package rcdb

import "database/sql"

func ConnectToDb() *sql.DB {
	connStr := "host=localhost user=postgres dbname=remotechess sslmode=disable password=admin"
	db, _ := sql.Open("postgres", connStr)

	return db
}
