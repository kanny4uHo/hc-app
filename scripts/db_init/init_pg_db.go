package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"healthcheckProject/internal/config"
)

func main() {
	file, err := os.ReadFile("/etc/initdb/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	appConfig := config.Config{}
	err = yaml.Unmarshal(file, &appConfig)

	if err != nil {
		log.Fatalf("failed to unmarshal config.yaml: %s", err)
	}

	pwdBytes, err := os.ReadFile("/etc/pgsecret/postgres-password")
	if err != nil {
		log.Fatal("failed to read postgres-password from /etc/pgsecret/postgres-password")
	}

	databaseString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		string(pwdBytes),
		appConfig.Database.Host,
		appConfig.Database.Port,
		appConfig.Database.DBName,
	)

	db, err := sql.Open("postgres", databaseString)
	if err != nil {
		log.Fatalf("failed to connect to userdb: %s", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping userdb: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    username TEXT not null,
    password_hash TEXT not null,
    email TEXT not null,
    first_name TEXT not null,
    last_name TEXT not null
)`)

	if err != nil {
		log.Printf("failed to create table 'users': %s", err)
	}

	_, err = db.Exec(`CREATE UNIQUE INDEX users_username_idx ON users(username)`)
	if err != nil {
		log.Printf("failed to create username index for table 'users': %s", err)
	}

	_, err = db.Exec(`CREATE UNIQUE INDEX users_email_idx ON users(email)`)
	if err != nil {
		log.Printf("failed to create email index for table 'users': %s", err)
	}
}
