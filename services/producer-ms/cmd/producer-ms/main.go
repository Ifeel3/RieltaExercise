package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"producer-ms/internal/jsonRPC"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var file *os.File

func ExitIfError(err error) {
	if err == nil {
		return
	}
	log.Println(err)
	os.Exit(1)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func init() {
	filename := "logs/log_" + time.Now().Add(3*time.Hour).Format("2006-01-02_15-04-05")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	} else {
		log.SetOutput(file)
	}

	queueusername := os.Getenv("QUEUE_USERNAME")
	queuepassword := os.Getenv("QUEUE_PASSWORD")
	queuehost := os.Getenv("QUEUE_HOST")
	queueport := os.Getenv("QUEUE_PORT")
	url := fmt.Sprintf("amqp://%s:%s@%s:%s", queueusername, queuepassword, queuehost, queueport)
	for i := 1; i < 11; i++ {
		conn, err = amqp.Dial(url)
		if err != nil {
			log.Printf("Connection try %d failure: %s. Url: %s\n", i, err.Error(), url)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	ExitIfError(err)
}

func main() {
	defer file.Close()
	defer conn.Close()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ch, err := conn.Channel()
	ExitIfError(err)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		false,
		false,
		nil,
	)
	ExitIfError(err)

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	ExitIfError(err)

	req := jsonRPC.JsonRPCRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)
	var resp []byte
	if err != nil {
		log.Println(err)
		resp, _ = json.Marshal(jsonRPC.ResponseWithError(0, jsonRPC.PARSE_ERROR, err))
		w.Write(resp)
		return
	}

	corrId := randomString(32)
	tmp, _ := json.Marshal(req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = ch.PublishWithContext(ctx,
		"",
		"rpc_queue",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          tmp,
		})
	ExitIfError(err)

	for mes := range msgs {
		if corrId == mes.CorrelationId {
			w.Write(mes.Body)
			break
		}
	}
}
