package main

import "os"
import "time"
import "image"
import "path/filepath"

import lib_png "image/png"
import lib_jpg "image/jpeg"
import lib_tif "golang.org/x/image/tiff"
import lib_web "golang.org/x/image/webp"
import lib_bmp "golang.org/x/image/bmp"

import core "github.com/hajimehoshi/ebiten/v2"
import coreutil "github.com/hajimehoshi/ebiten/v2/ebitenutil"
import "github.com/hajimehoshi/ebiten/v2/inpututil"

const VERSION = "v0.0.0"
const PROGRAM = "Voyeur 👁️ 👄 👁️ " + VERSION

type Runtime struct {
	skip_frame   bool
	update_timer int

	active_index     int
	active_directory string
	active_image     *core.Image
	scale            float64
	position_x       float64
	position_y       float64

	is_grabbed    bool
	grab_offset_x float64
	grab_offset_y float64

	auto_size bool

	all_files []File_Info
}

type File_Info struct {
	file_name  string
	cant_load  bool
	mod_time   time.Time
	handle     *core.Image
}

func image_format_supported(test string) bool {
	switch test {
	case
		".jpg", ".jpeg",
		".tiff", ".tif",
		".png",
		".bmp",
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
		return
	}

	img, success := load_any_image(absolute_path)
	if !success {
		return
	}

	run.active_directory = filepath.Dir(absolute_path)
	run.all_files        = get_files_in_folder(run.active_directory)

	if len(run.all_files) == 0 {
		return
	}

	for index := range run.all_files {
		file := &run.all_files[index]
		if file.file_name == absolute_path {
			run.active_index = index
			file.handle = img
			break
		}
	}

	run.active_image = img
	run.auto_size = true

	core.SetWindowTitle(PROGRAM)
	core.SetWindowSize(1920, 1080)
	core.SetWindowResizingMode(core.WindowResizingModeEnabled)
	core.SetFullscreen(true)
	core.SetVsyncEnabled(true)
	core.SetWindowSizeLimits(512, 512, 999_999_999, 999_999_999)

	run.scale = calculate_fit(img)

	err = core.RunGame(run)
	if err != nil {
		panic("Voyeur exploded at runtime")
	}
}

func (run *Runtime) Update() error {
	run.update_timer += 1

	if run.update_timer > 120 { // two seconds in ebiten ticks
		run.update_timer = 0

		new_files := get_files_in_folder(run.active_directory)
		if len(new_files) == 0 {
			return core.Termination // exit because there's nothing we can do here
		}

		active_file := run.all_files[run.active_index]
		active_was_deleted := true

		for new_index, new_file := range new_files {
			for _, old_file := range run.all_files {
				if new_file.file_name == old_file.file_name {
					if !new_file.mod_time.After(old_file.mod_time) {
						new_file.handle    = old_file.handle
						new_file.cant_load = old_file.cant_load
					}
					break
				}
			}

			if active_file.file_name == new_file.file_name {
				active_was_deleted = false
				run.active_index = new_index
			}
		}

		if active_was_deleted {
			run.active_index = 0
		}

		run.all_files = new_files

		// we don't update the displayed image because hotloading in real-time
		// generally causes crashes while waiting for files to be written and
		// I didn't want to implement a retry system — for now, you just get
		// your new image by either cycling to the next and back pressing R
	}

	if inpututil.IsKeyJustPressed(core.KeyEscape) {
		return core.Termination
	}
	if inpututil.IsKeyJustPressed(core.KeyZ) {
		run.auto_size = !run.auto_size
	}
	if inpututil.IsKeyJustPressed(core.KeyF) {
		core.SetFullscreen(!core.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(core.KeyR) {
		active_image := &run.all_files[run.active_index]
		img, success := load_any_image(active_image.file_name)
		if !success {
			return core.Termination
		}
		active_image.handle = img
		run.active_image    = img
		run.skip_frame = true
	}

	if inpututil.IsKeyJustPressed(core.KeyLeft) {
		run.skip_frame = decrement_image(run)
		if run.auto_size {
			run.scale = calculate_fit(run.active_image)
			run.position_x = 0
			run.position_y = 0
		}
	}
	if inpututil.IsKeyJustPressed(core.KeyRight) {
		run.skip_frame = increment_image(run)
		if run.auto_size {
			run.scale = calculate_fit(run.active_image)
			run.position_x = 0
			run.position_y = 0
		}
	}

	if inpututil.IsKeyJustPressed(core.Key1) {
		run.scale = 1
		run.position_x = 0
		run.position_y = 0
	}
	if inpututil.IsKeyJustPressed(core.Key2) {
		run.scale = 2
		run.position_x = 0
		run.position_y = 0
	}
	if inpututil.IsKeyJustPressed(core.Key3) {
		run.scale = calculate_fit(run.active_image)
		run.position_x = 0
		run.position_y = 0
	}
	if inpututil.IsKeyJustPressed(core.Key4) {
		run.scale = 4
		run.position_x = 0
		run.position_y = 0
	}

	{
		mx, my := core.CursorPosition()

		if inpututil.IsMouseButtonJustPressed(core.MouseButtonLeft) {
			run.grab_offset_x = float64(mx) - run.position_x
			run.grab_offset_y = float64(my) - run.position_y
			run.is_grabbed    = true
		}
		if inpututil.IsMouseButtonJustReleased(core.MouseButtonLeft) {
			run.is_grabbed    = false
		}
		if run.is_grabbed {
			run.position_x = float64(mx) - run.grab_offset_x
			run.position_y = float64(my) - run.grab_offset_y
		}
	}

	_, wheel_y := core.Wheel()
	if wheel_y != 0 {
		run.scale += wheel_y / 15
	}

	return nil
}

func (run *Runtime) Draw(screen *core.Image) {
	if run.skip_frame {
		run.skip_frame = false
		return
	}

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	w, h   := run.active_image.Bounds().Dx(), run.active_image.Bounds().Dy()

	options := core.DrawImageOptions{}

	options.Filter = core.FilterNearest

	options.GeoM.Translate(float64(-w) / 2, float64(-h) / 2)
	options.GeoM.Scale(run.scale, run.scale)
	options.GeoM.Translate(float64(sw) / 2, float64(sh) / 2)
	options.GeoM.Translate(run.position_x, run.position_y)

	screen.DrawImage(run.active_image, &options)

	coreutil.DebugPrintAt(screen, run.all_files[run.active_index].file_name, 10, 10)
	if !run.auto_size {
		coreutil.DebugPrintAt(screen, "Auto-Sizing Off", 10, 30)
	}
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
		case ".bmp":
			source_image, img_err = lib_bmp.Decode(source_file)
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
	s := core.Monitor().DeviceScaleFactor()
	wx, wy := get_real_window_size()
	xx, yy := int(wx * s), int(wy * s)

	return xx, yy
}

func decrement_image(run *Runtime) bool { // true if first-time load
	run.active_index -= 1
	if run.active_index < 0 {
		run.active_index = len(run.all_files) - 1
	}

	new_image := &run.all_files[run.active_index]
	if new_image.cant_load {
		return decrement_image(run)
	}

	if new_image.handle == nil {
		img, success := load_any_image(new_image.file_name)
		if !success {
			new_image.cant_load = true
			return decrement_image(run)
		}
		new_image.handle = img
		run.active_image = img
		return true
	}

	run.active_image = new_image.handle
	return false
}

func increment_image(run *Runtime) bool { // true if first-time load
	run.active_index += 1
	if run.active_index > len(run.all_files) - 1 {
		run.active_index = 0
	}

	new_image := &run.all_files[run.active_index]
	if new_image.cant_load {
		return increment_image(run)
	}

	if new_image.handle == nil {
		img, success := load_any_image(new_image.file_name)
		if !success {
			new_image.cant_load = true
			return increment_image(run)
		}
		new_image.handle = img
		run.active_image = img
		return true
	}

	run.active_image = new_image.handle
	return false
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

		extra_info, err := file.Info()
		if err != nil {
			continue
		}

		var info File_Info
		info.mod_time  = extra_info.ModTime()
		info.file_name = filepath.Join(dir_name, name)

		file_info = append(file_info, info)
	}

	return file_info
}

func calculate_fit(img *core.Image) float64 {
	s := core.Monitor().DeviceScaleFactor()

	xx, yy := get_real_window_size()
	wx, wy := xx * s, yy * s
	ix, iy := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())

	winr := wx / wy
	canr := ix / iy
	if winr > canr {
		return wy / iy
	}
	return wx / ix
}

func get_real_window_size() (float64, float64) {
	if core.IsFullscreen() {
		xx, yy := core.Monitor().Size()
		return float64(xx), float64(yy)
	}
	xx, yy := core.WindowSize()
	return float64(xx), float64(yy)
}
