package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/farseeker/go-flamenco/flamenco-imports/websetup"
)

func linkExchange(w http.ResponseWriter, r *http.Request) {

	var payload websetup.KeyExchangeRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode payload: %s", err.Error()))
		return
	}

	thisLinkExchange := linkExchangeConversation{
		ManagerPayload: payload,
		ManagerKey:     payload.KeyHex,
	}

	thisConversation := managerDoc{
		ManagerID:    newUUIDForBson(),
		CreatedAt:    time.Now(),
		CreatedBy:    r.RemoteAddr,
		LinkExchange: thisLinkExchange,
	}

	conversationDoc := fsGetLink(thisConversation.ManagerID)
	wr, err := conversationDoc.Create(ctx, thisConversation)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to save link conversation %s", err.Error()))
		fmt.Println(wr)
		return
	}

	response := websetup.KeyExchangeResponse{
		Identifier: thisConversation.ManagerID,
	}

	responseJSON, _ := json.Marshal(&response)
	fmt.Fprintf(w, string(responseJSON))
}

func linkChoose(w http.ResponseWriter, r *http.Request) {
	mac := r.FormValue("hmac")
	identifier := r.FormValue("identifier")
	returnURL := r.FormValue("return")

	conversationDoc := fsGetLink(identifier)
	conversationSnapshot, err := conversationDoc.Get(ctx)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to fetch linker conversation for %s: %s", identifier, err.Error()))
		return
	}

	thisConversation := managerDoc{}
	conversationSnapshot.DataTo(&thisConversation)

	thisChooseConversation := chooseConversation{
		QueryHMAC:                 mac,
		QueryIdentifier:           identifier,
		QueryReturn:               returnURL,
		ResponseManagerUUIDString: newUUIDForBson(),
	}

	linkerKey, err := hex.DecodeString(thisConversation.LinkExchange.ManagerKey)
	if err != nil {
		httpError(w, fmt.Errorf("Unable to decode mac %s: %s", thisConversation.LinkExchange.ManagerKey, err.Error()))
		return
	}

	msg := []byte(identifier + "-" + thisChooseConversation.ResponseManagerUUIDString)
	hmac := hmac.New(sha256.New, linkerKey)
	hmac.Write(msg)

	thisChooseConversation.ResponseHMAC = hex.EncodeToString(hmac.Sum(nil))
	thisChooseConversation.ResponseURL = fmt.Sprintf("%s?hmac=%s&oid=%s", returnURL, thisChooseConversation.ResponseHMAC, thisChooseConversation.ResponseManagerUUIDString)

	conversationDoc.Update(ctx, []firestore.Update{firestore.Update{
		Path:  "_ChooseConversation",
		Value: thisChooseConversation,
	}})

	//Make a new doc with the managers final ID and copy the conversation doc into it
	thisConversation.ChooseConversation = thisChooseConversation
	managerDoc := fsGetManager(thisChooseConversation.ResponseManagerUUIDString)
	managerDoc.Set(ctx, thisConversation)

	http.Redirect(w, r, thisChooseConversation.ResponseURL, http.StatusTemporaryRedirect)
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
