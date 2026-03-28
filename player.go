package main

type Player struct {
	ships   ShipCoordinates
	id      string
	guesses []int
}

type Game struct {
	p1   Player
	p2   Player
	turn int
}

func (g *Game) validateTurn() {

}

func (g *Game) playTurn() {

}

func (g *Game) getLivingShips(p *Player) {

}
