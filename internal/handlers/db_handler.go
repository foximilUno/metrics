package handlers

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
)

func PingDb(driverName string, dbConnectionString string) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		fmt.Println(sql.Drivers())
		connect, err := sql.Open(driverName, dbConnectionString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = connect.Ping(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(connect *sql.DB) {
			err := connect.Close()
			if err != nil {
				log.Printf("error while closing conntection: %e\n", err)
			}
		}(connect)
	}
}
