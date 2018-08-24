package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/farseeker/go-flamenco/flamenco-imports/flamenco"
	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

func initManagerRequest(r *http.Request) (string, *firestore.DocumentRef, error) {
	vars := mux.Vars(r)
	identity := vars["identity"]
	if identity == "" {
		return "", nil, fmt.Errorf("Unable to determine identity from request %s", r.RequestURI)
	}

	managerDoc := fsGetManager(identity)
	managerDoc.Update(ctx, []firestore.Update{firestore.Update{
		Path:  "LastSeen",
		Value: time.Now(),
	}})

	return identity, managerDoc, nil
}

func taskUpdateBatch(w http.ResponseWriter, r *http.Request) {
	_, managerDoc, err := initManagerRequest(r)

	updates := []flamenco.TaskUpdate{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&updates)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}

	var updatesManaged []bson.ObjectId
	updateCollection := managerDoc.Collection("UpdateBatches")
	timeNow := time.Now()
	timeNowString := strconv.FormatInt(timeNow.Unix(), 10)
	if len(updates) > 0 {
		updateDoc := updateDoc{
			ReceivedAt: timeNow,
			Count:      len(updates),
			//Received:   updates,
		}
		updateBatchDocRef := updateCollection.Doc(timeNowString)
		updateBatchDocRef.Create(ctx, updateDoc)
		if err != nil {
			httpError(w, fmt.Errorf("Unable to create update doc: %s", err.Error()))
			return
		}

		for _, update := range updates {
			updateID := update.ID.Hex()
			updateDocRef := updateBatchDocRef.Collection("Updates").Doc(updateID)

			_, err := updateDocRef.Create(ctx, update)
			if err != nil {
				httpError(w, fmt.Errorf("Unable to log update: %s", err.Error()))
				continue
			}

			_, err = updateDocRef.Update(ctx, []firestore.Update{firestore.Update{
				Path:  "ID",
				Value: update.ID.Hex(),
			}, firestore.Update{
				Path:  "TaskID",
				Value: update.TaskID.Hex(),
			}})
			if err != nil {
				httpError(w, fmt.Errorf("Unable to log update updates: %s", err.Error()))
				continue
			}

			updatesManaged = append(updatesManaged, update.ID)
		}
	}

	response := flamenco.TaskUpdateResponse{
		ModifiedCount:    len(updatesManaged),
		HandledUpdateIds: updatesManaged,
	}

	responseJSON, _ := json.Marshal(response)
	w.Header().Set("Content-Type", jsonType)
	w.Write(responseJSON)
}

func startup(w http.ResponseWriter, r *http.Request) {
	_, managerDoc, err := initManagerRequest(r)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to read payload: %s", err.Error()))
		return
	}

	//Decode so we know we received a valid response
	startupNotification := flamenco.StartupNotification{}
	err = json.Unmarshal(bodyBytes, &startupNotification)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}

	startupCollection := managerDoc.Collection("StartupNotices")
	timeNow := strconv.FormatInt(time.Now().Unix(), 10)
	startupCollection.Doc(timeNow).Create(ctx, startupNotification)
}

func depsgraph(w http.ResponseWriter, r *http.Request) {
	_, _, err := initManagerRequest(r)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}

	lastUpdated := r.Header.Get("X-Flamenco-If-Updated-Since")
	fmt.Println("X-Flamenco-If-Updated-Since: ", lastUpdated)

	//http.StatusNotModified
	//http.StatusNoContent

	scheduledTasks := flamenco.ScheduledTasks{
		Depsgraph: []flamenco.Task{},
	}

	/*
		now := time.Now()
		scheduledTasks.Depsgraph = append(scheduledTasks.Depsgraph, flamenco.Task{
			Name:        "Test Task",
			Manager:     bson.ObjectIdHex(identity),
			Job:         bson.ObjectIdHex(newUUIDForBson()),
			Project:     bson.ObjectIdHex(newUUIDForBson()),
			ID:          bson.ObjectIdHex(newUUIDForBson()),
			User:        bson.ObjectIdHex(newUUIDForBson()),
			TaskType:    "echo",
			Activity:    "queued",
			Etag:        "abcd",
			Priority:    10,
			JobPriority: 10,
			JobType:     "test",
			LastUpdated: &now,
			Status:      flamenco.StatusQueued,
			Commands: []flamenco.Command{
				flamenco.Command{Name: "echo", Settings: bson.M{"message": "Running Blender from {blender}"}},
				flamenco.Command{Name: "sleep", Settings: bson.M{"time_in_seconds": 3}},
			},
		})
	*/

	if len(scheduledTasks.Depsgraph) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	scheduledJSON, err := json.Marshal(scheduledTasks)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}

	fmt.Println(string(scheduledJSON))

	w.Header().Set("Content-Type", jsonType)
	w.Write(scheduledJSON)
	return
}
