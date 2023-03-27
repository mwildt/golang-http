package httputils

import (
	"encoding/json"
	"log"
	"net/http"
)

func SendJsonResponse[T any](w http.ResponseWriter, responsePayload T) {
	if data, err := json.Marshal(responsePayload); err != nil {
		log.Printf("error marshaling respons data as JSON: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
