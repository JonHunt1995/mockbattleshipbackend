package main

type Game struct {
	p1   Player
	p2   Player
	turn int
}

func (g *Game) validateTurn(player_id string) error {
	if (g.turn%2 == 0 && g.p1.id == player_id) ||
		(g.turn%2 != 0 && g.p2.id == player_id) {
		return nil
	}
	return &malformedRequest{
		status: 400, // Bad Request
		msg:    "It is not your turn",
	}
}

func (g *Game) playTurn() {

}

func (g *Game) getLivingShips(p *Player) {

}
