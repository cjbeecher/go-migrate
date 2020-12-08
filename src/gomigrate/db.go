package gomigrate

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq" // Used to create a postgres connection
)

func postgresConnect(parameters *Parameters) (*sql.DB, error) {
	connInfo := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		parameters.Host, parameters.Port, parameters.Database, parameters.Username, parameters.Password,
	)
	return sql.Open("postgres", connInfo)
}

// ConnectDB to a database (requires passwordless authentication)
func ConnectDB(database string, parameters *Parameters) (*sql.DB, error) {
	if database == "postgres" {
		return postgresConnect(parameters)
	}
	return nil, errors.New("Unknown/unsupported database type \"" + database + "\"")
}
