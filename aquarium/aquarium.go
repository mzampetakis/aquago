package aquarium

import (
	"fmt"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

type Direction int
type Speed int
type Scale float64

const (
	bblDirectionPossibility = 0.3
	bblDisappearPossibility = 0.0002
	Left                    = -1
	Up                      = -1
	Right                   = 1
	Down                    = 1

	fishDirectionPossibility = 0.001
	fishAnglePossibility     = 0.01
	maxSpeed                 = Speed(2)
)

type Sprite struct {
	image *ebiten.Image
	name  string
	x     int
	y     int
}

type Bubble struct {
	image     *ebiten.Image
	x         int
	y         int
	direction Direction
	speed     Speed
	scale     Scale
}

type Fish struct {
	image         *ebiten.Image
	name          string
	x             int
	y             int
	direction     Direction
	speed         Speed
	angle         float64
	skew          float64
	skewDirection int
}

type Game struct {
	strokes map[*Stroke]struct{}
	sprites []*Sprite
	bubbles []*Bubble
	fishes  []*Fish
}

var screenWidth, screenHeight int

var bblMux sync.Mutex
var spritesMux sync.Mutex
var fishMux sync.Mutex

var bgImageFilePath string
var bgImagesFolderPath string
var fgImagesFolderPath string

var ebitenImage *ebiten.Image
var bgImage *ebiten.Image
var bblImage *ebiten.Image
var frameCount = 0

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initGame() {
	//Load Bubbles
	img, err := loadImage("assets/bbl.png")
	if err != nil {
		log.Fatal(err)
	}
	bblImage = img

	//Load Background
	img, err = loadImage(bgImageFilePath)
	if err != nil {
		log.Fatal(err)
	}
	bgImage = img
}

func NewGame() *Game {
	// Initialize the sprites.
	sprites := []*Sprite{}
	newGame := &Game{
		strokes: map[*Stroke]struct{}{},
		sprites: sprites,
	}
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			newGame.checkForNewSprites(bgImagesFolderPath)
			newGame.checkForNewFishes(fgImagesFolderPath)
		}
	}()
	// Initialize the game.
	return newGame
}

func (g *Game) checkForNewSprites(bgImagesFolderPath string) {
	dir, _ := ioutil.ReadDir(bgImagesFolderPath)
	for _, d := range dir {
		exist := false
		for _, sprite := range g.sprites {
			if sprite.name == d.Name() {
				exist = true
			}
		}
		if !exist {
			spriteImage, err := loadImage(bgImagesFolderPath + d.Name())
			if err != nil {
				fmt.Println(err)
				continue
			}
			w, h := spriteImage.Size()
			s := &Sprite{
				image: spriteImage,
				name:  d.Name(),
				x:     rand.Intn(screenWidth - w),
				y:     rand.Intn(screenHeight - h),
			}
			g.sprites = append(g.sprites, s)
		}
	}
}

func (g *Game) checkForNewFishes(fgImagesFolderPath string) {
	dir, _ := ioutil.ReadDir(fgImagesFolderPath)
	for _, d := range dir {
		exist := false
		for _, fish := range g.fishes {
			if fish.name == d.Name() {
				exist = true
			}
		}
		if !exist {
			fishImage, err := loadImage(fgImagesFolderPath + d.Name())
			if err != nil {
				fmt.Println(err)
				continue
			}
			w, h := fishImage.Size()
			randDirection := Direction(rand.Intn(2))
			if randDirection == 0 {
				randDirection = -1
			}
			angleDirection := rand.Intn(2)
			if angleDirection == 0 {
				angleDirection = -1
			}
			f := &Fish{
				image:         fishImage,
				name:          d.Name(),
				x:             rand.Intn(screenWidth - w),
				y:             rand.Intn(screenHeight - h),
				direction:     randDirection,
				speed:         Speed(rand.Intn(int(maxSpeed+1)) + 1),
				angle:         rand.Float64() / 2 * float64(angleDirection),
				skew:          float64(0.05),
				skewDirection: 1,
			}
			g.fishes = append(g.fishes, f)
		}
	}
}

func (g *Game) addBubble() {
	bblMux.Lock()
	defer bblMux.Unlock()
	randDirection := Direction(rand.Intn(2))
	if randDirection == 0 {
		randDirection = -1
	}
	scale := Scale(rand.Float64() + 0.5)
	b := &Bubble{
		image:     bblImage,
		x:         100,
		y:         screenHeight,
		direction: randDirection,
		speed:     Speed(rand.Intn(int(maxSpeed+1)) + 1),
		scale:     scale,
	}
	g.bubbles = append(g.bubbles, b)
}

func (g *Game) Update(screen *ebiten.Image) error {
	frameCount++
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s := NewStroke(&MouseStrokeSource{})
		s.SetDraggingObject(g.spriteAt(s.Position()))
		g.strokes[s] = struct{}{}
	}
	for _, id := range inpututil.JustPressedTouchIDs() {
		s := NewStroke(&TouchStrokeSource{id})
		s.SetDraggingObject(g.spriteAt(s.Position()))
		g.strokes[s] = struct{}{}
	}

	for s := range g.strokes {
		g.updateStroke(s)
		if s.IsReleased() {
			delete(g.strokes, s)
		}
	}
	if frameCount%50 == 0 {
		g.addBubble()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	draggingSprites := map[*Sprite]struct{}{}
	for s := range g.strokes {
		if sprite := s.DraggingObject().(*Sprite); sprite != nil {
			draggingSprites[sprite] = struct{}{}
		}
	}
	// Draw Sprites
	for _, s := range g.sprites {
		if _, ok := draggingSprites[s]; ok {
			continue
		}
		s.Draw(screen, 0, 0, 1)
	}
	for s := range g.strokes {
		dx, dy := s.PositionDiff()
		if sprite := s.DraggingObject().(*Sprite); sprite != nil {
			sprite.Draw(screen, dx, dy, 0.5)
		}
	}

	// Draw Bubbles
	for idx, b := range g.bubbles {
		//Determines speed by changing bubble's position according to frame count
		if frameCount%int(b.speed*-1+(maxSpeed+2)) == 0 {
			bblMux.Lock()
			_, h := b.image.Size()
			//remove if bbl reaches top or randomly
			if b.y <= -h || rand.Float64() < bblDisappearPossibility {
				g.bubbles = append(g.bubbles[:idx], g.bubbles[idx+1:]...)
				bblMux.Unlock()
				continue
			}
			b.y--
			// 33% chance to go just straight up
			// looks more smooth in transition
			if rand.Intn(3) == 1 {
				if rand.Float64() < bblDirectionPossibility {
					b.direction *= -1
				}
				b.x += int(b.direction)
			}
			bblMux.Unlock()
		}
		b.Draw(screen, b.scale)
	}

	// Draw Fishes
	for _, f := range g.fishes {
		//Determines speed by changing fishe's position according to frame count
		if frameCount%int(f.speed*-1+(maxSpeed+2)) == 0 {
			fishMux.Lock()
			w, h := f.image.Size()
			if f.y <= 0 || f.y >= screenHeight-h {
				f.angle *= -1
			} else {
				if rand.Float64() < fishAnglePossibility {
					angleDirection := rand.Intn(2)
					if angleDirection == 0 {
						angleDirection = -1
					}
					f.angle += (rand.Float64() - 0.5) / 2
				}
			}
			changeSpeedRand := rand.Float64()
			if changeSpeedRand < 0.05 && f.speed > 1 {
				f.speed--
			}
			if changeSpeedRand > 0.95 && f.speed <= maxSpeed {
				f.speed++
			}
			if f.speed > 0 {
				//change direction if close to left/right or randomly
				if f.x <= 0 || f.x >= screenWidth-w || rand.Float64() < fishDirectionPossibility {
					f.direction = Direction(int(f.direction) * -1)
				}
				f.x += int(f.direction)

				f.y += int(f.angle * 3)
				if frameCount%(int(f.speed*-1+(maxSpeed+2))) == 0 {
					if f.skew >= 0.2 || f.skew <= -0.2 {
						f.skewDirection *= -1
					}
					f.skew += float64(f.skewDirection) * 0.01
				}
			}
			fishMux.Unlock()
		}
		f.Draw(screen)
	}

	// Draw Background
	scaleGeoM := &ebiten.GeoM{}
	scaleGeoM.Scale(2.5, 2.8)
	op := &ebiten.DrawImageOptions{
		GeoM:          *scaleGeoM,
		CompositeMode: 4,
	}
	screen.DrawImage(bgImage, op)

	// Print FPS
	msg := fmt.Sprintf(`FPS: %0.2f`, ebiten.CurrentFPS())
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func Start(bgImageFile string, bgImagesFolder string, fgImagesFolder string) {
	bgImageFilePath = bgImageFile
	bgImagesFolderPath = bgImagesFolder
	fgImagesFolderPath = fgImagesFolder
	ebiten.SetMaxTPS(60)
	initGame()
	ebiten.SetFullscreen(true)
	screenWidth, screenHeight = ebiten.ScreenSizeInFullscreen()
	ebiten.SetWindowTitle("Aquarium")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}

//Draw draws the sprites.
func (s *Sprite) Draw(screen *ebiten.Image, dx, dy int, alpha float64) {
	op := &ebiten.DrawImageOptions{
		CompositeMode: 0,
	}
	op.GeoM.Translate(float64(s.x+dx), float64(s.y+dy))
	op.ColorM.Scale(1, 1, 1, alpha)
	screen.DrawImage(s.image, op)
}

//Draw draws the bubbles.
func (b *Bubble) Draw(screen *ebiten.Image, scale Scale) {
	op := &ebiten.DrawImageOptions{
		CompositeMode: 0,
	}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(b.x), float64(b.y))
	screen.DrawImage(b.image, op)
}

//Draw draws the fishes.
func (f *Fish) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{
		CompositeMode: 0,
	}

	if f.direction == -1 {
		op.GeoM.Scale(-1, 1)
		w, _ := f.image.Size()
		op.GeoM.Translate(float64(w), 0)
	}
	op.GeoM.Skew(0, f.skew)
	w, h := f.image.Size()
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(float64(f.direction) * f.angle)
	op.GeoM.Translate(float64(f.x), float64(f.y))

	screen.DrawImage(f.image, op)
}
