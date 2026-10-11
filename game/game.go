package game

import (
	"errors"
	"sync"

	"github.com/JonHunt1995/mockbattleshipbackend/player"
	"github.com/JonHunt1995/mockbattleshipbackend/ship"
)

var (
	ErrPlayerAlreadyInGame = errors.New("Player is already in this game")
	ErrGameFull            = errors.New("This game already has 2 players")
	ErrPlayerNotFound      = errors.New("Player not found")
	ErrNoOpponent          = errors.New("No opponent found")
	ErrNotYourTurn         = errors.New("It is not your turn")
)

type Game struct {
	Players []*player.Player
	Turn    int
	Mu      sync.Mutex
}

func NewGame(players []*player.Player) *Game {
	return &Game{
		Players: players,
		Turn:    1,
		Mu:      sync.Mutex{},
	}
}

func (g *Game) GetPlayerCount() int {
	return len(g.Players)
}

func (g *Game) AddPlayer(p *player.Player) error {
	for _, existing := range g.Players {
		if existing.Id == p.Id {
			return ErrPlayerAlreadyInGame
		}
	}

	if g.GetPlayerCount() >= 2 {
		return ErrGameFull
	}
	g.Players = append(g.Players, p)

	return nil
}

type ClonedGame struct {
	Players []*player.Player `json:"Players"`
	Turn    int              `json:"Turn"`
}

func (g *Game) Clone() *ClonedGame {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	playersCopy := make([]*player.Player, len(g.Players))
	for i, p := range g.Players {
		guessesCopy := make([]int, len(p.Guesses))
		copy(guessesCopy, p.Guesses)
		playersCopy[i] = &player.Player{
			Ships:   p.Ships,
			Id:      p.Id,
			Guesses: guessesCopy,
		}
	}

	return &ClonedGame{
		Players: playersCopy,
		Turn:    g.Turn,
	}
}

func (g *Game) GetPlayerNumber(playerID string) (int, bool) {
	for i, p := range g.Players {
		if p.Id == playerID {
			return i + 1, true
		}
	}

	return 0, false
}

func (g *Game) GetPlayer(playerID string) (*player.Player, error) {
	idx, found := g.GetPlayerNumber(playerID)

	if !found {
		return nil, ErrPlayerNotFound
	}

	idx -= 1

	return g.Players[idx], nil
}

func (g *Game) GetOpponent(playerID string) (*player.Player, error) {
	other, found := g.GetPlayerNumber(playerID)

	if !found {
		return nil, ErrPlayerNotFound
	}

	other %= 2

	if len(g.Players) <= 1 || g.Players[other].Id == playerID {
		return nil, ErrNoOpponent
	}

	return g.Players[other], nil
}

func (g *Game) ValidateTurn(playerID string) error {
	playersTurn := (g.Turn + 1) % 2
	if g.Players[playersTurn].Id == playerID {
		return nil
	}
	return ErrNotYourTurn
}

func (g *Game) PlayTurn(playerID string, guess int) error {
	if err := g.ValidateTurn(playerID); err != nil {
		return err
	}

	g.Turn++

	if err := g.Players[g.Turn%2].AddGuess(guess); err != nil {
		g.Turn--
		return err
	}

	return nil
}

type GameStateResponse struct {
	PlayerShips         []int
	PlayerHits          []int
	PlayerMisses        []int
	PlayerLivingShips   ship.LivingShips
	OpponentHits        []int
	OpponentMisses      []int
	OpponentLivingShips ship.LivingShips
	IsYourTurn          bool
	TurnNumber          int
	Victor              int
	GameIsReady         bool
}

func CalculateVictoryScore(playerLiving, opponentLiving ship.LivingShips) int {
	if !playerLiving.Battleship && !playerLiving.Carrier && !playerLiving.Cruiser && !playerLiving.Submarine && !playerLiving.Destroyer {
		return -1
	} else if !opponentLiving.Battleship && !opponentLiving.Carrier && !opponentLiving.Cruiser && !opponentLiving.Submarine && !opponentLiving.Destroyer {
		return 1
	}

	return 0
}

func (g *Game) GetGameState(playerID string) (*GameStateResponse, error) {
	player, err := g.GetPlayer(playerID)
	if err != nil {
		return &GameStateResponse{}, err
	}
	playerShips := player.Ships.GetFlattenedCoords()
	var playerHits []int
	var playerMisses []int
	playerLivingShips := ship.NewLivingShips()
	var opponentHits []int
	var opponentMisses []int
	var isYourTurn bool
	opponentLivingShips := ship.NewLivingShips()
	playerNumber, fullGame := g.GetPlayerNumber(playerID)

	if fullGame && playerNumber-1 == (g.Turn-1)%2 {
		isYourTurn = true
	} else {
		isYourTurn = false
	}

	opp, err := g.GetOpponent(playerID)
	if err == nil {
		opponentHits, opponentMisses = opp.GetHitsAndMisses(player)
		playerHits, playerMisses = player.GetHitsAndMisses(opp)
		playerLivingShips = player.GetLivingShips(opp)
		opponentLivingShips = opp.GetLivingShips(player)
	}

	victoryScore := CalculateVictoryScore(playerLivingShips, opponentLivingShips)

	return &GameStateResponse{
		PlayerShips:         playerShips,
		PlayerHits:          playerHits,
		PlayerMisses:        playerMisses,
		PlayerLivingShips:   playerLivingShips,
		OpponentHits:        opponentHits,
		OpponentMisses:      opponentMisses,
		OpponentLivingShips: opponentLivingShips,
		IsYourTurn:          isYourTurn,
		TurnNumber:          g.Turn,
		Victor:              victoryScore,
		GameIsReady:         len(g.Players) == 2,
	}, nil
}
