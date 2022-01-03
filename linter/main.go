package main

import (
    "fmt"
    "log"
    "net/http"
)

func homePage(responseWriter http.ResponseWriter, 
              request *http.Request){
    fmt.Fprintf(responseWriter, "Hello world!")
    fmt.Println("Endpoint hit: homePage")
}

func handleRequests() {
    http.HandleFunc("/", homePage)
    log.Fatal(http.ListenAndServe(":8136", nil))
}

func main() {
    handleRequests()
}
