package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"math"
	"math/rand"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// FPS - Set frames per second
	FPS = 15
	// ROWS - Number of rows in grid
	ROWS = 288
	// COLS - Number of columns in grid
	COLS = 512
	// CellWidth - Width of individual cell
	CellWidth = 5
	// CellHeight - Height of individual cell
	CellHeight = 5
	// SeedCount - Number of seeds to start with
	SeedCount = 50
	// Spread - Radius of initial pattern
	Spread = 10
)

var (
	// BG - Background color (Catppuccin Mocha - Base)
	BG = rl.NewColor(30, 30, 46, 255)
	// FG - Foreground color (Catppuccin Mocha)
	FG = []rl.Color{
		rl.NewColor(137, 180, 250, 255),
		rl.NewColor(245, 224, 220, 255),
		rl.NewColor(242, 205, 205, 255),
		rl.NewColor(245, 194, 231, 255),
		rl.NewColor(203, 166, 247, 255),
		rl.NewColor(243, 139, 168, 255),
		rl.NewColor(235, 160, 172, 255),
		rl.NewColor(250, 179, 135, 255),
		rl.NewColor(249, 226, 175, 255),
		rl.NewColor(166, 227, 161, 255),
		rl.NewColor(148, 226, 213, 255),
		rl.NewColor(137, 220, 235, 255),
		rl.NewColor(116, 199, 236, 255),
		rl.NewColor(137, 180, 250, 255),
		rl.NewColor(180, 190, 254, 255),
	}
)

// Result - Describes multi value result from a function's return
type Result struct {
	Grid [][]rl.Color
	Dead bool
}

func gol(grid *[][]rl.Color) *Result {
	directions := [8][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	newGrid := make([][]rl.Color, COLS)
	for x := range newGrid {
		newGrid[x] = make([]rl.Color, ROWS)
		for y := range newGrid[x] {
			newGrid[x][y] = BG
		}
	}

	dead := false
	for r := range ROWS {
		for c := range COLS {
			count := 0
			colors := []rl.Color{}

			for _, dir := range directions {
				nx, ny := r+dir[0], c+dir[1]
				if nx >= 0 && nx < ROWS && ny >= 0 && ny < COLS && (*grid)[ny][nx] != BG {
					count++
					colors = append(colors, (*grid)[ny][nx])
					dead = true
				}
			}

			if (*grid)[c][r] != BG && (count == 2 || count == 3) {
				newGrid[c][r] = (*grid)[c][r]
			} else if (*grid)[c][r] == BG && count == 3 {
				newGrid[c][r] = colors[rand.Intn(len(colors))]
			}
		}
	}

	return &Result{Grid: newGrid, Dead: dead}
}

func seeds(grid *[][]rl.Color, xx, yy int) {
	setColor := func(x, y int, v float64, col rl.Color) {
		if v != 0.0 && rand.Float64() < v {
			(*grid)[x][y] = col
		}
	}

	curve := func(x float64) float64 {
		if x <= 0.15 {
			return 1
		}
		t := (x - 0.15) / (1 - 0.15)
		return math.Pow(1-t, 3)
	}

	col := FG[rand.Intn(len(FG))]

	for j := 1; j <= Spread; j++ {
		normX := float64(j) / Spread
		if j >= ROWS-yy || j >= yy || j >= COLS-xx || j >= xx {
			continue
		}

		for d := -j; d <= j; d++ {
			setColor(j+xx, d+yy, curve(normX), col)
			setColor(-j+xx, d+yy, curve(normX), col)
			setColor(d+xx, j+yy, curve(normX), col)
			setColor(d+xx, -j+yy, curve(normX), col)
		}
	}
}

func serialize(grid *[][]rl.Color) string {
	buf := new(bytes.Buffer)
	for _, row := range *grid {
		for _, col := range row {
			binary.Write(buf, binary.LittleEndian, col.R)
			binary.Write(buf, binary.LittleEndian, col.G)
			binary.Write(buf, binary.LittleEndian, col.B)
			binary.Write(buf, binary.LittleEndian, col.A)
		}
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func main() {
	os.Setenv("GLFW_PLATFORM", "x11")
	rl.InitWindow(COLS*CellWidth, ROWS*CellHeight, "Game of Life")
	defer rl.CloseWindow()
	rl.SetTargetFPS(FPS)

	grid := make([][]rl.Color, COLS)
	area := make([][]rl.Color, COLS)
	for x := range grid {
		grid[x] = make([]rl.Color, ROWS)
		area[x] = make([]rl.Color, ROWS)
		for y := range grid[x] {
			grid[x][y] = BG
			area[x][y] = BG
		}
	}

	states := make([]string, 0, 3)
	for i := range SeedCount {
		q := min(int(4*(i-1)/(SeedCount-1)), 3) + 1
		var x, y int
		switch q {
		case 1:
			x, y = rand.Intn(ROWS/2), rand.Intn(COLS/2)
		case 2:
			x, y = rand.Intn(ROWS/2), rand.Intn(COLS/2)+COLS/2
		case 3:
			x, y = rand.Intn(ROWS/2)+ROWS/2, rand.Intn(COLS/2)
		case 4:
			x, y = rand.Intn(ROWS/2)+ROWS/2, rand.Intn(COLS/2)+COLS/2
		}
		seeds(&grid, y, x)
	}

	pause := false
	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyQ) {
			break
		}
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			pause = !pause
		}

		rl.BeginDrawing()
		rl.ClearBackground(BG)

		for i := range COLS {
			for j := range ROWS {
				col := grid[i][j]
				if col == BG {
					rl.DrawRectangle(CellWidth*int32(i), CellHeight*int32(j), CellWidth, CellHeight, area[i][j])
				} else {
					area[i][j] = rl.NewColor(col.R, col.G, col.B, 32)
					rl.DrawRectangle(CellWidth*int32(i), CellHeight*int32(j), CellWidth, CellHeight, grid[i][j])
				}
			}
		}

		if len(states) == 3 {
			states = states[1:]
		}
		states = append(states, serialize(&grid))

		if !pause {
			result := gol(&grid)
			grid = result.Grid
			if !result.Dead {
				pause = true
			}
		}

		if len(states) == 3 && states[0] == states[2] {
			pause = true
		}

		rl.EndDrawing()
	}
}
