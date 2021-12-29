package main


// var NUM_FINGERS int = 5
var NUM_FINGERS int8 = 3

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

type player struct {
    lh int8
    rh int8
}

func (p *player) equals(other *player) bool {
	return p.lh == other.lh && p.rh == other.rh
}

// If necessary swap lh and rh so the lh <= rh, 
// return True if we swapped and false if not
func (p *player) normalize() (*player, bool) {
	if p.lh > p.rh {
		p.rh, p.lh = p.lh, p.rh	
		return p, true
	}
	return p, false
}

func (p *player) isNormalized() bool {
	return p.lh <= p.rh
}

func (p player) copyAndNormalize() (*player, bool) {
	return p.normalize()
}

func (p *player) isEliminated() bool {
	return p.lh == 0 && p.rh == 0
}

func (p *player) getHand(h Hand) int8 {
	if h == Left {
		return p.lh
	} else {
		return p.rh
	}
}

func (p *player) setHand(h Hand, value int8) *player {
	if h == Left {
		p.lh = value
	} else {
		p.rh = value
	}
	return p
}
