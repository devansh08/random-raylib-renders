package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"image/color"
	"image/png"
	"os"
)

func main() {
	f, err := os.Open("./image.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	image, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	width := int32(image.Bounds().Dx())
	height := int32(image.Bounds().Dx())

	rl.InitWindow(width, height, "Fog of War [raylib 1920 1080]")
	defer rl.CloseWindow()

	imageData := rl.NewImageFromImage(image)
	texture := rl.LoadTextureFromImage(imageData)

	rl.UnloadImage(imageData)

	rl.SetTargetFPS(60)

	radius := float32(100.0)
	BG := func(alpha uint8) color.RGBA {
		return color.RGBA{R: 30, G: 30, B: 46, A: alpha}
	}

	alphaMask := make([][]uint8, int(radius)*2)
	blurThickness := float32(25)

	for i := 0; i < int(radius)*2; i++ {
		alphaMask[i] = make([]uint8, int(radius)*2)
		for j := 0; j < int(radius)*2; j++ {
			d := rl.Vector2Distance(rl.Vector2{X: float32(i), Y: float32(j)}, rl.Vector2{X: radius, Y: radius})
			if d > radius {
				alphaMask[i][j] = 255
			} else if radius-d < blurThickness {
				alphaMask[i][j] = 255 - uint8(255*(radius-d)/blurThickness)
			}
		}
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(BG(255))

		mousePos := rl.GetMousePosition()
		mx, my := mousePos.X, mousePos.Y

		rl.DrawTextureRec(texture, rl.Rectangle{X: mx - radius, Y: my - radius, Width: 2.0 * radius, Height: 2.0 * radius}, rl.Vector2{X: mx - radius, Y: my - radius}, rl.White)

		for i := 0; i < int(radius)*2; i++ {
			for j := 0; j < int(radius)*2; j++ {
				if alpha := alphaMask[i][j]; alpha > 0 {
					rl.DrawRectangleV(rl.Vector2{X: float32(i) + mx - radius, Y: float32(j) + my - radius}, rl.Vector2{X: 1.0, Y: 1.0}, BG(alpha))
				}
			}
		}

		rl.DrawFPS(10, 10)

		rl.EndDrawing()
	}
}
