package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type MyEvent struct {
	Name string `json:"name"`
}

type MyResponse struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, event *MyEvent) (*MyResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}
	message := fmt.Sprintf("Hello %s!", event.Name)
	return &MyResponse{Message: message}, nil
}

func azureHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("This HTTP triggered function executed successfully. Pass a name in the query string for a personalized response.\n")
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error on reading request body: %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var event MyEvent
	err = json.Unmarshal(reqData, &event)
	if err != nil {
		fmt.Printf("error unmarshalling request body: %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	response, err := HandleRequest(r.Context(), &event)
	if err != nil {
		fmt.Printf("error handling request: %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("error marshalling response: %v\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/HttpExample", azureHandler)
	http.ListenAndServe(listenAddr, nil)
}
