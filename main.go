package main

import (
	"fmt"
	"os"

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

func (grid Grid) addCell(i int, j int) {
	grid[Cell{i, j}] = true
}

func (grid Grid) removeCell(i int, j int) {
	delete(grid, Cell{i, j})
}

type Universe struct {
	grid   Grid
	window [4]int
	rows   int
	cols   int
}

func (universe Universe) draw() {
	for cell, _ := range universe.grid {
		draw(cell.X, cell.Y, "0")
	}
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
	fmt.Print("\x1b[?25l")
}

func hideCussor() {
	fmt.Print("\x1b[?25l")
}

func main() {

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

	universe := Universe{
		grid:   make(Grid),
		window: [4]int{-rows / 2, -cols / 2, rows / 2, cols / 2},
		rows:   rows,
		cols:   cols,
	}
	universe.grid.addCell(0, 0)

	// Render Loop
	buf := make([]byte, 3)
	for {
		universe.draw()

		_, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println("Error Reading input")
			break
		}

		// Reading the Escape Code - ESC[{code};{string};{...}p
		key := buf[0]
		// Code for CTRL+C is 3
		if key == 3 {
			break
		}

	}

	// Diabling Raw Mode
	clearScreen()
	moveToHome()
	showCussor()
	defer term.Restore(fd, oldState)

}
