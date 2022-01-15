package main

import (
  "fmt"
  "testing"
)

func TestUnmarshalMove(t *testing.T) {
  fmt.Println("starting TestStateCopy")
  j := `{"p1":{"lh":1,"rh":1},"p2":{"lh":1,"rh":1},"turn":"p1"}`
  _, err := parseUiMove([]byte(j))
  if err != nil {
    t.Fatal(err.Error())
  }
}