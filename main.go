package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"math/rand"
	"os"
	"time"
)

type Obj [2]int

func (a *Obj) Plus(b Obj) Obj {
	return Obj{a[0] + b[0], a[1] + b[1]}
}

type Game struct {
	snake      []Obj
	food       Obj
	vector     Obj
	score      int
	gameOver   bool
	gameWidth  int
	gameHeight int
}

func (game *Game) nextStep(s tcell.Screen) {
	head := game.snake[len(game.snake)-1]

	for i := 0; i < len(game.snake)-1; i++ {
		game.snake[i] = game.snake[i+1]
	}

	next := head.Plus(game.vector)
	if game.crashGame(next) > 0 {
		game.gameOver = true
	}
	game.snake[len(game.snake)-1] = next

	if game.food[0] == head[0] && game.food[1] == head[1] {
		game.food = randomXY(game, s)
		game.snake = append(game.snake, game.snake[len(game.snake)-1])
	}

	if game.snake[len(game.snake)-1][1] > game.gameHeight {
		game.snake[len(game.snake)-1][1] = 0
	}
	if game.snake[len(game.snake)-1][1] < 0 {
		game.snake[len(game.snake)-1][1] = game.gameHeight
	}
	if game.snake[len(game.snake)-1][0] > game.gameWidth {
		game.snake[len(game.snake)-1][0] = 0
	}
	if game.snake[len(game.snake)-1][0] < 0 {
		game.snake[len(game.snake)-1][0] = game.gameWidth
	}

}

func (game *Game) showSnake(c tcell.Screen, style tcell.Style) {
	for i := 0; i < len(game.snake); i++ {
		c.SetContent(game.snake[i][0], game.snake[i][1], ' ', nil, style)
		//c.SetContent(game.snake[i][0]+1, game.snake[i][1], ' ', nil, style)
	}
}

func (game *Game) move(s tcell.Screen, e *tcell.EventKey) {

	if e.Key() == tcell.KeyEscape || e.Key() == tcell.KeyCtrlC {
		s.Fini()
		os.Exit(0)
	} else if e.Key() == tcell.KeyRight {
		if game.vector != (Obj{-1, 0}) {
			game.vector = Obj{1, 0}
		}
	} else if e.Key() == tcell.KeyLeft {
		if game.vector != (Obj{1, 0}) {
			game.vector = Obj{-1, 0}
		}

	} else if e.Key() == tcell.KeyDown {
		if game.vector != (Obj{0, -1}) {
			game.vector = Obj{0, 1}
		}

	} else if e.Key() == tcell.KeyUp {
		if game.vector != (Obj{0, 1}) {
			game.vector = Obj{0, -1}
		}
	}
}

func (game *Game) createFood(c tcell.Screen, style tcell.Style) {
	c.SetContent(game.food[0], game.food[1], ' ', nil, style)
	//c.SetContent(game.food[0]+1, game.food[1], ' ', nil, style)
}

func (game *Game) gameSpeed() time.Duration {
	return time.Second * 3 / time.Duration(50+game.score/3)
}

func (game *Game) crashGame(vector Obj) int {
	count := 0
	for i := 0; i < len(game.snake); i++ {
		if game.snake[i] == vector {
			count++
		}
	}
	return count
}

func (game *Game) getOffset(s tcell.Screen) (x1, x2, y1, y2 int) {
	termW, termH := s.Size()
	x1 = termW/2 - game.gameWidth/2
	x2 = termW/2 + game.gameWidth/2
	y1 = termH/2 - game.gameHeight/2
	y2 = termH/2 + game.gameHeight/2

	return x1, x2, y1, y2
}

func randomXY(game *Game, s tcell.Screen) Obj {
	x1, x2, y1, y2 := game.getOffset(s)

	x := rand.Intn((x2 - x1) + x1)
	y := rand.Intn((y2 - y1) + y1)
	return Obj{x, y}
}

func randomFood() Obj {
	a := rand.Intn(50)
	b := rand.Intn(20)
	return Obj{a, b}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	snakeStyle := tcell.StyleDefault.Background(tcell.ColorGreen)
	foodStyle := tcell.StyleDefault.Background(tcell.ColorRed)
	spaceStyle := tcell.StyleDefault.Background(tcell.ColorWhite)

	s, err := tcell.NewScreen()
	game := startGame(s)
	errorP(err)
	err = s.Init()
	errorP(err)
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.Clear()

	quit := func() {
		s.Fini()
		os.Exit(0)
	}

	ticker := time.NewTicker(game.gameSpeed())

	events := make(chan tcell.Event)
	go func() {
		for {
			events <- s.PollEvent()
		}
	}()

	for {

		if game.gameOver {
			quit()
		}
		//s.Clear()
		game.drawSpace(s, spaceStyle, defStyle)
		game.showSnake(s, snakeStyle)
		game.createFood(s, foodStyle)

		//game.nextStep()
		s.Show()
		select {
		case <-ticker.C:
			game.nextStep(s)
			ticker.Reset(game.gameSpeed())

		case ev := <-events:
			switch e := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				game.move(s, e)
			}
		}
	}

}

func startGame(s tcell.Screen) *Game {
	snake := make([]Obj, 4)
	food := Obj{}
	vector := Obj{0, 1}
	score := 0
	for i := 0; i < len(snake); i++ {
		snake[i] = Obj{0, i}
	}

	ret := &Game{
		snake:      snake,
		food:       food,
		vector:     vector,
		score:      score,
		gameOver:   false,
		gameWidth:  50,
		gameHeight: 20,
	}

	ret.food = randomXY(ret, s)

	return ret
}

func (game *Game) drawSpace(s tcell.Screen, background, line tcell.Style) {

	x1, x2, y1, y2 := game.getOffset(s)

	for row := 0; row <= game.gameHeight; row++ {
		for col := 0; col <= game.gameWidth; col++ {
			s.SetContent(x1+col, y1+row, ' ', nil, background)
		}
	}

	for row := 1; row < game.gameHeight; row++ {
		s.SetContent(x1, y1+row, tcell.RuneVLine, nil, line)
		s.SetContent(x2, y1+row, tcell.RuneVLine, nil, line)
	}

	for col := 1; col < game.gameWidth; col++ {
		s.SetContent(x1+col, y1, tcell.RuneHLine, nil, line)
		s.SetContent(x2-col, y2, tcell.RuneHLine, nil, line)
	}
	s.SetContent(x1, y1, tcell.RuneULCorner, nil, line)
	s.SetContent(x1, y2, tcell.RuneLLCorner, nil, line)
	s.SetContent(x2, y1, tcell.RuneURCorner, nil, line)
	s.SetContent(x2, y2, tcell.RuneLRCorner, nil, line)
}

func errorP(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

//package main
//
//import (
//	"fmt"
//	"github.com/gdamore/tcell/v2"
//	"math/rand"
//	"time"
//)
//
//var (
//	styleBG       = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
//	styleBold     = styleBG.Bold(true)
//	styleSnake    = styleBG.Background(tcell.ColorLightGreen)
//	styleFood     = styleBG.Background(tcell.ColorTeal)
//	styleGameOver = styleBold.Background(tcell.ColorGray).Foreground(tcell.ColorWhite)
//	gw            = 13 // game width
//	gh            = 13 // game height
//)
//
//type Vec [2]int
//
//func (a Vec) Plus(b Vec) Vec {
//	return Vec{a[0] + b[0], a[1] + b[1]}
//}
//
//// SnakeGame stores the state of the board and current score.
//// Implements the rules. Not responsible for user input or output.
//type SnakeGame struct {
//	width, height int
//	snake         []Vec
//	food          []Vec
//	vel           Vec // snake velocity: up, down, left, or right
//	isOver        bool
//	score         int
//}
//
//func CreateSnakeGame(width, height int) *SnakeGame {
//	// snake tail starts at 0,0, with the snake moving down
//	snake := make([]Vec, 4)
//	for i := 0; i < len(snake); i++ {
//		snake[i] = Vec{0, i}
//	}
//	vel := Vec{0, 1}
//
//	// place random food pieces, not overlapping each other or the snake
//	food := make([]Vec, 2)
//
//	ret := &SnakeGame{
//		width:  width,
//		height: height,
//		snake:  snake,
//		food:   food,
//		vel:    vel,
//	}
//
//	for i := range food {
//		food[i] = ret.placeRandomFood()
//	}
//
//	return ret
//}
//
//// SetVel sets the new snake velocity, in squares per tick.
//// The only allowed directions are up, down (0, 1), left, and right (1, 0).
//func (game *SnakeGame) SetVel(vel Vec) {
//	if vel.Plus(game.vel) == (Vec{0, 0}) {
//		return // ignore direction reversal
//	}
//	game.vel = vel
//}
//
//// Step advances the game by one step. Moves the snake, etc.
//func (game *SnakeGame) Step() {
//	if game.isOver {
//		return
//	}
//
//	// find the new head of the snake
//	tip := game.snake[len(game.snake)-1]
//	nt := tip.Plus(game.vel)
//	for i := 0; i < len(game.snake)-1; i++ {
//		game.snake[i] = game.snake[i+1]
//	}
//	game.snake[len(game.snake)-1] = nt
//	//// bounds check
//	//if nt[0] < 0 || nt[0] >= game.width || nt[1] < 0 || nt[1] >= game.height {
//	//	game.isOver = true
//	//	return
//	//}
//	//// collision check
//	//if indexOf(nt, game.snake) >= 0 {
//	//	game.isOver = true
//	//	return
//	//}
//	//
//	//foodIndex := indexOf(nt, game.food)
//	//if foodIndex < 0 {
//	//	// rotate in place, snake length remains the same
//	//	for i := 0; i < len(game.snake)-1; i++ {
//	//		game.snake[i] = game.snake[i+1]
//	//	}
//	//	game.snake[len(game.snake)-1] = nt
//	//} else {
//	//	// append tip to snake
//	//	game.snake = append(game.snake, nt)
//	//	// eat the food, place new food at random location
//	//	game.food[foodIndex] = game.placeRandomFood()
//	//	// incrmeent score
//	//	game.score++
//	//}
//}
//
//// Gets the current game tick speed. The game speeds up as the score increases.
//func (game *SnakeGame) GetTickDuration() time.Duration {
//	return time.Second * 3 / time.Duration(10+game.score/3)
//}
//
//func (game *SnakeGame) placeRandomFood() (ret Vec) {
//	for {
//		ret[0] = rand.Int() % game.width
//		ret[1] = rand.Int() % game.height
//		if indexOf(ret, game.snake) < 0 && indexOf(ret, game.food) < 0 {
//			return
//		}
//	}
//}
//
//func indexOf(loc Vec, arr []Vec) int {
//	for i := 0; i < len(arr); i++ {
//		if arr[i] == loc {
//			return i
//		}
//	}
//	return -1
//}
//
//func main() {
//	// Create game
//	game := CreateSnakeGame(gw, gh)
//	//Create canvas
//	s, err := tcell.NewScreen()
//	must(err)
//	must(s.Init())
//	s.SetStyle(styleBG)
//
//	// Poll events, allowing a single synchronous game loop using select{}
//	events := make(chan tcell.Event)
//	go func() {
//		for {
//			events <- s.PollEvent()
//		}
//	}()
//
//	// Game loop
//	ticker := time.NewTicker(game.GetTickDuration())
//	moveQ := []Vec{}
//	for {
//		// Render
//		s.Clear()
//		drawFrame(s)
//		drawScore(s, game.score)
//		drawGame(s, game)
//		s.Show()
//
//		// Poll event
//		select {
//		case <-ticker.C:
//			// game tick. start by turning the snake, if applicable.
//			// to make the game more fun, player can queue up a few moves
//			// that run on subsequent game ticks
//			origVel := game.vel
//			for len(moveQ) > 0 && game.vel == origVel {
//				game.SetVel(moveQ[0])
//				//copy(moveQ, moveQ[1:])
//				moveQ = moveQ[:len(moveQ)-1]
//			}
//			// next, advance the game by one turn
//			score, isOver := game.score, game.isOver
//			game.Step()
//			// beep when something happens
//			if game.score != score || game.isOver != isOver {
//				s.Beep()
//			}
//			// speed up the game tick as the score increases
//			ticker.Reset(game.GetTickDuration())
//
//		case ev := <-events:
//			// no game gick. instead, just deal with events
//			switch e := ev.(type) {
//			case *tcell.EventResize:
//				s.Sync()
//			case *tcell.EventKey:
//				switch e.Key() {
//				case tcell.KeyEsc:
//				case tcell.KeyCtrlC:
//					// exit. call s.Fini to return terminal to its normal state.
//					s.Fini()
//					return
//				default:
//					// player turns the snake at next tick, tick after etc.
//					moveQ = append(moveQ, getMove(e))
//				}
//			}
//		}
//	}
//}
//
////Gets the new snake velocity from WASD and arrow keys.
//func getMove(e *tcell.EventKey) Vec {
//	switch e.Key() {
//	case tcell.KeyUp:
//		return Vec{0, -1}
//	case tcell.KeyDown:
//		return Vec{0, 1}
//	case tcell.KeyLeft:
//		return Vec{-1, 0}
//	case tcell.KeyRight:
//		return Vec{1, 0}
//	case tcell.KeyRune:
//		switch e.Rune() {
//		case 'w':
//			return Vec{0, -1}
//		case 's':
//			return Vec{0, 1}
//		case 'a':
//			return Vec{-1, 0}
//		case 'd':
//			return Vec{1, 0}
//		}
//	}
//	return Vec{0, 0}
//}
//
//// Draws a box around the game board
//func drawFrame(s tcell.Screen) {
//	ox, oy := getCenterOffset(s)
//	for i := 0; i < gw*2+1; i++ {
//		s.SetContent(ox+i, oy+1, '-', nil, styleBG)
//		s.SetContent(ox+i, oy+gh+2, '-', nil, styleBG)
//	}
//	for i := 1; i < gh+2; i++ {
//		s.SetContent(ox, oy+i, '|', nil, styleBG)
//		s.SetContent(ox+gw*2+1, oy+i, '|', nil, styleBG)
//	}
//	s.SetContent(ox, oy+1, '+', nil, styleBG)
//	s.SetContent(ox, oy+gh+2, '+', nil, styleBG)
//	s.SetContent(ox+gw*2+1, oy+1, '+', nil, styleBG)
//	s.SetContent(ox+gw*2+1, oy+gh+2, '+', nil, styleBG)
//}
//
//// Draws SCORE ... XYZ below the game board
//func drawScore(s tcell.Screen, score int) {
//	ox, oy := getCenterOffset(s)
//	setText(s, ox+gw*2-5, oy+gh+3, fmt.Sprintf("%06d", score), styleBold)
//	setText(s, ox+1, oy+gh+3, "SCORE", styleBold)
//}
//
//// Draws the game board: snake, food, GAME OVER banner
//func drawGame(s tcell.Screen, game *SnakeGame) {
//	ox, oy := getCenterOffset(s)
//
//	// ...snake and food
//	for _, vec := range game.snake {
//		s.SetContent(ox+vec[0]*2+1, oy+vec[1]+2, ' ', nil, styleSnake)
//		s.SetContent(ox+vec[0]*2+2, oy+vec[1]+2, ' ', nil, styleSnake)
//	}
//	for _, vec := range game.food {
//		s.SetContent(ox+vec[0]*2+1, oy+vec[1]+2, ' ', nil, styleFood)
//		s.SetContent(ox+vec[0]*2+2, oy+vec[1]+2, ' ', nil, styleFood)
//	}
//
//	// ...game over
//	if game.isOver {
//		setText(s, ox+gw-6, oy+gh/2+1, "              ", styleGameOver)
//		setText(s, ox+gw-6, oy+gh/2+2, "  GAME  OVER  ", styleGameOver)
//		setText(s, ox+gw-6, oy+gh/2+3, "              ", styleGameOver)
//	}
//}
//
//// Gets the offset to center the game board in the terminal.
//func getCenterOffset(s tcell.Screen) (ox, oy int) {
//	sx, sy := s.Size()
//	ox = sx/2 - gw
//	oy = (sy - gh) / 2
//	return
//}
//
//func setText(s tcell.Screen, x, y int, text string, style tcell.Style) {
//	for i := 0; i < len(text); i++ {
//		s.SetContent(x+i, y, rune(text[i]), nil, style)
//	}
//}
//
//func must(err error) {
//	if err != nil {
//		panic(err)
//	}
//}
