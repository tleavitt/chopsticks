package main

import (
    "os"
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "github.com/gorilla/mux"
)

func getImageRequestHandler(path string) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        fmt.Printf("serving image", r)
        fileBytes, err := ioutil.ReadFile(path)
        if err != nil {
            panic(err)
        }
        fmt.Printf("num fyle bytes: %d", len(fileBytes))
        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/octet-stream")
        w.Write(fileBytes)
    }
    return http.HandlerFunc(fn)
}

func getHomeHandler(_ *gameState) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        fmt.Printf("serving request %+v\n", r)
        body, err := os.ReadFile("frontend/index.html")
        if err != nil {
            log.Fatalf("Error when serving index.html, Err: %s", err)
        }
        fmt.Fprintf(w, "%s", body)
    }
    return http.HandlerFunc(fn)
}

func getMoveHandler(solveMap map[gameState]*playNode) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    }
    return http.HandlerFunc(fn)
}

func serve(initGs *gameState, solveMap map[gameState]*playNode) {
    r := mux.NewRouter()
    r.Handle("/", getHomeHandler(initGs))
    r.Handle("/static/hands.png", getImageRequestHandler("./frontend/static/hands.png"))
    r.Handle("/move", getMoveHandler(solveMap))
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8080", nil))
}