package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"consumer-ms/internal/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

var pool *pgxpool.Pool
var conn *amqp.Connection
var file *os.File = nil

func ExitIfError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
	os.Exit(1)
}

func init() {
	filename := "logs/log_" + time.Now().Add(3*time.Hour).Format("2006-01-02_15-04-05")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	} else {
		log.SetOutput(file)
	}

	dbusername := os.Getenv("DB_USERNAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbhost := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	database := os.Getenv("DB_DATABASE")
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbusername, dbpassword, dbhost, dbport, database)
	for i := 1; i < 11; i++ {
		pool, err = pgxpool.New(context.Background(), url)
		if err != nil {
			log.Printf("Connection try %d failure: %s. Url: %s\n", i, err.Error(), url)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	ExitIfError(err)

	queueusername := os.Getenv("QUEUE_USERNAME")
	queuepassword := os.Getenv("QUEUE_PASSWORD")
	queuehost := os.Getenv("QUEUE_HOST")
	queueport := os.Getenv("QUEUE_PORT")
	url = fmt.Sprintf("amqp://%s:%s@%s:%s", queueusername, queuepassword, queuehost, queueport)
	for i := 1; i < 11; i++ {
		conn, err = amqp.Dial(url)
		if err != nil {
			log.Printf("Connection try %d failure: %s. Url: %s\n", i, err.Error(), url)
			time.Sleep(1 * time.Second)
		}
	}
	ExitIfError(err)
}

func main() {
	defer file.Close()
	defer pool.Close()
	defer conn.Close()

	err := handlers.Loop(pool, conn)
	ExitIfError(err)
}
