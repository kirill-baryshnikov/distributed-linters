package main

import (
    "log"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

const CONTENT_LENGTH_LIMIT = 60000

func handleLintJava(responseWriter http.ResponseWriter, request *http.Request) {
    log.Println("Endpoint hit: /lint/java")
    handleLint(responseWriter, request, Java)
}

func handleLintPython(responseWriter http.ResponseWriter, request *http.Request) {
    log.Println("Endpoint hit: /lint/python")
    handleLint(responseWriter, request, Python)
}

func handleLint(responseWriter http.ResponseWriter, request *http.Request, 
				language Language) {
	requestBody, _ := ioutil.ReadAll(request.Body)
    var fileToLint SourceFile
    err := json.Unmarshal(requestBody, &fileToLint)
    if err != nil || fileToLint.Content == "" || len(fileToLint.Content) > CONTENT_LENGTH_LIMIT {
    	responseWriter.WriteHeader(http.StatusBadRequest)
    	return
    }

    log.Println("Received content for linting:\n" + fileToLint.Content)
 
    lintedFile := SourceFile { lintSourceCode(fileToLint.Content, language) }
    log.Println("Content after linting:\n" + lintedFile.Content)

    json.NewEncoder(responseWriter).Encode(lintedFile)
}

func handleHealthy(responseWriter http.ResponseWriter, 
              request *http.Request) {
    log.Println("Endpoint hit: /healthy")
}

func serve() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/lint/java", handleLintJava).Methods("POST")
	myRouter.HandleFunc("/lint/python", handleLintPython).Methods("POST")
    myRouter.HandleFunc("/healthy", handleHealthy).Methods("GET")

    log.Println("Linter service listening on port 8136.")
    log.Fatal(http.ListenAndServe(":8136", myRouter))
}

func main() {
    serve()
}
