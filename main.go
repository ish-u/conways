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

func main() {

	// Enabling Raw Mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(int(fd))
	if err != nil {
		fmt.Println("Error enabling raw mode:", err)
		panic(err)
	}

	// Diabling Raw Mode
	fmt.Print("\x1B[2J") // Clear Screen
	fmt.Print("\x1B[H")  // Move Cursor to Home
	defer term.Restore(fd, oldState)

}
