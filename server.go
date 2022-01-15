package main

import (
    "os"
    "fmt"
    "log"
    "net/http"
)

func getHomeHandler(_ *gameState) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        fmt.Printf("serving request %+v", r)
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
    http.Handle("/", getHomeHandler(initGs))
    http.Handle("/move", getMoveHandler(solveMap))
    log.Fatal(http.ListenAndServe(":8080", nil))
}