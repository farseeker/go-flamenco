package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var ctx context.Context
var fsClient *firestore.Client

func fsConnect(projectID, credsFilePath string) error {
	ctx = context.Background()

	credsFile, err := ioutil.ReadFile("firebase-auth.json")
	if err != nil {
		return err
	}
	creds, err := google.CredentialsFromJSON(ctx, credsFile, "https://www.googleapis.com/auth/datastore")
	if err != nil {
		return err
	}

	fsClient, err = firestore.NewClient(ctx, projectID, option.WithCredentials(creds))
	if err != nil {
		return err
	}

	return nil
}

func fsGetManager(id string) *firestore.DocumentRef {
	return fsClient.Doc(fmt.Sprintf("Managers/%s", id))
}

func fsGetLink(id string) *firestore.DocumentRef {
	return fsClient.Doc(fmt.Sprintf("ManagerLink/%s", id))
}
