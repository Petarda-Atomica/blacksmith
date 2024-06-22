package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path"

	_ "image/png"

	"github.com/gopxl/pixel"
	"github.com/gopxl/pixel/imdraw"
	"github.com/gopxl/pixel/pixelgl"
	"github.com/gopxl/pixel/text"
	"github.com/sqweek/dialog"
	"golang.org/x/image/font/basicfont"
)

var sceneName string = "Home Page"
var openProject modDirectory
var projectLocation string

var win1 *pixelgl.Window
var win2 *pixelgl.Window
var folderpic pixel.Picture
var newfilepic pixel.Picture
var atlas *text.Atlas

var WD string

type mod struct {
	Name        string `json:"name"`
	DownloadURL string `json:"link"`
	Version     string `json:"version"`
}

type modDirectory struct {
	Name    string         `json:"name"`
	Mods    []mod          `json:"mods"`
	Subdirs []modDirectory `json:"folders"`
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
	newfilepic, err = loadPicture(path.Join(WD, "/icons/new-document.png"))
	if err != nil {
		panic(err)
	}

	// Make windows
	cfg := pixelgl.WindowConfig{
		Title:     "Blacksmith",
		Bounds:    pixel.R(0, 0, 1024, 768),
		Resizable: true,
		VSync:     true,
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
	}
	win2, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	for !win1.Closed() && !win2.Closed() {
		switch sceneName {
		case "Home Page":
			homePage()
		}
		win2.Update()
		win1.Update()
	}
}

var openedFileSelectDialog = false

func homePage() {
	// Variables

	// Set windows
	win1.Hide()
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

func main() {
	pixelgl.Run(run)
}
