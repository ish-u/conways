package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

/*

	Ref -> https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life
*/

// ASCII Sequence Code - https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797

const (
	Horizontal  = "\u2500" // ─
	Vertical    = "\u2502" // │
	TopLeft     = "\u250C" // ┌
	TopRight    = "\u2510" // ┐
	BottomLeft  = "\u2514" // └
	BottomRight = "\u2518" // ┘
)

// Grid
type Cell struct {
	X, Y int
}
type Grid map[Cell]bool

func (grid Grid) addCell(cell Cell) {
	grid[cell] = true
}

func (grid Grid) removeCell(cell Cell) {
	delete(grid, cell)
}

func (grid Grid) getNeighbourCells(cell Cell) (dead []Cell, live []Cell) {
	directions := []Cell{
		{-1, 1},  // TOP LEFT
		{0, 1},   // TOP
		{1, 1},   // TOP RIGHT
		{1, 0},   // RIGHT
		{1, -1},  // BOTTOM RIGHT
		{0, -1},  // BOTTOM
		{-1, -1}, // BOTTOM LEFT
		{-1, 0},  // LEFT
	}
	for _, direction := range directions {
		if (grid[Cell{X: cell.X + direction.X, Y: cell.Y + direction.Y}]) {
			live = append(live, Cell{X: cell.X + direction.X, Y: cell.Y + direction.Y})
		} else {
			dead = append(dead, Cell{X: cell.X + direction.X, Y: cell.Y + direction.Y})
		}
	}

	return dead, live
}

// Seeds
type Seed string

const (
	ACORN       Seed = "ACORN"
	GLIDER      Seed = "GLIDER"
	R_PENTOMINO Seed = "R_PENTOMINO"
)

var SeedOrder = []Seed{ACORN, GLIDER, R_PENTOMINO}

var UniverSeeds = map[Seed]Grid{
	ACORN: {
		Cell{X: -2, Y: 1}:  true,
		Cell{X: -3, Y: -1}: true,
		Cell{X: -2, Y: -1}: true,
		Cell{X: 3, Y: -1}:  true,
		Cell{X: 0, Y: 0}:   true,
		Cell{X: 1, Y: -1}:  true,
		Cell{X: 2, Y: -1}:  true,
		Cell{X: 3, Y: -1}:  true,
	},
	GLIDER: {
		Cell{X: 0, Y: 0}:  true,
		Cell{X: 1, Y: 0}:  true,
		Cell{X: 2, Y: 0}:  true,
		Cell{X: 2, Y: -1}: true,
		Cell{X: 1, Y: -2}: true,
	},
	R_PENTOMINO: {
		{X: 0, Y: 0}:  true,
		{X: 1, Y: 0}:  true,
		{X: -1, Y: 1}: true,
		{X: 0, Y: 1}:  true,
		{X: 0, Y: 2}:  true,
	},
}

// Universe
type Universe struct {
	grid           Grid
	rows           int
	cols           int
	paused         bool
	generation     int
	currentSeedIdx int
}

func (universe *Universe) play() {
	universe.paused = false
}

func (universe *Universe) pause() {
	universe.paused = true
}

func (universe Universe) draw() {
	clearScreen()

	// Universe Window
	drawOnTerminal(1, 1, strings.Repeat(Horizontal, universe.cols))
	drawOnTerminal(1, 1, TopLeft)
	drawOnTerminal(1, universe.cols, TopRight)
	for i := range universe.rows {
		drawOnTerminal(i+2, 1, Vertical)
		drawOnTerminal(i+2, universe.cols, Vertical)
	}
	drawOnTerminal(universe.rows, 1, strings.Repeat(Horizontal, universe.cols))
	drawOnTerminal(universe.rows, 1, BottomLeft)
	drawOnTerminal(universe.rows, universe.cols, BottomRight)

	// Top Status Bar
	seedName := fmt.Sprintf(" %s ", string(SeedOrder[universe.currentSeedIdx]))
	drawOnTerminal(1, (universe.cols-len(seedName))/2, seedName)

	// Bottom Status Bar
	status := fmt.Sprintf(" GENERATION: %d | POPULATION: %d | PAUSED: %t ",
		universe.generation, len(universe.grid), universe.paused)
	drawOnTerminal(universe.rows, (universe.cols-len(status))/2, status)

	offsetX := universe.cols / 2
	offsetY := universe.rows / 2
	for cell := range universe.grid {
		row := offsetY - cell.Y + 1
		col := offsetX + cell.X + 1
		if row > 1 && col > 1 && row < universe.rows && col < universe.cols {
			drawOnTerminal(row, col, "\u25A3")
		}
	}
}

func (universe *Universe) tick() {
	newGrid := make(Grid)
	for cell := range universe.grid {
		deadNeighbours, aliveNeighbours := universe.grid.getNeighbourCells(cell)

		if len(aliveNeighbours) == 2 || len(aliveNeighbours) == 3 {
			newGrid.addCell(Cell{cell.X, cell.Y})
		}

		for _, deadNeighbour := range deadNeighbours {
			_, aliveNeighboursAroundDead := universe.grid.getNeighbourCells(deadNeighbour)
			if len(aliveNeighboursAroundDead) == 3 {
				newGrid.addCell(Cell{deadNeighbour.X, deadNeighbour.Y})
			}
		}

	}
	universe.grid = newGrid
	universe.generation++
}

func (universe *Universe) resetSeed() {
	universe.grid = UniverSeeds[SeedOrder[universe.currentSeedIdx]]
	universe.generation = 0
}

// Raw mode helpers
func drawOnTerminal(row int, col int, element string) {
	fmt.Printf("\x1B[%d;%dH%s", row, col, element)
}

func clearScreen() {
	fmt.Print("\x1B[2J")
}

func moveToHome() {
	fmt.Print("\x1B[H")
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}

func enableMouseEvents() {
	fmt.Print("\x1B[?1000h")
	fmt.Print("\x1B[?1006h")
}

func disableMouseEvents() {
	fmt.Print("\x1B[?1000l")
	fmt.Print("\x1B[?1006l")
}

func main() {

	// Handle Input
	inputChannel := make(chan []byte)
	go func() {
		buf := make([]byte, 32)
		for {
			_, err := os.Stdin.Read(buf)
			if err != nil {
				close(inputChannel)
				fmt.Println("Error Reading input")
				return
			}
			inputChannel <- buf
		}
	}()

	// Enabling Raw Mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(int(fd))
	if err != nil {
		fmt.Println("Error enabling raw mode:", err)
		panic(err)
	}
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		panic(err)
	}
	hideCursor()
	clearScreen()
	moveToHome()
	enableMouseEvents()

	// Disable Raw mode on return
	defer clearScreen()
	defer moveToHome()
	defer showCursor()
	defer disableMouseEvents()
	defer term.Restore(fd, oldState)

	universe := Universe{
		grid:           UniverSeeds[SeedOrder[0]],
		rows:           rows,
		cols:           cols,
		paused:         true,
		generation:     0,
		currentSeedIdx: 0,
	}

	// Render Loop
	for {
		universe.draw()
		if !universe.paused {
			universe.tick()
		}

		select {
		case keyBuffer := <-inputChannel:
			// Reading the Escape Code - ESC[{code};{string};{...}p
			key := keyBuffer[0]
			// Code for CTRL+C       -> 3
			// Code for Space  	     -> 32
			// Code for Mouse Click  -> \x1b[<%_BUTTON_;%_COL_;%_ROW_M
			if key == 3 {
				return
			} else if key == 32 {
				if universe.paused {
					universe.play()
				} else {
					universe.pause()
				}
			} else if key == 27 && len(keyBuffer) > 2 && keyBuffer[1] == '[' && keyBuffer[2] == '<' && universe.paused {
				var button, col, row int
				n, _ := fmt.Sscanf(string(keyBuffer), "\x1b[<%d;%d;%dM", &button, &col, &row)
				if n == 3 {
					cellX := col - universe.cols/2 - 1
					cellY := universe.rows/2 - row + 1
					if button == 0 {
						universe.grid.addCell(Cell{cellX, cellY})
					} else if button == 2 {
						universe.grid.removeCell(Cell{cellX, cellY})
					}
				}
			} else if key == 27 && (keyBuffer[1]) == '[' && (keyBuffer[2]) == 'C' && universe.paused {
				universe.currentSeedIdx = (universe.currentSeedIdx + 1) % len(SeedOrder)
				universe.resetSeed()
			} else if key == 27 && (keyBuffer[1]) == '[' && (keyBuffer[2]) == 'D' && universe.paused {
				universe.currentSeedIdx = (universe.currentSeedIdx - 1 + len(SeedOrder)) % len(SeedOrder)
				universe.resetSeed()
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

}
