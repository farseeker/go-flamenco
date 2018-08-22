package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
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

	http.HandleFunc("/api/flamenco/managers/link/exchange", linkExchange)
	http.HandleFunc("/flamenco/managers/link/choose", linkChoose)
	http.HandleFunc("/api/flamenco/managers/link/reset-token", linkReset)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		output := fmt.Sprintf("------\n%s\n%s\n", r.RequestURI, string(body))
		ioutil.WriteFile("log.txt", []byte(output), 0744)
		fmt.Fprintf(w, r.RequestURI)
		fmt.Fprintf(w, string(body))
		//fmt.Fprintf(w, "Welcome to my website!")
	})

	http.ListenAndServe(":8123", nil)
}
