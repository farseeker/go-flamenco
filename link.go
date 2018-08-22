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
		fmt.Fprintf(w, "Unable to decode payload: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Key Exchange started: ", payload.KeyHex)

	key, err := hex.DecodeString(payload.KeyHex)
	if err != nil {
		fmt.Fprintf(w, "Unable to decode mac %s: %s", payload.KeyHex, err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	linkerKey = key

	tx, err := db.Begin(true)
	if err != nil {
		fmt.Fprintf(w, "Unable to open database transaction: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	knownAS := uuid.New()
	knownAsString := knownAS.String()

	bkt, err := tx.CreateBucketIfNotExists([]byte(knownAsString))
	if err != nil {
		fmt.Fprintf(w, "Unable to create user bucket: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	time, _ := time.Now().GobEncode()
	bkt.Put([]byte("linker-key"), linkerKey)
	bkt.Put([]byte("created-at"), time)
	bkt.Put([]byte("created-by"), []byte(r.RemoteAddr))

	if err := tx.Commit(); err != nil {
		fmt.Fprintf(w, "Unable to commit user bucket: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
	//mac := r.FormValue("hmac")
	identifier := r.FormValue("identifier")
	returnURL := r.FormValue("return")

	tx, err := db.Begin(true)
	if err != nil {
		fmt.Fprintf(w, "Unable to open database transaction: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	bkt := tx.Bucket([]byte(identifier))
	if bkt == nil {
		fmt.Fprintf(w, "User bucket %s does not exist", identifier)
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//knownAsByte := bkt.Get([]byte("known-as"))
	//knownAsUUID := uuid.New()
	//knownAsUUID.UnmarshalText(string(knownAsByte))
	//knownAsString := knownAsUUID.String()

	projectUUID := uuid.New()
	projectUUIDByte, _ := projectUUID.MarshalBinary()
	projectUUIDString := projectUUID.String()

	bkt.Put([]byte("project-id"), projectUUIDByte)
	if err := tx.Commit(); err != nil {
		fmt.Fprintf(w, "Unable to commit user bucket: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg := []byte(identifier + "-" + projectUUIDString)
	hmac := hmac.New(sha256.New, linkerKey)
	hmac.Write(msg)

	computedMac := hex.EncodeToString(hmac.Sum(nil))

	redirectToString := fmt.Sprintf("%s?hmac=%s&oid=%s", returnURL, computedMac, projectUUIDString)

	http.Redirect(w, r, redirectToString, http.StatusTemporaryRedirect)
}

func linkReset(w http.ResponseWriter, r *http.Request) {
	//{"manager_id":"ef13a0f6-d12f-4b11-94d8-7c13bd78ef0b","identifier":"asdf","padding":"0b97ea50d31f6a726253ac548df7c5c46ceec1fc80aab37dbf73595d7a6b1ad1","hmac":"3dbbce08590ff4aae82ef55fa8cb279c53f7847ffe9ae01f3488f593121603e9"}

	var resetRequest websetup.AuthTokenResetRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&resetRequest)
	if err != nil {
		fmt.Fprintf(w, "Unable to decode payload: %s", err.Error())
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resetResponse := websetup.AuthTokenResetResponse{
		ExpireTime: "", //ignored
		Token:      "abcd",
	}
	resetJSON, _ := json.Marshal(resetResponse)
	w.Write(resetJSON)
}
