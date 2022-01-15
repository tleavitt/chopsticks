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

type uiGameState struct {
    p1 player
    p2 player
    turn string
}

func getAndCastMap(mapIf map[string]interface{}, key string) (map[string]interface{}, error) {
    if valIf, ok := mapIf[key]; ok {
        val, ok := val.(map[string]interface{})
        if !ok {
            return nil, fmt.Errorf("Value for key %s is not a map: %+v", key, val)
        }
        return val, nil
    } else {
        return nil, fmt.Errorf("Missing key: %s", key)
    }
}

func getAndCastInt8(mapIf map[string]interface{}, key string) (int8, error) {
    if valIf, ok := mapIf[key]; ok {
        val, ok := val.(int8)
        if !ok {
            return 0, fmt.Errorf("Value for key %s is not an int8: %+v", key, val)
        }
        return val, nil
    } else {
        return 0, fmt.Errorf("Missing key: %s", key)
    }
}

func interfaceToTurn(turnIf interface{}) (Turn, error) {
    turnStr, ok := turnIf.(string)
    if !ok {
        return Player1, fmt.Errorf("Turn is not a string %+v", turnIf)
    }
    if turnStr == "p1" {
        return Player1, nil
    } else if turnStr == "p2" {
        return Player2, nil
    } else {
        return Player2, fmt.Errorf("Unrecognized turn %s", turnStr)
    }
}

func interfaceToHand(handIf interface{}) (Hand, error) {
    handStr, ok := handIf.(string)
    if !ok {
        return Player1, fmt.Errorf("Hand is not a string %+v", handIf)
    }
    if handStr == "lh" {
        return Left, nil
    } else if handStr == "rh" {
        return Right, nil
    } else {
        return Left, fmt.Errorf("Unrecognized hand %s", handStr)
    }
}

func playerFromMap(playerMap map[string]interface{}) (*player, error) {
    p := player{1, 1}
    if lh, err := getAndCastInt8(playerMap, "lh"); err == nil {
        p.lh = lh;
    } else {
        return nil, err
    }
    if rh, err := getAndCastInt8(playerMap, "rh"); err == nil {
        p.rh = rh;
    } else {
        return nil, err
    }
    return &p, nil
}


// HOLY MOTHER OF FUCK THIS SUCKSSSSS
func parseUiMove(jsonBody []byte) (*gameState, *move, error) {
    var body map[string]interface{}
    if err := json.Unmarshal(jsonBody, &body); err != nil {
        return nil, nil, err
    }

    // Parse the gamestate
    gs := initGame()
    if gsMap, err := getAndCastMap(body, "gs"); err == nil {
        if player1Map, err := getAndCastMap(gsMap, "p1"); err == nil {
            if player1, err := playerFromMap(player1Map); err == nil {
                gs.player1 = player1
            } else {
                return nil, nil, err
            }
        } else {
            return nil, nil, err
        }
        if player2Map, err := getAndCastMap(gsMap, "p2"); err == nil {
            if player2, err := playerFromMap(player2Map); err == nil {
                gs.player2 = player2
            } else {
                return nil, nil, err
            }
        } else {
            return nil, nil, err
        }
        if turnIf, ok := gsMap["turn"]; ok {
            if turn, err := interfaceToTurn(turnIf); err == nil {
                gs.turn = turn
            } else {
                return nil, nil, err
            }
        } else {
            return nil, nil, errors.New("Missing key: turn")
        }
    } else {
        return nil, nil, err
    }

    // Parse the move
    var m *move = &move{Left, Left}
    if moveMap, err := getAndCastMap(body, "move"); err == nil {
        if playerHandIf, ok := moveMap["playerHand"]; ok {
            if playerHand, err := interfaceToHand(playerHandIf); err == nil {
                m.playerHand = playerHand
            } else {
                return nil, nil, err
            }
        } else {
            return, nil, nil, errors.New("Missing key: playerHand")
        }
        if receiverHandIf, ok := moveMap["receiverHand"]; ok {
            if receiverHand, err := interfaceToHand(receiverHandIf); err == nil {
                m.receiverHand = receiverHand
            } else {
                return nil, nil, err
            }
        } else {
            return, nil, nil, errors.New("Missing key: playerHand")
        }
    } else {
        return nil, nil, err
    }
    fmt.Printf("Got %+v, %+v\n", gs, m)
    return gs, m, nil
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
        // var uigs *uiGameState = nil
        // if err := json.Unmarshal(body, uigs); err != nil {
        //     fmt.Println(err.Error())
        //     panic(err)
        // }
        // fmt.Printf("Got %+v\n", uigs)
        // Send it back at them.
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