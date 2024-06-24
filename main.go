package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path"

	_ "image/png"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
	"github.com/gopxl/pixel/pixelgl"
	"github.com/gopxl/pixel/text"
	"github.com/sqweek/dialog"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var sceneName string = "Home Page"
var openProject modDirectory
var projectLocation string

var win1Visible = false
var win2Visible = false

var win1 *pixelgl.Window
var win2 *pixelgl.Window
var folderpic pixel.Picture
var closedfolderpic pixel.Picture
var newfilepic pixel.Picture
var cubepic pixel.Picture
var atlas *text.Atlas

var WD string

type mod struct {
	Name        string `json:"name"`
	URL         string `json:"presentation"`
	DownloadURL string `json:"link"`
	Version     string `json:"version"`
}

type modDirectory struct {
	Name      string         `json:"name"`
	Mods      []mod          `json:"mods"`
	Subdirs   []modDirectory `json:"folders"`
	collapsed bool
}

func RGBA(r uint8, g uint8, b uint8, a uint8) color.Color {
	return color.RGBA{R: r,
		G: g,
		B: b,
		A: a,
	}
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

var err error

func run() {
	// Init vars
	WD, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	// Make a text atlas
	atlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// Read files
	folderpic, err = loadPicture(path.Join(WD, "/icons/folder.png"))
	if err != nil {
		panic(err)
	}
	closedfolderpic, err = loadPicture(path.Join(WD, "/icons/folder-closed.png"))
	if err != nil {
		panic(err)
	}
	newfilepic, err = loadPicture(path.Join(WD, "/icons/new-document.png"))
	if err != nil {
		panic(err)
	}
	cubepic, err = loadPicture(path.Join(WD, "/icons/block.png"))
	if err != nil {
		panic(err)
	}

	// Make windows
	cfg := pixelgl.WindowConfig{
		Title:     "Blacksmith",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
		VSync:     true,
		Invisible: true,
	}
	win1, err = pixelgl.NewWindow(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	if err != nil {
		panic(err)
	}
	cfg = pixelgl.WindowConfig{
		Title:     "Blacksmith",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: false,
		VSync:     true,
		Invisible: true,
	}
	win2, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for !win1.Closed() && !win2.Closed() {
		switch sceneName {
		case "Home Page":
			homePage()
		case "Project Browser":
			projectBrowser()
		}
		if win2Visible {
			win2.Update()
		}
		if win1Visible {
			win1.Update()
		}
	}
}

func drawSidebar() mod {
	// Variables
	screen := win1.Bounds()
	var pressedMod mod

	// Render folders toolbar
	leftBar := imdraw.New(nil)
	leftBar.Color = RGBA(1, 0, 20, 255)
	leftBar.Push(pixel.V(0, 0), pixel.V(0, screen.H()), pixel.V(300, screen.H()), pixel.V(300, 0))
	leftBar.Polygon(0)
	leftBar.Draw(win1)

	// Render folders
	var drawIndex int
	var objectHeight float64
	for _, val := range openProject.Subdirs {
		// Draw folder
		var sprite *pixel.Sprite
		if val.collapsed {
			sprite = pixel.NewSprite(folderpic, folderpic.Bounds())
		} else {
			sprite = pixel.NewSprite(closedfolderpic, closedfolderpic.Bounds())
		}
		objectHeight = drawSidebarItem(val.Name, sprite, screen, drawIndex, sidebarScrollIndex, false)
		drawIndex++

		// Check if collapsed
		if !val.collapsed {
			continue
		}

		// Render subfiles
		for _, vaj := range val.Mods {
			sprite = pixel.NewSprite(cubepic, cubepic.Bounds())
			drawSidebarItem(vaj.Name, sprite, screen, drawIndex, sidebarScrollIndex, true)
			drawIndex++
		}

	}

	// Render non-foldered mods
	for _, val := range openProject.Mods {
		sprite := pixel.NewSprite(cubepic, cubepic.Bounds())
		drawSidebarItem(val.Name, sprite, screen, drawIndex, sidebarScrollIndex, false)
		drawIndex++
	}

	// Handle sidebar scrolling
	if win1.MousePosition().X <= 300 {
		scrollingPower := 7.5
		sidebarScrollIndex += win1.MouseScroll().Y * scrollingPower

		if screen.H()-(float64(drawIndex+1)*objectHeight-sidebarScrollIndex) > 0 {
			sidebarScrollIndex = float64(drawIndex+1)*objectHeight - screen.H()
		}

		if sidebarScrollIndex < 0 || float64(drawIndex+1)*objectHeight < screen.H() {
			sidebarScrollIndex = 0
		}

	}

	// Handle clicking
	if win1.MousePosition().X < 300 && win1.JustPressed(pixelgl.MouseButtonLeft) {
		objectNO := int(math.Floor((screen.H() - win1.MousePosition().Y + sidebarScrollIndex) / objectHeight))

		// Determine what was clicked and take action
		func() {
			for i, val := range openProject.Subdirs {
				if objectNO == 0 {
					openProject.Subdirs[i].collapsed = !openProject.Subdirs[i].collapsed
					return
				}

				if val.collapsed {
					if len(val.Mods) < objectNO {
						objectNO -= len(val.Mods)
					} else {
						pressed := val.Mods[objectNO-1]
						pressedMod = pressed
						return
					}
				}

				objectNO--
			}

			if objectNO >= len(openProject.Mods) {
				return
			}

			pressed := openProject.Mods[objectNO]
			pressedMod = pressed
		}()
	}

	return pressedMod
}

func drawSidebarItem(name string, sprite *pixel.Sprite, screen pixel.Rect, drawIndex int, scrollIndex float64, child bool) float64 {
	// Load icon
	marginTop := 10.0
	marginLeft := 10.0
	spriteScale := 0.055
	additionalMargin := 0.0
	if child {
		additionalMargin += 20.0
	}

	// Draw icon
	sprite.Draw(win1, pixel.IM.Scaled(pixel.V(0, 0), spriteScale).Moved(pixel.V(sprite.Frame().W()*spriteScale/2+marginLeft+additionalMargin, scrollIndex+screen.H()-(sprite.Frame().H()*spriteScale+marginTop)*float64(drawIndex)-marginTop-sprite.Frame().H()*spriteScale/2)))

	// Draw folder title
	txt := text.New(pixel.V(0, 0), atlas)
	txt.Color = colornames.White
	origtextScale := .55
	fmt.Fprint(txt, string(name[0]))
	textScale := (sprite.Frame().H() * spriteScale / txt.Bounds().H()) * origtextScale
	for i, val := range name {
		if i == 0 {
			continue
		}
		if txt.Bounds().W()*textScale > 200 {
			fmt.Fprint(txt, "...")
			break
		}
		fmt.Fprint(txt, string(val))
	}
	txt.Draw(win1, pixel.IM.Scaled(pixel.V(0, 0), textScale).Moved(pixel.V(additionalMargin+2*marginLeft+sprite.Frame().W()*spriteScale, scrollIndex+screen.H()-float64(drawIndex+1)*(marginTop+sprite.Frame().H()*spriteScale))))

	return marginTop + sprite.Frame().H()*spriteScale
}

var openedFileSelectDialog = false

func homePage() {
	// Set windows
	if win1Visible {
		win1.Hide()
		win1Visible = false
	}
	if !win2Visible {
		win2.Show()
		win2Visible = true
	}
	win2.SetBounds(pixel.R(0, 0, 600, 400))

	// Update
	win2.Clear(color.RGBA{R: 6, G: 24, B: 38, A: 255})

	mouse := win2.MousePosition()

	// Draw open file bounds
	imd := imdraw.New(nil)
	imd.Color = RGBA(249, 251, 239, 255)
	if mouse.X > 30 && mouse.X < 255 && mouse.Y > 115 && mouse.Y < 370 {
		imd.Color = RGBA(0, 107, 143, 255)
	}
	imd.EndShape = imdraw.RoundEndShape
	imd.Push(pixel.V(30, 370), pixel.V(255, 370), pixel.V(255, 115), pixel.V(30, 115), pixel.V(30, 370))
	imd.Line(10)
	imd.Draw(win2)

	// Draw new file bounds
	imd.Clear()
	imd.Color = RGBA(249, 251, 239, 255)
	if 600-mouse.X > 30 && 600-mouse.X < 255 && mouse.Y > 115 && mouse.Y < 370 {
		imd.Color = RGBA(0, 107, 143, 255)
	}
	imd.EndShape = imdraw.RoundEndShape
	imd.Push(pixel.V(600-30, 370), pixel.V(600-255, 370), pixel.V(600-255, 115), pixel.V(600-30, 115), pixel.V(600-30, 370))
	imd.Line(10)
	imd.Draw(win2)

	// Draw open file
	sprite := pixel.NewSprite(folderpic, folderpic.Bounds())
	sprite.Draw(win2, pixel.IM.Moved(win2.Bounds().Center()).Scaled(win2.Bounds().Center(), .3).Moved(pixel.V(-300+15+128, 110-30)))
	title := text.New(pixel.V(0, 0), atlas)
	fmt.Fprintln(title, "Open Project")
	title.Draw(win2, pixel.IM.Scaled(pixel.V(0, 0), 2).Moved(pixel.V(138-title.Bounds().W(), 115+40)))

	// Draw new file
	sprite = pixel.NewSprite(newfilepic, newfilepic.Bounds())
	sprite.Draw(win2, pixel.IM.Moved(win2.Bounds().Center()).Scaled(win2.Bounds().Center(), .25).Moved(pixel.V(300-128, 110-30)))
	title = text.New(pixel.V(0, 0), atlas)
	fmt.Fprintln(title, "New Project")
	title.Draw(win2, pixel.IM.Scaled(pixel.V(0, 0), 2).Moved(pixel.V(600-(138+title.Bounds().W()), 115+40)))

	// Check mouse events
	if win2.JustPressed(pixelgl.MouseButtonLeft) && !openedFileSelectDialog {
		// Pressed open file
		if mouse.X > 30 && mouse.X < 255 && mouse.Y > 115 && mouse.Y < 370 {
			openedFileSelectDialog = true
			go func() {
				// Ask user what file to load
				defer func() { openedFileSelectDialog = false }()
				file, err := dialog.File().Load()
				if err != dialog.Cancelled && err != nil {
					panic(err)
				} else if err == dialog.Cancelled {
					log.Println("User cancelled file selection")
					return
				}

				// Read file
				f, err := os.ReadFile(file)
				if err != nil {
					log.Println(err)
					return
				}

				// Load file
				err = json.Unmarshal(f, &openProject)
				if err != nil {
					log.Println(err)
					return
				}

				projectLocation = file
				sceneName = "Project Browser"
			}()
		} else if 600-mouse.X > 30 && 600-mouse.X < 255 && mouse.Y > 115 && mouse.Y < 370 {
			openedFileSelectDialog = true
			go func() {
				defer func() { openedFileSelectDialog = false }()
				file, err := dialog.File().Save()
				if err != dialog.Cancelled && err != nil {
					panic(err)
				} else if err == dialog.Cancelled {
					log.Println("User cancelled file selection")
					return
				}

				// Encode the JSON
				f, err := json.Marshal(openProject)
				if err != nil {
					log.Println(err)
					return
				}

				// Create the project file
				if err := os.WriteFile(file, f, os.ModePerm); err != nil {
					log.Println(err)
					return
				}

				projectLocation = file
				sceneName = "Project Browser"
			}()
		}
	}
}

var sidebarScrollIndex float64

func projectBrowser() {
	// Manage windows
	if !win1Visible {
		win1.Show()
		win1Visible = true
	}
	if win2Visible {
		win2.Hide()
		win2Visible = false
	}

	// Update screen
	win1.Clear(RGBA(6, 24, 38, 255))

	press := drawSidebar()
	if press.Name != "" {
		log.Println("Pressed mod", press.Name)
	}
}

func main() {
	pixelgl.Run(run)
}
