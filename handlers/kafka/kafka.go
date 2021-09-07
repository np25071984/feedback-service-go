package handlers

import (
	"context"
	"encoding/json"
	repository "feedback-service-go/repositories"
	"fmt"
	"log"
	"os"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic         = "test"
	brokerAddress = "kafka:9092"
)

func Consume(ctx context.Context, repo repository.Repository) {
	// create a new logger that outputs to stdout
	// and has the `kafka reader` prefix
	l := log.New(os.Stdout, "kafka reader: ", 0)
	// initialize a new reader with the brokers and topic
	// the groupID identifies the consumer and prevents
	// it from receiving duplicate messages
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		// GroupID: "feedback-group",
		// assign the logger to the reader
		Logger: l,
	})
	for {
		// the `ReadMessage` method blocks until we receive the next event
		rawMsg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}

		var inputRequest kafkaRequest
		err = json.Unmarshal(rawMsg.Value, &inputRequest)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("type: %T, value: %v", inputRequest.Payload, inputRequest.Payload)

		switch inputRequest.Action {
		case "create-action":
			createFeedback(inputRequest.Payload, repo)
		case "update-action":
			updateFeedback(inputRequest.Payload, repo)
		case "delete-action":
			deleteFeedback(inputRequest.Payload, repo)
		default:
			panic("Unknown action")
		}

		fmt.Println("sucessfully got:", string(rawMsg.Value))
	}
}

func createFeedback(payload json.RawMessage, repo repository.Repository) {
	// TODO: check if inputFeedback.Version valid
	var inputData createRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	request := repository.FeedbackRequest{
		ParentId:   inputData.ParentId,
		SenderId:   inputData.SenderId,
		ReceiverId: inputData.ReceiverId,
		TradeId:    inputData.TradeId,
		Message:    inputData.Message,
		Type:       inputData.Type,
		CreatedAt:  inputData.CreatedAt,
	}

	if len(request.Validate()) == 0 {
		repo.Create(&request)
	} else {
		log.Println("request validation error")
	}
}

func updateFeedback(payload json.RawMessage, repo repository.Repository) {
	// TODO: check if inputFeedback.Version valid
	var inputData updateRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	request := repository.FeedbackRequest{
		ParentId:   inputData.ParentId,
		SenderId:   inputData.SenderId,
		ReceiverId: inputData.ReceiverId,
		TradeId:    inputData.TradeId,
		Message:    inputData.Message,
		Type:       inputData.Type,
		CreatedAt:  inputData.CreatedAt,
	}

	if len(request.Validate()) == 0 {
		repo.Update(inputData.Id, &request)
	} else {
		log.Println("request validation error")
	}
}

func deleteFeedback(payload json.RawMessage, repo repository.Repository) {
	// TODO: check if inputFeedback.Version valid
	var inputData deleteRequest
	err := json.Unmarshal(payload, &inputData)
	if err != nil {
		panic(err.Error())
	}

	repo.Delete(inputData.Id)
}

type kafkaRequest struct {
	Action  string
	Payload json.RawMessage
}

type createRequest struct {
	Version    string
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}

type updateRequest struct {
	Version    string
	Id         int
	ParentId   int `json:"parent_id"`
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
	TradeId    int `json:"trade_id"`
	Message    string
	Type       string
	CreatedAt  string `json:"created_at"`
}

type deleteRequest struct {
	Version string
	Id      int
}
