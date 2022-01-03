package main

import (
    "log"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

type SourceFile struct {
    Content string `json:"content"`
}

func handleLint(responseWriter http.ResponseWriter, 
              request *http.Request){
    log.Println("Endpoint hit: /lint")

    requestBody, _ := ioutil.ReadAll(request.Body)
    var fileToLint SourceFile
    json.Unmarshal(requestBody, &fileToLint)
    log.Println("Received content for linting:" + fileToLint.Content)
 
    lintedFile := SourceFile{lintSourceCode(fileToLint.Content)}
    log.Println("Content after linting" + lintedFile.Content)

    json.NewEncoder(responseWriter).Encode(lintedFile)
}

func handleHealthy(responseWriter http.ResponseWriter, 
              request *http.Request){
    log.Println("Endpoint hit: /healthy")
}

func serve() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/lint", handleLint).Methods("POST")
    myRouter.HandleFunc("/healthy", handleHealthy).Methods("GET")

    log.Println("Server starting on port 8136.")
    log.Fatal(http.ListenAndServe(":8136", myRouter))
}

func main() {
    serve()
}
