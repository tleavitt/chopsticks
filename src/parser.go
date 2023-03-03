package main

import (
    "fmt"
    "errors"
    "encoding/json"
)

func getAndCastMap(mapIf map[string]interface{}, key string) (map[string]interface{}, error) {
    if valIf, ok := mapIf[key]; ok {
        val, ok := valIf.(map[string]interface{})
        if !ok {
            return nil, fmt.Errorf("Value for key %s is not a map: %+v", key, val)
        }
        return val, nil
    } else {
        return nil, fmt.Errorf("Missing key: %s", key)
    }
}

func getAndCastInt(mapIf map[string]interface{}, key string) (int, error) {
    if valIf, ok := mapIf[key]; ok {
        valF, ok := valIf.(float64)
        if !ok {
            return 0, fmt.Errorf("Value for key %s is not a float: %+v", key, valIf)
        }
        valI := int(valF)
        return valI, nil
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

func PlayerFromMap(playerMap map[string]interface{}) (*Player, error) {
    p := Player{1, 1}
    if lh, err := getAndCastInt(playerMap, "lh"); err == nil {
        p.Lh = int8(lh);
    } else {
        return nil, err
    }
    if rh, err := getAndCastInt(playerMap, "rh"); err == nil {
        p.Rh = int8(rh);
    } else {
        return nil, err
    }
    return &p, nil
}


// HOLY MOTHER OF FUCK THIS SUCKSSSSS
func parseUiMove(jsonBody []byte) (*GameState, error) {
    var body map[string]interface{}
    if err := json.Unmarshal(jsonBody, &body); err != nil {
        return nil, err
    }

    // Parse the gamestate
    gs := initGame()
    if player1Map, err := getAndCastMap(body, "p1"); err == nil {
        if player1, err := PlayerFromMap(player1Map); err == nil {
            gs.Player1 = *player1
        } else {
            return nil, err
        }
    } else {
        return nil, err
    }
    if player2Map, err := getAndCastMap(body, "p2"); err == nil {
        if player2, err := PlayerFromMap(player2Map); err == nil {
            gs.Player2 = *player2
        } else {
            return nil, err
        }
    } else {
        return nil, err
    }
    if turnIf, ok := body["turn"]; ok {
        if turn, err := interfaceToTurn(turnIf); err == nil {
            gs.T = turn
        } else {
            return nil, err
        }
    } else {
        return nil, errors.New("Missing key: turn")
    }

    fmt.Printf("Got %+v\n", gs)
    return gs, nil
}