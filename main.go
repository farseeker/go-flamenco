package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var configPath = flag.String("config", "go-flamenco.toml", "Configuration file to read for go-flamenco settings")

func main() {
	flag.Parse()
	config = appConfig{}
	configFile, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Println("Unable to open configuration file", *configPath)
		os.Exit(1)
	}

	_, err = toml.Decode(string(configFile), &config)
	if err != nil {
		fmt.Println("Unable to parse configuration file", *configPath)
		os.Exit(1)
	}

	err = fsConnect(config.FirestoreProjectID, config.FirestoreAuthFile)
	if err != nil {
		fmt.Printf("Unable to connect to Firestore project %s with creds from %s", config.FirestoreProjectID, config.FirestoreAuthFile)
		os.Exit(1)
	}

	r := mux.NewRouter()

	r.HandleFunc("/flamenco/managers/link/choose", linkChoose)
	r.HandleFunc("/api/flamenco/managers/link/exchange", linkExchange)
	r.HandleFunc("/api/flamenco/managers/link/reset-token", linkReset)
	r.HandleFunc("/api/flamenco/managers/{identity}/task-update-batch", taskUpdateBatch)
	r.HandleFunc("/api/flamenco/managers/{identity}/startup", startup)
	r.HandleFunc("/api/flamenco/managers/{identity}/depsgraph", depsgraph)
	r.HandleFunc("/api/flamenco/tasks/{taskid}", taskByID)

	r.PathPrefix("/").HandlerFunc(logDefault)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8123",
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
	fmt.Fprintln(w, err.Error())
	fmt.Println(err)
}

func newUUIDForBson() string {
	UUID := uuid.New()
	UUIDString := UUID.String()
	UUIDString = strings.Replace(UUIDString, "-", "", -1)
	UUIDString = UUIDString[:24]
	return UUIDString
}
