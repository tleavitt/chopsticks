package main

import (
  "fmt"
  "testing"
  "encoding/json"
)

func TestUnmarshalState(t *testing.T) {
  fmt.Println("starting TestUnmarshalState")
  j := `{"p1":{"lh":1,"rh":1},"p2":{"lh":1,"rh":1},"turn":"p1"}`
  _, err := parseUiMove([]byte(j))
  if err != nil {
    t.Fatal(err.Error())
  }
}

func TestMarshalState(t *testing.T) {
  fmt.Println("starting TestMarshalState")
  nextStateAndMove := &NextStateAndMove {
    *initGame(),
    Move{Left, Right,},
  }

  jsonResp, err := json.Marshal(nextStateAndMove) 
  if err != nil {
    t.Fatal(err.Error())
  }
  fmt.Printf("Serialized as %s", string(jsonResp))
}