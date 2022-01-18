package main

import (
    "context"
    "errors"
    "encoding/json"
    "fmt"
    "github.com/gorilla/mux"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "sync"
    "syscall"
    "time"
)

const CONTENT_LENGTH_LIMIT = 60000

type shutdownFunc func()

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

func handleHealthz(responseWriter http.ResponseWriter, request *http.Request) {
    log.Println("Endpoint hit: /healthz")
}

func newServer(sf shutdownFunc, port int) *http.Server {
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/lint/java", handleLintJava).Methods("POST")
    router.HandleFunc("/lint/python", handleLintPython).Methods("POST")
    router.HandleFunc("/healthz", handleHealthz).Methods("GET")
    router.HandleFunc("/shutdown", func(_ http.ResponseWriter, _ *http.Request) { sf() })

    return &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: router,
    }
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

    stopCh := make(chan struct{})
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go func() {
        <-stopCh
        cancel()
    }()

    c := make(chan os.Signal, 2)
    signal.Notify(c, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)
    go func() {
        s := <-c
        log.Printf("Received first shutdown signal: %s. Shutting down gracefully.", s)
        close(stopCh)
        <-c
        log.Printf("Received second shutdown signal: %s. Exiting.", s)
        os.Exit(1)
    }()

    var wg sync.WaitGroup
    defer wg.Wait()

    sf := func() {
        close(stopCh)
    }
    server := newServer(sf, port)

    wg.Add(1)
    go func() {
        defer wg.Done()

        log.Printf("Starting Linter server on: %s", server.Addr)

        err = server.ListenAndServe()
        if err != nil && !errors.Is(err, http.ErrServerClosed) {
            log.Fatalf("Couldn't listen on %s: %v", server.Addr, err)
        }
    }()

    wg.Add(1)
    go func() {
        defer wg.Done()
        <-ctx.Done()

        log.Println("Shutting down Linter server")
        defer log.Println("Linter server shut down")

        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer shutdownCancel()
        err := server.Shutdown(shutdownCtx)
        if err != nil {
            log.Fatalf("Couldn't terminate gracefully: %v", err)
        }
    }()
}
