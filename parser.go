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

func interfaceToHand(handIf interface{}) (Hand, error) {
    handStr, ok := handIf.(string)
    if !ok {
        return Left, fmt.Errorf("Hand is not a string %+v", handIf)
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
    if lh, err := getAndCastInt(playerMap, "lh"); err == nil {
        p.lh = int8(lh);
    } else {
        return nil, err
    }
    if rh, err := getAndCastInt(playerMap, "rh"); err == nil {
        p.rh = int8(rh);
    } else {
        return nil, err
    }
    return &p, nil
}


// HOLY MOTHER OF FUCK THIS SUCKSSSSS
// 150 lines of code just to parse some fucking json
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
                gs.player1 = *player1
            } else {
                return nil, nil, err
            }
        } else {
            return nil, nil, err
        }
        if player2Map, err := getAndCastMap(gsMap, "p2"); err == nil {
            if player2, err := playerFromMap(player2Map); err == nil {
                gs.player2 = *player2
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
            return nil, nil, errors.New("Missing key: playerHand")
        }
        if receiverHandIf, ok := moveMap["receiverHand"]; ok {
            if receiverHand, err := interfaceToHand(receiverHandIf); err == nil {
                m.receiverHand = receiverHand
            } else {
                return nil, nil, err
            }
        } else {
            return nil, nil, errors.New("Missing key: playerHand")
        }
    } else {
        return nil, nil, err
    }
    fmt.Printf("Got %+v, %+v\n", gs, m)
    return gs, m, nil
}