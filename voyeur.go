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
import "github.com/hajimehoshi/ebiten/v2/inpututil"

const VERSION = "v0.0.0"
const PROGRAM = "Voyeur " + VERSION

type Runtime struct {
	skip_frame       bool
	initial_image    string
	active_index     int
	active_directory string
	active_image     *core.Image
	all_files        []File_Info
}

type File_Info struct {
	file_name  string
	unloadable bool
	handle     *core.Image
}

func image_format_supported(test string) bool {
	switch test {
	case
		".jpg", ".jpeg",
		".tiff", ".tif",
		".png",
		".webp":
			return true
	}
	return false
}

func main() {
	run := new(Runtime)

	args := os.Args[1:]
	if len(args) == 0 {
		return
	}

	absolute_path, err := filepath.Abs(args[0])
	if err != nil {
		panic("failed to absolutise path")
	}

	run.initial_image    = absolute_path
	run.active_directory = filepath.Dir(absolute_path)

	run.all_files = get_files_in_folder(run.active_directory)

	img, success := load_any_image(run.initial_image)
	if !success {
		panic("Voyeur exploded early")
	}

	for index := range run.all_files {
		file := &run.all_files[index]
		if file.file_name == run.initial_image {
			run.active_index = index
			file.handle = img
			break
		}
	}

	run.active_image = img

	core.SetWindowTitle(PROGRAM)
	core.SetWindowSize(1920, 1080)
	core.SetWindowResizingMode(core.WindowResizingModeEnabled)
	core.SetFullscreen(true)
	core.SetVsyncEnabled(true)

	err = core.RunGame(run)
	if err != nil {
		panic("Voyeur exploded early")
	}
}

func (run *Runtime) Update() error {
	if inpututil.IsKeyJustPressed(core.KeyEscape) {
		return core.Termination
	}
	if inpututil.IsKeyJustPressed(core.KeyF11) {
		core.SetFullscreen(!core.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(core.KeyLeft) {
		decrement_image(run)
		run.skip_frame = true
		fmt.Println(run.active_index)
	}
	if inpututil.IsKeyJustPressed(core.KeyRight) {
		increment_image(run)
		run.skip_frame = true
		fmt.Println(run.active_index)
	}
	return nil
}

func (run *Runtime) Draw(screen *core.Image) {
	if run.skip_frame {
		run.skip_frame = false
		return
	}
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
			return nil, false
		}
	}

	return core.NewImageFromImage(source_image), true
}

func (run *Runtime) Layout(outside_width, outside_height int) (int, int) {
	size := run.active_image.Bounds().Size()
	return size.X, size.Y
}

func decrement_image(run *Runtime) {
	run.active_index -= 1
	if run.active_index < 0 {
		run.active_index = len(run.all_files) - 1
	}

	new_image := &run.all_files[run.active_index]
	if new_image.unloadable {
		decrement_image(run)
		return
	}

	if new_image.handle == nil {
		img, success := load_any_image(new_image.file_name)
		if !success {
			new_image.unloadable = true
			decrement_image(run)
			return
		}
		new_image.handle = img
	}

	run.active_image = new_image.handle
}

func increment_image(run *Runtime) {
	run.active_index += 1
	if run.active_index > len(run.all_files) - 1 {
		run.active_index = 0
	}

	new_image := &run.all_files[run.active_index]
	if new_image.unloadable {
		increment_image(run)
		return
	}

	if new_image.handle == nil {
		img, success := load_any_image(new_image.file_name)
		if !success {
			new_image.unloadable = true
			increment_image(run)
			return
		}
		new_image.handle = img
	}

	run.active_image = new_image.handle
}

func get_files_in_folder(dir_name string) []File_Info {
	entries, err := os.ReadDir(dir_name)
	if err != nil {
		panic("failed to read directory")
	}

	file_info := make([]File_Info, 0, len(entries))

	for _, file := range entries {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !image_format_supported(filepath.Ext(name)) {
			continue
		}

		var info File_Info
		info.file_name = filepath.Join(dir_name, name)
		file_info = append(file_info, info)
	}

	return file_info
}
