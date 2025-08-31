package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		DBName   string `yaml:"db_name"`
	} `yaml:"database"`
	UserappPassword         string `yaml:"userapp_password"`
	OrderappPassword        string `yaml:"orderapp_password"`
	NotificationappPassword string `yaml:"notificationapp_password"`
}

func main() {
	file, err := os.ReadFile("/etc/initdb/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	appConfig := Config{}
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

	db.Exec(`CREATE DATABASE userapp`)

	databaseString = fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		string(pwdBytes),
		appConfig.Database.Host,
		appConfig.Database.Port,
		"userapp",
	)

	db, err = sql.Open("postgres", databaseString)
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

	db.Exec(fmt.Sprintf("CREATE USER userapp WITH PASSWORD '%s'", appConfig.UserappPassword))
	db.Exec("GRANT ALL ON TABLE users TO userapp")

	db.Exec(`CREATE DATABASE orderapp`)

	databaseString = fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		string(pwdBytes),
		appConfig.Database.Host,
		appConfig.Database.Port,
		"orderapp",
	)

	db, err = sql.Open("postgres", databaseString)
	if err != nil {
		log.Fatalf("failed to connect to userdb: %s", err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping userdb: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
    id serial PRIMARY KEY,
    item TEXT not null,
    price INT not null,
    status TEXT not null,
    user_id INT not null
)`)

	db.Exec(fmt.Sprintf("CREATE USER orderapp WITH PASSWORD '%s'", appConfig.OrderappPassword))
	db.Exec("GRANT ALL ON TABLE orders TO orderapp")

	db.Exec(`CREATE DATABASE notificationapp`)

	databaseString = fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		appConfig.Database.Username,
		string(pwdBytes),
		appConfig.Database.Host,
		appConfig.Database.Port,
		"notificationapp",
	)

	db, err = sql.Open("postgres", databaseString)
	if err != nil {
		log.Fatalf("failed to connect to userdb: %s", err)
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping userdb: %s", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
    id serial PRIMARY KEY,
    timestamp timestamp not null,
    order_id INT NOT NULL,
    recipient_email TEXT not null
)`)

	db.Exec(fmt.Sprintf("CREATE USER notificationapp WITH PASSWORD '%s'", appConfig.NotificationappPassword))
	db.Exec("GRANT ALL ON TABLE notifications TO notificationapp")
}
