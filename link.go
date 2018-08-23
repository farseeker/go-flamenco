package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/farseeker/go-flamenco/flamenco-imports/websetup"
	"github.com/google/uuid"
)

var linkerKey []byte

func linkExchange(w http.ResponseWriter, r *http.Request) {
	var payload websetup.KeyExchangeRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}
	fmt.Println("Key Exchange started: ", payload.KeyHex)

	key, err := hex.DecodeString(payload.KeyHex)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode mac %s: %s", payload.KeyHex, err.Error()))
		return
	}
	linkerKey = key

	tx, err := db.Begin(true)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to open database transaction: %s", err.Error()))
		return
	}
	defer tx.Rollback()

	knownAS := uuid.New()
	knownAsString := knownAS.String()

	bkt, err := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("linker-%s", knownAsString)))
	if err != nil {
		httpError(w, fmt.Errorf("Unable to create linker bucket: %s", err.Error()))
		return
	}

	time, _ := time.Now().GobEncode()
	bkt.Put([]byte("linker-key"), linkerKey)
	bkt.Put([]byte("created-at"), time)
	bkt.Put([]byte("created-by"), []byte(r.RemoteAddr))

	if err := tx.Commit(); err != nil {
		httpError(w, fmt.Errorf("Unable to commit linker bucket: %s", err.Error()))
		return
	}

	response := websetup.KeyExchangeResponse{
		Identifier: knownAsString,
	}

	fmt.Println("Known as:", knownAsString)

	responseJSON, _ := json.Marshal(&response)
	fmt.Fprintf(w, string(responseJSON))
}

func linkChoose(w http.ResponseWriter, r *http.Request) {
	mac := r.FormValue("hmac")
	identifier := r.FormValue("identifier")
	returnURL := r.FormValue("return")

	fmt.Println("Choose identifier:", identifier)
	fmt.Println("Choose hmac:", mac)

	tx, err := db.Begin(true)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to open database transaction: %s", err.Error()))
		return
	}
	defer tx.Rollback()

	bkt := tx.Bucket([]byte(fmt.Sprintf("linker-%s", identifier)))
	if bkt == nil {
		httpError(w, fmt.Errorf("Linker bucket %s does not exist", identifier))
		return
	}

	managerUUIDString := newUUIDForBson()
	bkt.Put([]byte("linked-manager-id"), []byte(managerUUIDString))

	mgrbkt, err := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("manager-%s", managerUUIDString)))
	if err != nil {
		httpError(w, fmt.Errorf("Unable to create manager bucket: %s", err.Error()))
		return
	}
	mgrbkt.Put([]byte("from-linker-id"), []byte("identifier"))

	if err := tx.Commit(); err != nil {
		httpError(w, fmt.Errorf("Unable to commit linker/manager bucket: %s", err.Error()))
		return
	}

	msg := []byte(identifier + "-" + managerUUIDString)
	hmac := hmac.New(sha256.New, linkerKey)
	hmac.Write(msg)

	computedMac := hex.EncodeToString(hmac.Sum(nil))

	redirectToString := fmt.Sprintf("%s?hmac=%s&oid=%s", returnURL, computedMac, managerUUIDString)

	http.Redirect(w, r, redirectToString, http.StatusTemporaryRedirect)
}

func linkReset(w http.ResponseWriter, r *http.Request) {
	var resetRequest websetup.AuthTokenResetRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&resetRequest)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}

	resetResponse := websetup.AuthTokenResetResponse{
		ExpireTime: "", //ignored
		Token:      "abcd",
	}
	resetJSON, _ := json.Marshal(resetResponse)
	w.Header().Set("Content-Type", jsonType)
	w.Write(resetJSON)
}
