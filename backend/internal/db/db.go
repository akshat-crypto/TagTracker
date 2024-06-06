package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {

	//Load the Env Variables
	//TODO: Get the variables from the SEcrets manager

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading.env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbSourceName := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err = sql.Open("postgres", dbSourceName)
	if err != nil {
		log.Fatal(err)
	}

	//TODO: You can change the max connections to this db by change the parameters go to this blog https://pkg.go.dev/database/sql#Open
	// db.SetConnMaxLifetime(0)
	// db.SetMaxIdleConns(50)
	// db.SetMaxOpenConns(50)

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

}

func GetDB() *sql.DB {
	return db
}
