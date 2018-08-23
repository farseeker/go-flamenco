package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/farseeker/go-flamenco/flamenco-imports/flamenco"
	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

func initManagerRequest(r *http.Request, rw bool) (string, *bolt.Tx, *bolt.Bucket, error) {
	vars := mux.Vars(r)
	identity := vars["identity"]
	if identity == "" {
		return "", nil, nil, fmt.Errorf("Unable to determine identity from request %s", r.RequestURI)

	}

	tx, err := db.Begin(rw)
	if err != nil {
		tx.Rollback()
		return "", nil, nil, err
	}

	bkt := tx.Bucket([]byte(fmt.Sprintf("manager-%s", identity)))
	if err != nil {
		tx.Rollback()
		return "", nil, nil, err
	}

	if bkt == nil {
		tx.Rollback()
		return "", nil, nil, fmt.Errorf("Unable to open bucket for identity %s", identity)
	}

	if rw {
		time, _ := time.Now().GobEncode()
		bkt.Put([]byte("last-seen"), time)
	}

	return identity, tx, bkt, nil
}

func taskUpdateBatch(w http.ResponseWriter, r *http.Request) {
	_, tx, _, err := initManagerRequest(r, true)
	defer tx.Rollback()
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}

	if err := tx.Commit(); err != nil {
		httpError(w, fmt.Errorf("Unable to commit manager bucket: %s", err.Error()))
		return
	}

	updates := []flamenco.TaskUpdate{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&updates)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}

	// Eventually we'll do something here, but for now let's just get it working
	response := flamenco.TaskUpdateResponse{}
	responseJSON, _ := json.Marshal(response)
	w.Header().Set("Content-Type", jsonType)
	w.Write(responseJSON)
}

func startup(w http.ResponseWriter, r *http.Request) {
	_, tx, bkt, err := initManagerRequest(r, true)
	defer tx.Rollback()
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

	bkt.Put([]byte("startup-notification"), bodyBytes)

	//fmt.Println(startupNotification.ManagerURL)
	//fmt.Println(startupNotification.NumberOfWorkers)
	//fmt.Println(startupNotification.PathReplacementByVarname)
	//fmt.Println(startupNotification.VariablesByVarname)

	if err := tx.Commit(); err != nil {
		httpError(w, fmt.Errorf("Unable to commit manager bucket: %s", err.Error()))
		return
	}
}

func depsgraph(w http.ResponseWriter, r *http.Request) {
	identity, tx, _, err := initManagerRequest(r, true)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}
	defer tx.Rollback()

	lastUpdated := r.Header.Get("X-Flamenco-If-Updated-Since")
	fmt.Println("X-Flamenco-If-Updated-Since: ", lastUpdated)

	//http.StatusNotModified
	//http.StatusNoContent

	scheduledTasks := flamenco.ScheduledTasks{
		Depsgraph: []flamenco.Task{},
	}

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
