package main

import "net/http"

func taskByID(w http.ResponseWriter, r *http.Request) {
	//If-None-Match is the header to match against the etag of the task

	w.WriteHeader(http.StatusNotModified)
	return
}
