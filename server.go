package main

import (
    "os"
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "github.com/gorilla/mux"
    "encoding/json"
)

type nextStateAndMove struct {
    nextState gameState
    m move
}

func getImageRequestHandler(path string) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        fmt.Printf("serving image for %s\n", r.URL)
        fileBytes, err := ioutil.ReadFile(path)
        if err != nil {
            panic(err)
        }
        fmt.Printf("num fyle bytes: %d\n", len(fileBytes))
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
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("Error reading body: %v", err)
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        gs, err := parseUiMove(body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        fmt.Printf("Got %+v\n", gs)

        // Normalize the game state:
        gps := createGamePlayState(gs)
        // Look up the state in our solve map
        curNode, exists := solveMap[*gps.normalizedState]
        if !exists {
            log.Printf("Did not find game state in solve map: %s", gps.toString())
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        // Get the best move for the current node:
        normalizedComputerMove, _, err := curNode.getBestMoveAndScoreForCurrentPlayer(DEBUG, true) // TODO: don't allow unscored child?
        if err != nil {
            log.Printf("Error finding best move for %s", curNode.toString())
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        // And translate it into the move the player would see
        guiComputerMove, err := gps.playNormalizedTurn(normalizedComputerMove)

        next := nextStateAndMove{
            *gps.state,
            guiComputerMove,
        }

        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        jsonResp, _ := json.Marshal(next) 
        w.Write(jsonResp)
        // Send them our move and the new state.
        fmt.Fprintf(w, "%s", body)
    }
    return http.HandlerFunc(fn)
}

func serve(initGs *gameState, solveMap map[gameState]*playNode) {
    r := mux.NewRouter()
    r.Handle("/", getHomeHandler(initGs))
    r.Handle("/static/hands.png", getImageRequestHandler("./frontend/static/hands.png"))
    r.Handle("/static/hands_green.png", getImageRequestHandler("./frontend/static/hands_green.png"))
    r.Handle("/static/hands_red.png", getImageRequestHandler("./frontend/static/hands_red.png"))
    r.Handle("/move", getMoveHandler(solveMap))
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8080", nil))
}