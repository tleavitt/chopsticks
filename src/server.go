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

type NextStateAndMove struct {
    NextState GameState
    M Move
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

func getHomeHandler(_ *GameState) http.Handler {
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

func getMoveHandler(solveMap map[GameState]*PlayNode) http.Handler {
    fn := func (w http.ResponseWriter, r *http.Request) {
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            log.Printf("Error reading body: %v", err)
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        gs, err := parseUiMove(body)
        if err != nil {
            fmt.Printf("ERROR PARSING JSON: " + err.Error())
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
        // Get the best Move for the current node:
        normalizedComputerMove, _, err := curNode.getBestMoveAndScoreForCurrentPlayer(DEBUG, true) // TODO: don't allow unscored child?
        if err != nil {
            log.Printf("Error finding best Move for %s", curNode.toString())
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        // And translate it into the Move the Player would see
        guiComputerMove, err := gps.playNormalizedTurn(normalizedComputerMove)

        next := &NextStateAndMove{
            *gps.state,
            guiComputerMove,
        }

        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        jsonResp, err := json.Marshal(next) 
        if err != nil {
            log.Printf("Error parsing json %s", err)
            http.Error(w, "can't read body", http.StatusBadRequest)
            return
        }
        fmt.Printf("Serialized as %s", string(jsonResp))

        w.Write(jsonResp)
        // Send them our Move and the new state.
    }
    return http.HandlerFunc(fn)
}

func serve(initGs *GameState, solveMap map[GameState]*PlayNode) {
    r := mux.NewRouter()
    r.Handle("/", getHomeHandler(initGs))
    r.Handle("/static/hands.png", getImageRequestHandler("./frontend/static/hands.png"))
    r.Handle("/static/hands_green.png", getImageRequestHandler("./frontend/static/hands_green.png"))
    r.Handle("/static/hands_red.png", getImageRequestHandler("./frontend/static/hands_red.png"))
    r.Handle("/move", getMoveHandler(solveMap))
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":8888", nil))
}