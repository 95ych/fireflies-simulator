package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	WIN_HEIGHT = 1080
	WIN_WIDTH  = 1920
	SCALE      = 1
	TIMER      = 36000
	TICK_RATE  = 100
	SYNC_UP    = 1000
	FADE_RATE  = 0.008
	radius     = 4
)

var dt = 0.15
var flyCount = 900
var proximity float64 = 200
var sync_up_factor float64 = 1
var main_game *Game
var fly_max_speed float64 = 8
var fly_min_speed float64 = 3
var labels = []string{"No. of fireflies", "Time nudge", "Proximity", "Fly Max Speed", "Fly Min Speed"}
var slider_min = []int{100, 1, 50, 1, 1}
var slider_max = []int{1000, 20, 400, 100, 100}
var texts = []*widget.Label{}

type position struct {
	x, y float64
}

type Firefly struct {
	pos   position
	speed float64
	angle float64
	alpha float64
	clock float64
}

func manhattanDist(f1 Firefly, f2 Firefly) float64 {
	return math.Abs(f1.pos.x-f2.pos.x) + math.Abs(f1.pos.y-f2.pos.y)
}

func (fly *Firefly) syncUp(neighborFly *Firefly) {
	if manhattanDist(*fly, *neighborFly) < proximity {
		nudgeFactor := neighborFly.clock / TIMER
		neighborFly.clock = math.Min(neighborFly.clock+nudgeFactor*SYNC_UP*sync_up_factor, TIMER)
	}
}

func (fly *Firefly) Init() {
	fly.pos.x = rand.Float64() * WIN_WIDTH * SCALE
	fly.pos.y = rand.Float64() * WIN_HEIGHT * SCALE
	fly.angle = rand.Float64()*2*math.Pi - math.Pi
	fly.speed = fly_min_speed + rand.Float64()*(fly_max_speed-fly_min_speed)
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

	vector.DrawFilledCircle(screen, float32(fly.pos.x), float32(fly.pos.y), radius, col, false)

}

type Game struct {
	flies []Firefly
	ui    *ebitenui.UI
	state int
}

func updateFireflies(args *widget.SliderChangedEventArgs) {
	texts[0].Label = labels[0] + ": " + fmt.Sprintf("%d", args.Current)
	main_game.state = 1
	flyCount = int(args.Current)
}

func updateTimeNudge(args *widget.SliderChangedEventArgs) {
	texts[1].Label = labels[1] + ": " + fmt.Sprintf("%d", args.Current)
	main_game.state = 1
	sync_up_factor = float64(args.Current)
}

func updateProximity(args *widget.SliderChangedEventArgs) {
	texts[2].Label = labels[2] + ": " + fmt.Sprintf("%d", args.Current)
	main_game.state = 1
	proximity = float64(args.Current)
}

func updateMaxSpeed(args *widget.SliderChangedEventArgs) {
	texts[3].Label = labels[3] + ": " + fmt.Sprintf("%d", args.Current)
	main_game.state = 1
	fly_max_speed = float64(args.Current)

}
func updateMinSpeed(args *widget.SliderChangedEventArgs) {
	texts[4].Label = labels[4] + ": " + fmt.Sprintf("%d", args.Current)
	main_game.state = 1
	fly_min_speed = float64(args.Current)
}

// var labels = []string{"No. of fireflies", "Time nudge", "Proximity", "Fly Max Speed", "Fly Min Speed"}
func NewGame() *Game {
	g := &Game{
		flies: make([]Firefly, flyCount),
	}
	for i := 0; i < flyCount; i++ {
		go g.flies[i].Init()
	}
	// This loads a font and creates a font face.
	ttfFont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("Error Parsing Font", err)
	}
	fontFace := truetype.NewFace(ttfFont, &truetype.Options{
		Size: 22,
	})
	sliderVal := 500

	// This creates a text widget that says "Hello World!"

	// construct a new container that serves as the root of the UI hierarchy
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			StretchHorizontal: true,
		})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(20),
		)))

	// construct a slider
	// add the slider as a child of the container
	// To display the text widget, we have to add it to the root container.
	sliders := []*widget.Slider{}

	for i, parameter := range labels {
		sc := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Spacing(40))),
			widget.ContainerOpts.AutoDisableChildren(),
		)
		rootContainer.AddChild(sc)

		var text *widget.Label

		slider := widget.NewSlider(
			// Set the slider orientation - n/s vs e/w
			widget.SliderOpts.Direction(widget.DirectionHorizontal),
			// Set the minimum and maximum value for the slider
			widget.SliderOpts.MinMax(slider_min[i], slider_max[i]),

			widget.SliderOpts.WidgetOpts(
				// Set the Widget to layout in the center on the screen
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
				// Set the widget's dimensions
				widget.WidgetOpts.MinSize(200, 6),
			),
			widget.SliderOpts.Images(
				// Set the track images
				&widget.SliderTrackImage{
					Idle:  image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
					Hover: image.NewNineSliceColor(color.NRGBA{100, 100, 100, 255}),
				},
				// Set the handle images
				&widget.ButtonImage{
					Idle:    image.NewNineSliceColor(color.NRGBA{255, 255, 0, 190}),
					Hover:   image.NewNineSliceColor(color.NRGBA{255, 255, 00, 255}),
					Pressed: image.NewNineSliceColor(color.NRGBA{255, 255, 100, 255}),
				},
			),
			widget.SliderOpts.FixedHandleSize(6),
			widget.SliderOpts.TrackOffset(0),
			widget.SliderOpts.PageSizeFunc(func() int {
				return 1
			}),
			widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
				switch args.Slider {
				case sliders[0]:
					updateFireflies(args)
				case sliders[1]:
					updateTimeNudge(args)
				case sliders[2]:
					updateProximity(args)
				case sliders[3]:
					updateMaxSpeed(args)
				case sliders[4]:
					updateMinSpeed(args)
				default:
					updateFireflies(args)
					println(i)
				}
			}),
		)
		sliders = append(sliders, slider)
		text = widget.NewLabel(
			widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}))),
			widget.LabelOpts.Text(parameter+": "+fmt.Sprint(sliderVal), fontFace, &widget.LabelColor{
				Idle:     color.White,
				Disabled: color.Black,
			}),
		)
		texts = append(texts, text)
		sc.AddChild(text)
		sc.AddChild(slider)
	}

	g.ui = &ebitenui.UI{
		Container: rootContainer,
	}
	return g
}

func (g *Game) Update() error {
	if g.state == 1 {
		g.flies = make([]Firefly, flyCount)
		g.state = 0
		for i := 0; i < flyCount; i++ {
			go g.flies[i].Init()
		}
		return nil
	}

	for i := 0; i < flyCount; i++ {
		go g.flies[i].Update()
		if g.flies[i].alpha == 1 {
			for j := 0; j < flyCount; j++ {
				if i != j {
					go g.flies[i].syncUp(&g.flies[j])
				}
			}
		}
	}
	g.ui.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.state == 1 {
		g.flies = make([]Firefly, flyCount)
		g.state = 0
		return
	}
	for i := 0; i < flyCount; i++ {
		g.flies[i].Draw(screen)
	}
	g.ui.Draw(screen)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WIN_WIDTH, WIN_HEIGHT
}

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func main() {
	ebiten.SetWindowSize(WIN_WIDTH, WIN_HEIGHT)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Fireflies Synchronization")
	main_game = NewGame()

	if err := ebiten.RunGame(main_game); err != nil {
		log.Fatal(err)
	}
}
