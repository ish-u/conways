package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

/*

	Ref -> https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life

	TODO
	--------
	- We need a Data Structure to store the current live cells -> we can use a Map with "(i,j)" string as the key
	- we will need to read through this map -> apply rules -> get the new state -> print it on the terminal
	- We will use the Raw mode and create a rectangle with borders -> (N-2)x(M-2)
	- At any point this rectangle will show a part of the infinite 2D Grid => [A, B, C, D]
	- At the start we will have the (0,0) on the center
	- We can move across the grid with arrow keys
	- We can allow the selection of seed using mouse at the start (?)
	--------
*/

// ASCII Sequence Code - https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797

// Grid
type Cell struct {
	X, Y int
}
type Grid map[Cell]bool

func (grid Grid) addCell(cell Cell) {
	grid[cell] = true
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

type Universe struct {
	grid   Grid
	window [4]int
	rows   int
	cols   int
	paused bool
}

func (universe *Universe) play() {
	universe.paused = false
}

func (universe *Universe) pause() {
	universe.paused = true
}

func (universe Universe) draw() {
	clearScreen()
	offsetX := universe.cols / 2
	offsetY := universe.rows / 2
	for cell := range universe.grid {
		row := offsetY - cell.Y
		col := offsetX + cell.X
		draw(row, col, "◻")
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
}

// Raw mode helpers
func draw(row int, col int, element string) {
	fmt.Printf("\x1B[%d;%dH%s", row, col, element)
}

func clearScreen() {
	fmt.Print("\x1B[2J")
}

func moveToHome() {
	fmt.Print("\x1B[H")
}

func showCussor() {
	fmt.Print("\x1b[?25h")
}

func hideCussor() {
	fmt.Print("\x1b[?25l")
}

func main() {

	// Handle Input
	inputChannel := make(chan []byte)
	go func() {
		buf := make([]byte, 3)
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
	hideCussor()
	clearScreen()
	moveToHome()

	// Disable Raw mode on return
	defer clearScreen()
	defer moveToHome()
	defer showCussor()
	defer term.Restore(fd, oldState)

	universe := Universe{
		grid:   make(Grid),
		window: [4]int{-rows / 2, -cols / 2, rows / 2, cols / 2},
		rows:   rows,
		cols:   cols,
		paused: true,
	}
	universe.grid.addCell(Cell{0, -1})
	universe.grid.addCell(Cell{0, 0})
	universe.grid.addCell(Cell{0, 1})

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
			// Code for CTRL+C is 3
			if key == 3 {
				return
			} else if key == 32 {
				if universe.paused {
					universe.play()
				} else {
					universe.pause()
				}
			}
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}

}
