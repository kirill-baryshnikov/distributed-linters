package main

import (
    "log"
    "bytes"
    "net/http"
    "sync"
    "math/rand"
    "errors"
    "time"
    "io/ioutil"
    "encoding/json"
    "os/exec"
    "strconv"
)

const (
    workerStateCreating = iota
    workerStateRunning
)

type Worker struct {
    address string
    version string
    state int
}

type Manager struct {
    mutex sync.Mutex

    workers []Worker
    target_workers_num int

    versions []string
    target_version string
    target_version_ratio float32
}

func NewManager(intial_version string) Manager {
    return Manager {
        mutex: sync.Mutex{},
        workers: make([]Worker, 0),
        target_workers_num: 4,
        versions: []string{intial_version},
        target_version: intial_version,
        target_version_ratio: 1.0,
    }
}

func (m *Manager) AddNewWorker() {
    m.target_workers_num += 1
}

func (m *Manager) RemoveWorker() {
    if m.target_workers_num > 1 {
        m.target_workers_num -= 1
    }
}

func (m* Manager) NewVersion(version string) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.versions = append(m.versions, version)
    m.target_version = version
    m.target_version_ratio = 0.0
}

func (m* Manager) RollbackVersion(rollbackTo string) {
    m.target_version = rollbackTo
    
    // versionUpdateStep will take care of removing bad workers
}

func (m* Manager) startPeriodicVersionUpdates() {
    go func() {
        for {
            m.versionUpdateStep()
            time.Sleep(5 * time.Second)
        }
    }()
}

func (m* Manager) versionUpdateStep() {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // First remove any workers that have newer version than desired
    // In case of rollback this will remove most workers, maybe all
    good_workers := make([]Worker, 0)

    for _, worker := range m.workers {
        for _, version := range m.versions {
            if version == m.target_version {
                // This worker's version is too new!
                // It should be removed.
                go shutdownWorker(worker)
                break;
            }

            if version == worker.version {
                // This worker has version older than target_version, it's ok
                good_workers = append(good_workers, worker)
                break;
            }
        }
    }

    m.workers = good_workers

    // Spawn new workers in place of the ones just deleted
    for len(m.workers) < m.target_workers_num {
        new_worker := Worker {
            address: "unitialized",
            version: m.target_version,
            state: workerStateCreating,
        }
        m.workers = append(m.workers, new_worker)
        go startupWorker(m, &new_worker)
    }

    // In case there are too many workers delete some
    for len(m.workers) > m.target_workers_num {
        m.removeWorker(0)
    }

    // Find out how much of the new workers should have m.target_version
    var new_ratio float32 = 0.0
    if m.target_version_ratio == 0 {
        new_ratio = 0.1
    } else {
        new_ratio = m.target_version_ratio * 2.0
    }

    if new_ratio > 1.0 {
        new_ratio = 1.0
    }

    min_cur_version_workers := int(new_ratio * float32(m.target_workers_num))

    current_version_workers := 0
    for _, worker := range m.workers {
        if worker.version == m.target_version {
            current_version_workers += 1
        }
    }

    // Update workers until the requirement is met
    for current_version_workers < min_cur_version_workers {
        for worker_index, worker := range m.workers {
            if worker.version != m.target_version {
                m.removeWorker(worker_index)
                break
            }
        }

        new_worker := Worker {
            address: "unitialized",
            version: m.target_version,
            state: workerStateCreating,
        }
        m.workers = append(m.workers, new_worker)

        go startupWorker(m, &new_worker)

        current_version_workers += 1;
    }
}

func (m* Manager) removeWorker(worker_index int) {
    last_index := len(m.workers) - 1

    m.workers[worker_index], m.workers[last_index] = m.workers[last_index], m.workers[worker_index]
    m.workers = m.workers[:last_index]
}

func startupWorker(m* Manager, worker* Worker) {
    randomPort := 10001 + rand.Intn(20000)

    binary_to_run := worker.version

    exec.Command(binary_to_run, "--port", strconv.Itoa(randomPort))

    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    worker.state = workerStateRunning
}

func shutdownWorker(worker Worker) {
    http.Get(worker.address + "/shutdown")
}

func (m* Manager) LintCode(code string, language string) (bool, error) {
    worker, err := m.chooseWorker()

    if err != nil {
        return false, err
    }

    to_lint := SourceFile {
        Content: code,
    }

    marshalled, _ := json.Marshal(to_lint)
    post_body := bytes.NewBuffer(marshalled)

    response, err := http.Post(worker.address + "/lint/" + language, "application/json", post_body)

    if err != nil {
        return false, err
    }

    defer response.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatalln(err)
    }

    var responseFile SourceFile

    err = json.Unmarshal(body, &responseFile)
    if err != nil{
        return false, err
    }

    if responseFile.Content == code {
        // Code hasn't changed - everything is alright
        return true, nil
    } else {
        return false, nil
    }
}

func (m* Manager) chooseWorker() (*Worker, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    possibleWorkers := make([]Worker, 0);

    for _, worker := range m.workers {
        if worker.state == workerStateRunning {
            possibleWorkers = append(possibleWorkers, worker)
        }
    }

    if len(possibleWorkers) == 0 {
        return nil, errors.New("No active worker found for this language")
    }

    randomIndex := rand.Intn(len(possibleWorkers))
    return &possibleWorkers[randomIndex], nil
}

type SourceFile struct {
    Content string `json:"content"`
}

const CONTENT_LENGTH_LIMIT = 60000

type LintResponse struct {
    result bool `json:"result"`
}

// /v1/lint
func handle_lint(w http.ResponseWriter, r *http.Request, m* Manager, language string) {
    requestBody, _ := ioutil.ReadAll(r.Body)
    var fileToLint SourceFile
    err := json.Unmarshal(requestBody, &fileToLint)
    if err != nil || fileToLint.Content == "" || len(fileToLint.Content) > CONTENT_LENGTH_LIMIT {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    is_good, err := m.LintCode(fileToLint.Content, language)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    response := LintResponse {
        result: is_good,
    }

    json.NewEncoder(w).Encode(response)
}

// /v1/admin/workers
func handle_admin_workers(w http.ResponseWriter, r *http.Request, m* Manager) {
    // POST - add new worker
    if r.Method == "POST" {
        m.AddNewWorker()
    }

    // DELETE - delete a worker
    if r.Method == "DELETE" {
        m.RemoveWorker();
    }
}

// /v1/admin/balance
func handle_admin_balance(w http.ResponseWriter, r *http.Request, m* Manager) {
    // Will be added in the future
}

type VersionJson struct {
    Version string `json:"version"`
}

// /v1/admin/version
func handle_admin_version(w http.ResponseWriter, r *http.Request, m* Manager) {
    requestBody, _ := ioutil.ReadAll(r.Body)
    var new_version VersionJson
    err := json.Unmarshal(requestBody, &new_version)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    m.NewVersion(new_version.Version)
}

// /v1/admin/rollback
func handle_admin_rollback(w http.ResponseWriter, r *http.Request, m* Manager) {
    requestBody, _ := ioutil.ReadAll(r.Body)
    var old_version VersionJson
    err := json.Unmarshal(requestBody, &old_version)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    m.RollbackVersion(old_version.Version)
}

func handleRequests(python_manager *Manager, java_manager *Manager) {
    http.HandleFunc("/v1/lint/python", func(w http.ResponseWriter, r *http.Request) {
        handle_lint(w, r, python_manager, "python")
    })

    http.HandleFunc("/v1/lint/java", func(w http.ResponseWriter, r *http.Request) {
        handle_lint(w, r, java_manager, "java")
    })

    http.HandleFunc("/v1/admin/workers/python", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_workers(w, r, python_manager)
    })

    http.HandleFunc("/v1/admin/workers/java", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_workers(w, r, java_manager)
    })

    http.HandleFunc("/v1/admin/balance/python", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_balance(w, r, python_manager)
    })

    http.HandleFunc("/v1/admin/balance/java", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_balance(w, r, java_manager)
    })

    http.HandleFunc("/v1/admin/version/python", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_version(w, r, python_manager)
    })

    http.HandleFunc("/v1/admin/version/java", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_version(w, r, python_manager)
    })

    http.HandleFunc("/v1/admin/rollback/python", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_rollback(w, r, python_manager)
    })

    http.HandleFunc("/v1/admin/rollback/java", func(w http.ResponseWriter, r *http.Request) {
        handle_admin_version(w, r, java_manager)
    })

    log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
    var python_manager Manager = NewManager("python-linter-1.0")
    var java_manager Manager = NewManager("java-linter-1.0")

    python_manager.startPeriodicVersionUpdates()
    java_manager.startPeriodicVersionUpdates()

    handleRequests(&python_manager, &java_manager)
}
