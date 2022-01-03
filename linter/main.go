package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)

func handleLint(responseWriter http.ResponseWriter, 
              request *http.Request){
    fmt.Fprintf(responseWriter, "Hi, I'm linting!")
    fmt.Println("Endpoint hit: /lint")
}

func handleHealthy(responseWriter http.ResponseWriter, 
              request *http.Request){
    fmt.Fprintf(responseWriter, "Hi, I'm healthy!")
    fmt.Println("Endpoint hit: /healthy")
}

func serve() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/lint", handleLint)
    myRouter.HandleFunc("/healthy", handleHealthy)
    log.Fatal(http.ListenAndServe(":8136", myRouter))
}

func main() {
    serve()
}