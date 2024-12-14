package main

import "os"
import "fmt"
import "image"
import "path/filepath"

import lib_png "image/png"
import lib_jpg "image/jpeg"
import lib_tif "golang.org/x/image/tiff"
import lib_web "golang.org/x/image/webp"

import core "github.com/hajimehoshi/ebiten/v2"

const VERSION = "v0.1.0"
const PROGRAM = "Voyeur " + VERSION

func main() {
	run := new(Runtime)

	if len(os.Args[1:]) > 0 {
		run.initial_image = os.Args[1]
	}

	if run.initial_image == "" {
		return
	}

	core.SetWindowTitle(PROGRAM)
	core.SetWindowSize(1920, 1080)
	core.SetWindowResizingMode(core.WindowResizingModeEnabled)
	core.MaximizeWindow()

	img, success := load_any_image(run.initial_image)
	if !success {
		panic("Voyeur exploded early")
	}

	run.active_image = img

	err := core.RunGame(run)
	if err != nil {
		panic("Voyeur exploded early")
	}
}

type Runtime struct {
	initial_image string
	active_image  *core.Image
}

func (run *Runtime) Update() error {
	return nil
}

func (run *Runtime) Draw(screen *core.Image) {
	options := core.DrawImageOptions{
		Filter: core.FilterNearest,
	}
	screen.DrawImage(run.active_image, &options)
}

func load_any_image(file_name string) (*core.Image, bool) {
	var source_image image.Image

	{
		source_file, err := os.Open(file_name)
		if err != nil {
			fmt.Println("failed here")
			return nil, false
		}

		var img_err error

		switch filepath.Ext(file_name) {
		case ".tiff", ".tif":
			source_image, img_err = lib_tif.Decode(source_file)
		case ".jpeg", ".jpg":
			source_image, img_err = lib_jpg.Decode(source_file)
		case ".webp":
			source_image, img_err = lib_web.Decode(source_file)
		case ".png":
			source_image, img_err = lib_png.Decode(source_file)
		}

		if img_err != nil {
			panic(img_err)
			return nil, false
		}
	}

	return core.NewImageFromImage(source_image), true
}

func (run *Runtime) Layout(outside_width, outside_height int) (int, int) {
	size := run.active_image.Bounds().Size()
	return size.X, size.Y
}
