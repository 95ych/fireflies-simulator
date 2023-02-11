package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	WIN_HEIGHT    = 1080
	WIN_WIDTH     = 1920
	SCALE         = 1
	MAX_FIREFLIES = 1000
	MAX_SPEED     = 8
	MIN_SPEED     = 3
	TIMER         = 36000
	TICK_RATE     = 100
	SYNC_UP       = 1000
	FADE_RATE     = 0.008
	flyCount      = 900
	radius        = 4
	proximity     = 200
)

var dt = 0.15

type position struct {
	x, y float64
}

type Firefly struct {
	pos   position
	speed float64
	angle float64
	alpha float64
	flash bool
	clock float64
}

func manhattanDist(f1 Firefly, f2 Firefly) float64 {
	return math.Abs(f1.pos.x-f2.pos.x) + math.Abs(f1.pos.y-f2.pos.y)
}

func (fly *Firefly) syncUp(neighborFly *Firefly) {
	if manhattanDist(*fly, *neighborFly) < proximity {
		nudgeFactor := neighborFly.clock / TIMER
		neighborFly.clock = math.Min(neighborFly.clock+nudgeFactor*SYNC_UP, TIMER)
	}
}

func (fly *Firefly) Init() {
	fly.pos.x = rand.Float64() * WIN_WIDTH * SCALE
	fly.pos.y = rand.Float64() * WIN_HEIGHT * SCALE
	fly.angle = rand.Float64()*2*math.Pi - math.Pi
	fly.speed = MIN_SPEED + rand.Float64()*(MAX_SPEED-MIN_SPEED)
	fly.clock = rand.Float64() * TIMER
}

func wrapAround(x, y float64) (float64, float64) {
	x = math.Mod(x, WIN_WIDTH)
	if x < 0 {
		x += WIN_WIDTH
	}
	y = math.Mod(y, WIN_HEIGHT)
	if y < 0 {
		y += WIN_HEIGHT
	}
	return x, y
}

func (fly *Firefly) Update() {
	fly.pos.x += fly.speed * math.Cos(fly.angle) * dt
	fly.pos.y += fly.speed * math.Sin(fly.angle) * dt
	fly.alpha = math.Max(fly.alpha-FADE_RATE, 0)
	// rand.Seed(time.Now().UnixNano())
	fly.angle += (rand.Float64() - 0.5) * (math.Pi / 6)
	fly.pos.x, fly.pos.y = wrapAround(fly.pos.x, fly.pos.y)
	fly.clock += TICK_RATE
	if fly.clock >= TIMER {
		fly.alpha = 1
		fly.clock = 0
	}

}

func (fly *Firefly) Draw(screen *ebiten.Image) {
	a := fly.alpha

	// once glowed, color fades from yellow back to gray
	col := color.RGBA{
		R: uint8(255*a + (1-a)*81),
		G: uint8(255*a + (1-a)*81),
		B: uint8((1 - a) * 81),
		A: uint8(255*a + (1-a)*30),
	}

	ebitenutil.DrawCircle(screen, fly.pos.x, fly.pos.y, radius, col)
}

type Game struct {
	flies [flyCount]Firefly
}

func NewGame() *Game {
	g := &Game{}
	for i := 0; i < flyCount; i++ {
		g.flies[i].Init()
	}
	return g
}

func (g *Game) Update() error {

	for i := 0; i < flyCount; i++ {
		g.flies[i].Update()
		if g.flies[i].alpha == 1 {
			for j := 0; j < flyCount; j++ {
				if i != j {
					g.flies[i].syncUp(&g.flies[j])
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for i := 0; i < flyCount; i++ {
		g.flies[i].Draw(screen)
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WIN_WIDTH, WIN_HEIGHT
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func main() {
	ebiten.SetWindowSize(WIN_WIDTH, WIN_HEIGHT)
	ebiten.SetWindowTitle("Fireflies Synchronization")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
