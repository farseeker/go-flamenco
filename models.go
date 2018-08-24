package main

import (
	"time"

	"github.com/farseeker/go-flamenco/flamenco-imports/websetup"
)

const jsonType = "application/json"

var config appConfig

type appConfig struct {
	FirestoreAuthFile  string
	FirestoreProjectID string
}

type managerDoc struct {
	ManagerID string    `firestore:"_ManagerID"`
	CreatedAt time.Time `firestore:"CreatedAt"`
	CreatedBy string    `firestore:"CreatedBy"`

	LinkExchange       linkExchangeConversation `firestore:"_LinkExchange"`
	ChooseConversation chooseConversation       `firestore:"_ChooseConversation"`
}

type linkExchangeConversation struct {
	ManagerPayload websetup.KeyExchangeRequest `firestore:"ManagerPayload"`
	ManagerKey     string                      `firestore:"ManagerKey"`
}

type chooseConversation struct {
	QueryHMAC                 string
	QueryIdentifier           string
	QueryReturn               string
	ResponseManagerUUIDString string
	ResponseHMAC              string
	ResponseURL               string
}

type updateDoc struct {
	ReceivedAt time.Time
	Count      int
	//Received   []flamenco.TaskUpdate
	//Sent     flamenco.TaskUpdateResponse `firestore:"Sent"`
}
