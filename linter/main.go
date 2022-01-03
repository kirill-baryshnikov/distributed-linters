package main

import (
    "log"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

type LintingRequest struct {
    Content string `json:"content"`
}

func handleLint(responseWriter http.ResponseWriter, 
              request *http.Request){
    log.Println("Endpoint hit: /lint")

    requestBody, _ := ioutil.ReadAll(request.Body)
    var lintingRequest LintingRequest
    json.Unmarshal(requestBody, &lintingRequest)

    log.Println("Received content for linting:")
    log.Println(lintingRequest.Content)

    // TODO linting logic

    json.NewEncoder(responseWriter).Encode(lintingRequest)
}

func handleHealthy(responseWriter http.ResponseWriter, 
              request *http.Request){
    log.Println("Endpoint hit: /healthy")
}

func serve() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/lint", handleLint).Methods("POST")
    myRouter.HandleFunc("/healthy", handleHealthy).Methods("GET")
    log.Fatal(http.ListenAndServe(":8136", myRouter))
}

func main() {
    serve()
}
