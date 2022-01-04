package main

import (
    "log"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "os"
    "strconv"
    "fmt"
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

func handleLint(responseWriter http.ResponseWriter, request *http.Request, language Language) {
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

func handleHealthy(responseWriter http.ResponseWriter, request *http.Request) {
    log.Println("Endpoint hit: /healthy")
}

func serve(port int) {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/lint/java", handleLintJava).Methods("POST")
    myRouter.HandleFunc("/lint/python", handleLintPython).Methods("POST")
    myRouter.HandleFunc("/healthy", handleHealthy).Methods("GET")
    myRouter.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request){ os.Exit(0) })

    log.Println("Linter service listening.")
    log.Fatal(http.ListenAndServe(":" + strconv.Itoa(port), myRouter))
}

func main() {
    if len(os.Args) != 3 || os.Args[1] != "--port" {
        fmt.Println("Bad arguments - usage: ./linter --port <listen port>")
        os.Exit(1)
    }

    port, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Println("Bad arguments - usage: ./linter --port <listen port>")
        os.Exit(1)
    }

    serve(port)
}
