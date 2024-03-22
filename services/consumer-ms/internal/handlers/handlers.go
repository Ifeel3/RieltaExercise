package handlers

import (
	"consumer-ms/internal/dbfunc"
	"consumer-ms/internal/jsonRPC"
	"consumer-ms/internal/structs"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Loop(p *pgxpool.Pool, c *amqp.Connection) (err error) {

	ch, err := c.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for message := range msgs {
		go handler(p, &message, ch)
	}
	return
}

func handler(p *pgxpool.Pool, m *amqp.Delivery, ch *amqp.Channel) {
	req := jsonRPC.JsonRPCRequest{}
	err := json.Unmarshal(m.Body, &req)
	if err != nil {
		log.Println(err)
		resp, _ := json.Marshal(jsonRPC.ResponseWithError(0, jsonRPC.PARSE_ERROR, err))
		err = send(m, ch, resp)
		if err != nil {
			log.Println(err)
		}
	}

	var resp []byte
	switch req.Method {
	case "add":
		if len(req.Params) != 2 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong number of parameters")))
			break
		}
		name, ok1 := req.Params[0].(string)
		balance, ok2 := req.Params[1].(float64)
		if !ok1 || !ok2 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong parameter")))
			break
		}
		err := dbfunc.AddUser(p, &structs.User{Name: name, Balance: int(balance)})
		if err != nil {
			log.Println(err)
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INTERNAL_ERROR, err))
			break
		}
		resp, _ = json.Marshal(jsonRPC.NewResponse(req.Id, "OK"))
	case "delete":
		if len(req.Params) != 1 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong number of parameters")))
			break
		}
		name, ok := req.Params[0].(string)
		if !ok {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong parameter")))
			break
		}
		err := dbfunc.DeleteUser(p, name)
		if err != nil {
			log.Println(err)
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INTERNAL_ERROR, err))
			break
		}
		resp, _ = json.Marshal(jsonRPC.NewResponse(req.Id, "OK"))
	case "getuserbyname":
		if len(req.Params) != 1 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong number of parameters")))
			break
		}
		name, ok := req.Params[0].(string)
		if !ok {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong parameter")))
			break
		}
		user, err := dbfunc.GetUserByName(p, name)
		if err != nil {
			log.Println(err)
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INTERNAL_ERROR, err))
			break
		}
		resp, _ = json.Marshal(jsonRPC.NewResponse(req.Id, user))
	case "transaction":
		if len(req.Params) != 3 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong number of parameters")))
			break
		}
		from, ok1 := req.Params[0].(string)
		to, ok2 := req.Params[1].(string)
		count, ok3 := req.Params[2].(float64)
		if !ok1 || !ok2 || !ok3 {
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INVALID_PARAMS, fmt.Errorf("wrong parameter")))
			break
		}
		err := dbfunc.Transaction(p, from, to, int(count))
		if err != nil {
			log.Println(err)
			resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.INTERNAL_ERROR, err))
			break
		}
		resp, _ = json.Marshal(jsonRPC.NewResponse(req.Id, "OK"))
	default:
		resp, _ = json.Marshal(jsonRPC.ResponseWithError(req.Id, jsonRPC.METHOD_NOT_FOUND, fmt.Errorf("unknown method")))
	}

	err = send(m, ch, resp)
	if err != nil {
		log.Println(err)
	}
}

func send(m *amqp.Delivery, ch *amqp.Channel, data []byte) error {
	err := ch.PublishWithContext(context.Background(),
		"",
		m.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: m.CorrelationId,
			Body:          data,
		})
	if err != nil {
		return err
	}
	m.Ack(false)
	return nil
}
