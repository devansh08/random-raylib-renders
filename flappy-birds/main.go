package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	cat "github.com/catppuccin/go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type f32 = float32
type f64 = float64
type i32 = int32
type ui32 = uint32

const WIDTH = 1600
const HEIGHT = 900

const G = f32(150)

var Mocha = cat.Mocha

/* Bird Props */
const BirdSize = 50
const BirdJump = 150

type Bird struct {
	Position Point
	Speed    f32
}

var bird Bird

/* Pillar Props */
const PillarDistance = 300
const PillarWidth = 75
const PillarGap = 150
const PillarSpeed = 100

// PillarGapMax will tell how far vertically can 2 contigous pillar gaps be.
// If they are more than this distance apart, its physically impossible for the bird to clear the gaps.
var PillarGapMax = math.Min((PillarDistance/PillarSpeed)*BirdJump, math.Pow((PillarDistance/PillarSpeed), 2)*0.5*f64(G)) - 100

type Point = rl.Vector2
type Pillar struct {
	Id       ui32
	Height   i32 // From top of the screen
	Position Point
}

var MaxPillarsOnScreen = int(math.Ceil(f64(WIDTH) / f64(PillarDistance+PillarWidth)))
var pillars []Pillar

const FontFile = "/usr/share/fonts/TTF/JetBrainsMonoNLNerdFontMono-Regular.ttf"
const FontSize = 16

var Font128 rl.Font
var Font32 rl.Font

func WriteText(text string, x, y f32) {
	vec := rl.Vector2{X: x, Y: y}
	font := Font32
	if FontSize == 128 {
		font = Font128
	}
	rl.DrawTextEx(font, text, vec, FontSize, 0, getCatColor(Mocha.Text()))
}

func getCatColor(col cat.Color) color.RGBA {
	rgb := col.RGB
	return color.RGBA{
		R: rgb[0],
		G: rgb[1],
		B: rgb[2],
		A: 255,
	}
}

func randPillar(i ui32, xOffset f32) Pillar {
	randHeight := rand.Int31n(HEIGHT - PillarGap)
	if i > 0 {
		prevHeight := pillars[len(pillars)-1].Height
		for {
			if math.Abs(f64(prevHeight-randHeight)) >= f64(PillarGapMax) {
				randHeight = rand.Int31n(HEIGHT - PillarGap)
			} else {
				break
			}
		}
	}

	return Pillar{
		Id:     ui32(i),
		Height: randHeight,
		Position: Point{
			X: WIDTH + xOffset,
			Y: 0,
		},
	}
}

func initStuff() {
	bird = Bird{Position: Point{
		X: WIDTH/2 - BirdSize/2,
		Y: HEIGHT/2 - BirdSize/2,
	}, Speed: 0}

	pillars = make([]Pillar, 0, MaxPillarsOnScreen)
	for i := range MaxPillarsOnScreen {
		pillars = append(pillars, randPillar(ui32(i), f32(i*(PillarWidth+PillarDistance))))
	}
}

func movePillars(dt f32) {
	for i := range pillars {
		pillars[i].Position = Point{
			X: pillars[i].Position.X - PillarSpeed*dt,
			Y: 0,
		}
	}
}

func moveBird(dt f32) {
	bird.Speed += G * dt
	bird.Position.Y += (bird.Speed*dt + 0.5*G*dt*dt)

	if bird.Position.Y+BirdSize >= HEIGHT {
		bird.Position.Y = HEIGHT - BirdSize
	} else if bird.Position.Y <= 0 {
		bird.Position.Y = 0
		bird.Speed = 0
	}
}

func checkCollisions() bool {
	var pillar Pillar
	for i := range pillars {
		pillar = pillars[i]
		if bird.Position.X <= pillar.Position.X+PillarWidth {
			break
		}
	}

	if bird.Position.X+BirdSize < pillar.Position.X {
		return false
	} else if bird.Position.Y > pillar.Position.Y+f32(pillar.Height) && bird.Position.Y+BirdSize < pillar.Position.Y+f32(pillar.Height)+PillarGap {
		return false
	}

	return true
}

func updatePillars() {
	if pillars[0].Position.X+PillarWidth <= 0 {
		pillars = pillars[1:]
		lastPillar := pillars[len(pillars)-1]
		pillars = append(pillars, randPillar(lastPillar.Id+1, lastPillar.Position.X+PillarDistance+PillarWidth-WIDTH))
	}
}

func main() {
	rl.InitWindow(WIDTH, HEIGHT, "Flappy Birds")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)
	Font128 = rl.LoadFontEx(FontFile, 128, nil, 0)
	Font32 = rl.LoadFontEx(FontFile, 32, nil, 0)

	initStuff()

	var lastPillar ui32
	score := 0
	pause := false
	for !rl.WindowShouldClose() {
		pause = checkCollisions()

		rl.BeginDrawing()
		rl.ClearBackground(getCatColor(Mocha.Base()))

		if rl.IsKeyPressed(rl.KeyQ) {
			break
		}
		if rl.IsKeyPressed(rl.KeySpace) {
			bird.Speed = -BirdJump
		}

		for _, pillar := range pillars {
			rl.DrawRectangleV(pillar.Position, Point{X: PillarWidth, Y: f32(pillar.Height)}, getCatColor(Mocha.Text()))
			rl.DrawRectangleV(Point{
				X: pillar.Position.X,
				Y: pillar.Position.Y + f32(pillar.Height) + PillarGap,
			}, Point{
				X: PillarWidth,
				Y: f32(HEIGHT - pillar.Height - PillarGap),
			}, getCatColor(Mocha.Text()))

			if pillar.Id == lastPillar && pillar.Position.X+PillarWidth < bird.Position.X {
				score++
				lastPillar = pillar.Id + 1
			}
		}

		rl.DrawRectangleV(bird.Position, Point{X: BirdSize, Y: BirdSize}, getCatColor(Mocha.Sky()))

		if pause {
			textSize := rl.MeasureTextEx(Font128, "YOU DIED", 128, 0)
			rl.DrawTextEx(Font128, "YOU DIED", rl.Vector2{
				X: (WIDTH - textSize.X) / 2,
				Y: (HEIGHT - textSize.Y) / 2,
			}, 128, 0, getCatColor(Mocha.Red()))
		}

		rl.DrawTextEx(Font32, fmt.Sprintf("Score: %d", score), rl.Vector2{X: 10, Y: 10}, FontSize*2, 0, getCatColor(Mocha.Text()))

		if rl.IsKeyPressed(rl.KeyR) {
			initStuff()
			pause = true
		}

		rl.EndDrawing()

		if !pause {
			movePillars(rl.GetFrameTime())
			moveBird(rl.GetFrameTime())

			updatePillars()
		}
	}
}
