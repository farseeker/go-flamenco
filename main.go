package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var db *bolt.DB

func main() {
	var err error
	db, err = bolt.Open("go-flamenco.db", 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		fmt.Println("Unable to open bolt db")
		return
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/flamenco/managers/link/choose", linkChoose)
	r.HandleFunc("/api/flamenco/managers/link/exchange", linkExchange)
	r.HandleFunc("/api/flamenco/managers/link/reset-token", linkReset)
	r.HandleFunc("/api/flamenco/managers/{identity}/task-update-batch", taskUpdateBatch)
	r.HandleFunc("/api/flamenco/managers/{identity}/startup", startup)
	r.HandleFunc("/api/flamenco/managers/{identity}/depsgraph", depsgraph)
	r.HandleFunc("/api/flamenco/tasks/{taskid}", taskByID)
	//

	r.PathPrefix("/").HandlerFunc(logDefault)
	srv := &http.Server{
		Handler: r,
		Addr:    ":8123",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func logDefault(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	output := fmt.Sprintf("------\n%s\n%s\n", r.RequestURI, string(body))
	ioutil.WriteFile("log.txt", []byte(output), 0744)
	fmt.Println(r.RequestURI)
	fmt.Fprintf(w, r.RequestURI)
	fmt.Fprintf(w, string(body))
}

func httpError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, err.Error())
	fmt.Println(err)
}

func newUUIDForBson() string {
	UUID := uuid.New()
	UUIDString := UUID.String()
	UUIDString = strings.Replace(UUIDString, "-", "", -1)
	UUIDString = UUIDString[:24]
	return UUIDString
}
