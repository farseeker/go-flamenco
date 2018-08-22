package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/farseeker/go-flamenco/flamenco-imports/flamenco"

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
	w.Write(responseJSON)
	defer tx.Rollback()
}

func startup(w http.ResponseWriter, r *http.Request) {
	_, tx, bkt, err := initManagerRequest(r, true)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to init request: %s", err.Error()))
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to read payload: %s", err.Error()))
		return
	}

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
