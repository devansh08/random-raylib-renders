package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const WIDTH = 1920
const HEIGHT = 1080
const FPS = 60

const CELL = 50

var ROWS int
var COLS int

type Cell struct {
	X, Y int32
}

var hoverCell Cell

type CellType int

const (
	Void CellType = iota
	Path
	Start
	End
)

type Direction int

const (
	Right = iota
	Down
	Left
	Up
	DirectionCount
)

var dirMap = map[Direction]rl.Vector2{
	Right: rl.Vector2{X: 1, Y: 0},
	Down:  rl.Vector2{X: 0, Y: 1},
	Left:  rl.Vector2{X: -1, Y: 0},
	Up:    rl.Vector2{X: 0, Y: -1},
}

var visitedMap map[Cell]map[Direction]bool

var grid [][]CellType

var startCells []Cell
var endCells []Cell

var COL_BG color.RGBA
var COL_FG color.RGBA
var COL_HOVER color.RGBA
var COL_START color.RGBA
var COL_END color.RGBA
var COL_MOB color.RGBA

type Mode int

const (
	Draw Mode = iota
	Play
)

var mode Mode

const MOB_COUNT = 1000
const MOB_RAD = 5
const MOB_SPEED = 100

type MobState int

const (
	Moving MobState = iota
	Standing
)

type Mob struct {
	startCell     Cell
	endCell       Cell
	path          []Cell
	pathPoints    []rl.Vector2
	position      rl.Vector2
	initPosition  rl.Vector2
	nextPosition  rl.Vector2
	nextIndex     int
	relPosition   rl.Vector2
	speed         float32
	timeToNextPos float32
	timeElapsed   float32
	moveState     MobState
}

var mobs []Mob

var mobPaths map[[2]Cell][]Cell

func initialize() {
	COL_BG = color.RGBA{R: 0x13, G: 0x1F, B: 0x20, A: 0xFF}
	COL_FG = color.RGBA{R: 0x56, G: 0x82, B: 0x70, A: 0xFF}
	COL_HOVER = color.RGBA{R: 0x91, G: 0xC4, B: 0xC0, A: 0x33}
	COL_START = color.RGBA{R: 0x94, G: 0xC6, B: 0x8B, A: 0xFF}
	COL_END = color.RGBA{R: 0xD8, G: 0x46, B: 0x4B, A: 0xFF}
	COL_MOB = color.RGBA{R: 0xFF, G: 0xD7, B: 0xAA, A: 0xFF}

	ROWS = HEIGHT / CELL
	COLS = WIDTH / CELL

	hoverCell = Cell{X: -1, Y: -1}

	grid = make([][]CellType, ROWS)
	for i := range grid {
		grid[i] = make([]CellType, COLS)
	}

	mobPaths = map[[2]Cell][]Cell{}
}

func renderGrid(drawLines bool) {
	for i := range grid {
		for j := range grid[i] {
			if i == int(hoverCell.Y) && j == int(hoverCell.X) {
				rl.DrawRectangle(int32(j*CELL), int32(i*CELL), CELL, CELL, COL_HOVER)
			} else if grid[i][j] == Start {
				rl.DrawRectangle(int32(j*CELL), int32(i*CELL), CELL, CELL, COL_START)
			} else if grid[i][j] == End {
				rl.DrawRectangle(int32(j*CELL), int32(i*CELL), CELL, CELL, COL_END)
			} else if grid[i][j] == Path {
				rl.DrawRectangle(int32(j*CELL), int32(i*CELL), CELL, CELL, COL_FG)
			}
		}
	}

	if drawLines {
		for i := range len(grid) {
			rl.DrawLineEx(rl.Vector2{X: 0, Y: float32(i * CELL)}, rl.Vector2{X: WIDTH, Y: float32(i * CELL)}, 1, COL_FG)
			rl.DrawText(fmt.Sprintf("%d", i), 0, int32(i*CELL), 16, rl.White)
		}
		for i := range len(grid[0]) {
			rl.DrawLineEx(rl.Vector2{X: float32(i * CELL), Y: 0}, rl.Vector2{X: float32(i * CELL), Y: HEIGHT}, 1, COL_FG)
			rl.DrawText(fmt.Sprintf("%d", i), int32(i*CELL), 0, 16, rl.White)
		}
	}
}

func printGrid() {
	rl.TraceLog(rl.LogInfo, "  "+strings.Repeat("-", COLS*2+5))
	for i := 0; i < ROWS; i++ {
		line := fmt.Sprintf("%3d|", i)
		for j := 0; j < COLS; j++ {
			switch grid[i][j] {
			case Void:
				line += " ·"
			case Path:
				line += " #"
			case Start:
				line += " S"
			case End:
				line += " E"
			}
		}
		rl.TraceLog(rl.LogInfo, line)
	}
}

func isValid(current Cell) bool {
	return current.X >= 0 && current.X < int32(COLS) && current.Y >= 0 && current.Y < int32(ROWS) && grid[current.Y][current.X] != Void
}

func (mob *Mob) calcPath(current, prev Cell, dir Direction, path []Cell) {
	if slices.Contains(path[:len(path)-1], current) {
		return
	}

	if isValid(current) && current == mob.endCell {
		if len(mob.path) == 0 || len(path) < len(mob.path) {
			mob.path = make([]Cell, len(path))
			copy(mob.path, path)

			mobPaths[[2]Cell{mob.startCell, mob.endCell}] = make([]Cell, len(path))
			copy(mobPaths[[2]Cell{mob.startCell, mob.endCell}], path)
		}
		if !visitedMap[prev][dir] {
			visitedMap[prev][dir] = true
		}

		return
	}

	if visitedMap[current] == nil {
		visitedMap[current] = map[Direction]bool{}
	}

	for i := range DirectionCount {
		d := Direction(i)
		next := Cell{X: current.X + int32(dirMap[d].X), Y: current.Y + int32(dirMap[d].Y)}
		if next != prev && isValid(next) && !visitedMap[current][d] {
			visitedMap[current][d] = true
			mob.calcPath(next, current, d, append(path, next))
		}
	}
}

func (mob *Mob) goToNext(current rl.Vector2) {
	if mob.nextIndex < len(mob.pathPoints)-1 {
		mob.nextIndex++
		mob.initPosition = current
		mob.nextPosition = mob.pathPoints[mob.nextIndex]

		mob.timeToNextPos = linearDistance(mob.initPosition, mob.nextPosition) / mob.speed
		mob.timeElapsed = 0
	} else {
		mob.moveState = Standing
	}
}

func (mob *Mob) update(delta float32) {
	if mob.position == mob.nextPosition {
		mob.goToNext(mob.position)
	}

	if mob.moveState == Moving {
		mob.timeElapsed += delta

		t := mob.timeElapsed / mob.timeToNextPos
		if t > 1.0 {
			t = 1.0
		}
		mob.position.X = mob.initPosition.X + (mob.nextPosition.X-mob.initPosition.X)*t
		mob.position.Y = mob.initPosition.Y + (mob.nextPosition.Y-mob.initPosition.Y)*t
	}
}

func (mob Mob) render() {
	rl.DrawCircleV(mob.position, MOB_RAD, COL_MOB)
}

func linearDistance(a, b rl.Vector2) float32 {
	return float32(math.Sqrt(float64((a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y))))
}

func genMob() Mob {
	startCell := startCells[rand.Intn(len(startCells))]
	endCell := endCells[rand.Intn(len(endCells))]

	mob := Mob{
		startCell:   startCell,
		endCell:     endCell,
		relPosition: rl.Vector2{X: rand.Float32() * CELL, Y: rand.Float32() * CELL},
		speed:       MOB_SPEED,
		path:        []Cell{},
		pathPoints:  []rl.Vector2{},
	}

	// printGrid()

	if len(mobPaths[[2]Cell{startCell, endCell}]) == 0 {
		visitedMap = map[Cell]map[Direction]bool{}

		mob.calcPath(startCell, Cell{}, Right, []Cell{mob.startCell})
	} else {
		mob.path = mobPaths[[2]Cell{startCell, endCell}]
	}

	if mob.path == nil || len(mob.path) == 0 {
		panic(fmt.Errorf("No path found for mob starting from %+v to %+v", startCell, endCell))
	}

	for _, cell := range mob.path {
		mob.pathPoints = append(mob.pathPoints, rl.Vector2{X: float32(cell.X*CELL) + mob.relPosition.X, Y: float32(cell.Y*CELL) + mob.relPosition.Y})
	}

	mob.nextIndex = 0
	mob.goToNext(mob.pathPoints[0])
	mob.position = mob.initPosition

	mob.moveState = Moving

	mob.render()

	return mob
}

func main() {
	initialize()

	rl.SetTraceLogLevel(rl.LogDebug)

	rl.InitWindow(WIDTH, HEIGHT, "Tower Defense [raylib 1920 1080]")
	rl.SetTargetFPS(FPS)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(COL_BG)

		renderGrid(false)

		if mode == Draw {
			mobPaths = map[[2]Cell][]Cell{}
			mobs = []Mob{}

			mousePos := rl.GetMousePosition()
			mouseCol := int32(math.Floor(float64(mousePos.X / CELL)))
			mouseRow := int32(math.Floor(float64(mousePos.Y / CELL)))

			hoverCell = Cell{X: mouseCol, Y: mouseRow}

			if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
				grid[mouseRow][mouseCol] = Path
			} else if rl.IsMouseButtonDown(rl.MouseButtonRight) {
				grid[mouseRow][mouseCol] = Void
			} else if rl.IsKeyPressed(rl.KeyS) {
				if rl.IsKeyDown(rl.KeyLeftShift) {
					grid[mouseRow][mouseCol] = Path
				} else if grid[mouseRow][mouseCol] == Path {
					grid[mouseRow][mouseCol] = Start
					startCells = append(startCells, Cell{X: mouseCol, Y: mouseRow})
				}
			} else if rl.IsKeyPressed(rl.KeyE) {
				if rl.IsKeyDown(rl.KeyLeftShift) {
					grid[mouseRow][mouseCol] = Path
				} else if grid[mouseRow][mouseCol] == Path {
					grid[mouseRow][mouseCol] = End
					endCells = append(endCells, Cell{X: mouseCol, Y: mouseRow})
				}
			}
		} else if mode == Play {
			if len(mobs) < MOB_COUNT {
				for i := len(mobs); i < MOB_COUNT; i++ {
					mobs = append(mobs, genMob())
				}
			}

			for i := range mobs {
				if rl.IsKeyPressed(rl.KeyLeftControl) {
					if mobs[i].moveState == Moving && rand.Float32() >= 0.5 {
						mobs[i].speed /= 2
					}
				} else if rl.IsKeyPressed(rl.KeyRightControl) {
					if mobs[i].moveState == Moving && rand.Float32() >= 0.5 {
						mobs[i].speed *= 2
					}
				}
				mobs[i].update(rl.GetFrameTime())
				mobs[i].render()
			}
		}

		if rl.IsKeyPressed(rl.KeySpace) {
			if mode == Draw {
				mode = Play
			} else if mode == Play {
				mode = Draw
			}
		}

		rl.EndDrawing()
	}
}
