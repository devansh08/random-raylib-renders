package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Point = rl.Vector2

type Ant struct {
	Position     Point
	NextPosition Point
	Path         []Point
	Target       Target
	HasFood      bool
	Color        rl.Color
	Id           uint
}

// Range - Radius of influence the target has on the ant.
// If the ant is within this range it should start being more focused towards this target.
// Focusing will be based on descentCurve function.
// Optional; if not specified there will be no influence an ant.
type Target struct {
	Position Point
	Size     int32
	Range    int32
}

// Time - How much longer will this spot stay alive ?
// Will be decided by some descent function (f(x)) based on time left (x).
// Size - Same as range for this.
type Pheromone struct {
	Position Point
	Size     float32
	Time     int32
	AntId    uint
}

/* Raylib props */
const Width = 2560.0
const Height = 1440.0
const FPS = 60

const FontFile = "/usr/share/fonts/TTF/JetBrainsMonoNLNerdFontMono-Regular.ttf"
const FontSize = 16

var Font rl.Font

var BG = rl.NewColor(30, 30, 46, 255)
var RED = rl.NewColor(210, 15, 57, 255)
var GREEN = rl.NewColor(64, 160, 43, 255)
var BLUE = rl.NewColor(30, 102, 245, 255)
var TEXT = rl.NewColor(205, 214, 244, 255)
var FG = []rl.Color{
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

/* Target Props */
var target = Target{Range: 250, Size: 25}

/* Nest Props */
var nest = Target{Range: 500, Size: 10}

const AntCount = 20

/* Ant Props */
var ants []Ant

// MaxAntAngle - Maximum angle of focus the ant can have, even if it is out of TargetRange.
const MaxAntAngle = 270

// MinAntAngle - Minimum angle of focus the ant can have, even if it is near the target.
const MinAntAngle = 5

// AntSpeed - Speed of ant movement.
const AntSpeed = 350

const AntSize = 5

// pheromones - Collective list of pheromone spots from all ants.
// Should make it easy to parse for each movement of an ant.
// Queue like structure; new spots added to the start.
// Elements should be removed if the time remaining is 0.
var pheromones []Pheromone

// PheromoneMaxRange - Range of influence a pheromone spot has on another ant.
const PheromoneMaxRange = 500

// PheromoneMaxTime - Time for a pheromone spot stays alive (in seconds).
// The influence should keep reducing over time (as a function over each second of time).
const PheromoneMaxTime = 30 * FPS

const PheromoneMaxSize = 10

// PheromoneFreq - Frequency per second for dropping a pheromone circle.
// Also conincides with the number of direction changes the ant will make in each second.
const PheromoneFreq = 5

/* General Props */
var MaxDistance = distance(Point{X: Width, Y: Height}, Point{X: 0, Y: 0})

/* Raylib util functions */
func WriteText(text string, x, y float32) {
	vec := rl.Vector2{X: x, Y: y}
	rl.DrawTextEx(Font, text, vec, FontSize, 0, TEXT)
}

func WriteTextPoint(text string, point Point) {
	vec := rl.Vector2{X: point.X, Y: point.Y}
	rl.DrawTextEx(Font, text, vec, FontSize, 0, TEXT)
}

func Lighten(color rl.Color, alpha uint8) rl.Color {
	return rl.NewColor(color.R, color.G, color.B, alpha)
}

/* Math functions */
// curve(float32) float32 - Emulates f(x) for some curve function.
// This would define how focused the ant is towards the target (y) based its distance to the target (x).
func curve(x float32) float32 {
	// Linear function
	return x
}

// normalize - Return normalized value between [0, 1].
func normalize(v, minimum, maximum float32) float32 {
	return min((v-minimum)/(maximum-minimum), 1)
}

// scale - Return a value between min and max based on a value in [0, 1].
func scale(v, minimum, maximum float32) float32 {
	return v*(maximum-minimum) + minimum
}

func distance(p1, p2 Point) float32 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return float32(math.Sqrt(dx*dx + dy*dy))
}

// angle - Returns the angle of a line between 2 points.
func angleRads(p1, p2 Point) float64 {
	// dAngle := rad2Deg(math.Atan2(float64(ant.Target.Position.Y-ant.Position.Y), float64(ant.Target.Position.X-ant.Position.X)))
	return math.Atan2(float64(p1.Y-p2.Y), float64(p1.X-p2.X))
}

func deg2Rad(deg float64) float64 {
	return deg * math.Pi / 180
}

func rad2Deg(rad float64) float64 {
	return rad * 180 / math.Pi
}

// extendLine - Returns second point of a line; given first point (ant), angle to x axis (focus angle) and distance of line (temp).
func extendLine(p Point, angle, dist float64) Point {
	return Point{X: float32(float64(p.X) + dist*math.Cos(angle)), Y: float32(float64(p.Y) + dist*math.Sin(angle))}
}

/* Functions */
func randPoint() Point {
	return Point{X: scale(rand.Float32(), 0.0, Width), Y: scale(rand.Float32(), 0.0, Height)}
}

// focusAngle - Get the angle to move the ant in given the distance of the ant from the target and the infleunce radius of the target.
func focusAngle(dist float32, maxVal float32) float32 {
	return scale(curve(normalize(dist, 0, maxVal)), MinAntAngle, MaxAntAngle)
}

func updatePheromones() {
	deleteFlag := false
	index := -1
	str := ""

	for i := range pheromones {
		spot := &pheromones[i]
		spot.Time--
		str = fmt.Sprintf("%s|%d", str, spot.Time)

		spot.Size = scale(curve(normalize(float32(spot.Time), 0, PheromoneMaxTime)), 0, PheromoneMaxRange)

		if spot.Time <= 0 {
			deleteFlag = true
		}
		if deleteFlag && index == -1 && (spot.Time > 0 || i == len(pheromones)-1) {
			index = i
		}
	}

	if deleteFlag && index > -1 {
		pheromones = pheromones[index:]
	}

	WriteText(fmt.Sprintf("%d | %v - %d | %s", len(pheromones), deleteFlag, index, str), 10, Height-FontSize-10)
}

func moveAnt(ant *Ant, frame int, dt float32) {
	if frame%(FPS/PheromoneFreq) == 0 {
		dist := distance(ant.Target.Position, ant.Position)
		if dist < AntSpeed/PheromoneFreq {
			ant.NextPosition = ant.Target.Position
		} else {
			target := ant.Target
			if !ant.HasFood && len(pheromones) > 1 && dist > float32(ant.Target.Range) {
				// Lesser the Time the better, as it means it is more closer to the target
				minTime := int32(PheromoneMaxTime)
				for _, p := range pheromones {
					if math.Abs(float64(ant.Position.X-p.Position.X)) < 50 && math.Abs(float64(ant.Position.Y-p.Position.Y)) < 50 {
						if p.Time < minTime {
							minTime = p.Time
							target = Target{Position: p.Position, Range: 0}
							if ant.Id == 0 {
								rl.DrawCircleV(target.Position, 30, rl.Red)
							}
						}
					}
				}
			}

			angle := focusAngle(dist, float32(target.Range))

			dAngle := rad2Deg(angleRads(target.Position, ant.Position))
			angle1, angle2 := deg2Rad(dAngle-float64(angle/2)), deg2Rad(dAngle+float64(angle/2))

			randAngle := scale(rand.Float32(), float32(min(angle1, angle2)), float32(max(angle1, angle2)))
			t := Point{}
			for {
				t = extendLine(ant.Position, float64(randAngle), AntSpeed/PheromoneFreq)
				if !(t.X >= Width || t.X < 0 || t.Y >= Height || t.Y < 0) {
					break
				}
				randAngle = scale(rand.Float32(), float32(min(angle1, angle2)), float32(max(angle1, angle2)))
			}

			ant.NextPosition = t

			if ant.HasFood {
				ant.Path = append(ant.Path, ant.Position)
				if len(ant.Path) > 1 {
					pheromones = append(pheromones, Pheromone{Position: ant.Path[len(ant.Path)-2], Time: PheromoneMaxTime, Size: PheromoneMaxRange, AntId: ant.Id})
				}
			}
		}
	} else {
		// And on every frame move the ant by a distance of 1/FPS of the distance it should actually move to reach its next position
		ant.Position = extendLine(ant.Position, angleRads(ant.NextPosition, ant.Position), float64(AntSpeed*dt))
	}

	WriteText(fmt.Sprintf("(%f, %f) -> (%f, %f)", ant.Position.X, ant.Position.Y, ant.Target.Position.X, ant.Target.Position.Y), 10, 10+float32(ant.Id)*FontSize)
}

func initPoints() {
	target.Position = randPoint()
	nest.Position = randPoint()

	ants = make([]Ant, 0, AntCount)
	for i := range AntCount {
		ants = append(ants, Ant{Position: nest.Position, Color: FG[rand.Intn(len(FG))], Path: []Point{}, Target: target, HasFood: false, Id: uint(i)})
	}
}

/* Main */
func main() {
	os.Setenv("GLFW_PLATFORM", "x11")

	rl.InitWindow(Width, Height, "Ants!")
	defer rl.CloseWindow()

	rl.SetTargetFPS(FPS)
	Font = rl.LoadFont(FontFile)

	initPoints()

	start := false
	frame := 0
	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyQ) {
			break
		}
		if rl.IsKeyPressed(rl.KeyR) {
			initPoints()
		}
		if rl.IsKeyPressed(rl.KeySpace) {
			start = true
		}

		rl.BeginDrawing()
		rl.ClearBackground(BG)

		frame++
		rl.DrawFPS(10, Height-FontSize*2-15)

		rl.DrawCircleV(target.Position, float32(target.Size), RED)
		WriteTextPoint(fmt.Sprintf("%d", target.Range), target.Position)
		rl.DrawCircleV(nest.Position, float32(nest.Size), GREEN)

		if start {
			updatePheromones()
			for j := range pheromones {
				rl.DrawCircleV(pheromones[j].Position, scale(normalize(pheromones[j].Size, 0, PheromoneMaxRange), 0, PheromoneMaxSize), Lighten(BLUE, 64))
			}

			for i := range ants {
				ant := &ants[i]
				rl.DrawCircleV(ant.Position, AntSize, ant.Color)

				if distance(ant.Position, ant.Target.Position) <= float32(ant.Target.Size) {
					if ant.Target == target {
						ant.HasFood = true
						ant.Target = nest
						target.Range--
					} else {
						ant.HasFood = false
						ant.Target = target
					}
					ant.Path = []Point{}
					continue
				}

				moveAnt(ant, frame, rl.GetFrameTime())
			}
		}

		rl.EndDrawing()
	}
}
