package main

import (
	"fmt"
)

var NUM_FINGERS int8 = 5
// var NUM_FINGERS int8 = 4

// TODO: don't make this a global variable
func setNumFingers(numFingers int8) int8 {
	prevNumFingers := NUM_FINGERS
	NUM_FINGERS = numFingers
	fmt.Printf("Setting NUM_FINGERS to %d (previously %d)\n", NUM_FINGERS, prevNumFingers)
	return prevNumFingers
}

type Hand int8

const (
	Left Hand = iota
	Right
)

func (h Hand) invert() Hand {
	if h == Left {
		return Right
	} else {
		return Left
	}
}

func toString(h Hand) string {
	if h == Left {
		return "LH"
	} else {
		return "RH"
	}
}

type Turn int8
const (
	Player1 Turn = 1
	Player2 = 2
)

type Player struct {
    Lh int8
    Rh int8
}

func (p *Player) equals(other *Player) bool {
	return p.Lh == other.Lh && p.Rh == other.Rh
}

// If necessary swap lh and rh so the lh <= rh, 
// return True if we swapped and false if not
func (p *Player) normalize() (*Player, bool) {
	if p.Lh > p.Rh {
		p.Rh, p.Lh = p.Lh, p.Rh	
		return p, true
	}
	return p, false
}

func (p *Player) isNormalized() bool {
	return p.Lh <= p.Rh
}

func (p Player) copyAndNormalize() (*Player, bool) {
	return p.normalize()
}

func (p *Player) isEliminated() bool {
	return p.Lh == 0 && p.Rh == 0
}

func (p *Player) getHand(h Hand) int8 {
	if h == Left {
		return p.Lh
	} else {
		return p.Rh
	}
}

func (p *Player) setHand(h Hand, value int8) *Player {
	if h == Left {
		p.Lh = value
	} else {
		p.Rh = value
	}
	return p
}


var DISTINCT_HANDS []Hand = []Hand{Left, Right}
var LEFT_HAND []Hand = []Hand{Left}
var RIGHT_HAND []Hand = []Hand{Right}

// Returns the distinct hands that a Player can use to play
// WLOG if the Player can use (i.e. play or receive) either hand, they always use their left hand.
func (p *Player) getDistinctPlayableHands() []Hand {
	if p.isEliminated() {
		return []Hand{} // Empty slice, no hands playable 
	} else if p.Lh == 0 {
		return RIGHT_HAND
	} else if p.Rh == 0 || p.Rh == p.Lh {
		return LEFT_HAND
	} else {
		return DISTINCT_HANDS
	}
}
