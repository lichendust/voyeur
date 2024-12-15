# 👁️ Voyeur

Voyeur is an image viewer that just does the things you want it to —

- Opens images in fullscreen
- Scales images to the window
- Pans and zooms with the mouse
- Cycles through a folder with the arrow keys
- Reloads the folder structure if changes occur

It does not open *slowly*, or at random window sizes or image scales. It does not disable the controls if you're zoomed in. It does not force you to tell it, over and over again, that you want the image to be the largest it can be on your screen.

## Formats

- BMP
- JPG
- PNG
- TIF
- WEBP[^1]

That's it, that's what you get.

## Controls

- `Left`/`Right` to switch images in the open directory
- `Z` to toggle auto-sizing when switching images (preserve scale/position for certain visual comparisons)
- `F` to toggle fullscreen mode
- `R` to reload the current image in place
- `1` to scale 1:1
- `2` to scale 2:1
- `3` to fit image to window
- `4` to scale 4:1
- Scroll to zoom
- Left click and drag to pan

## Notes

- I wrote this thing in like an hour of actual programming time.
- Zooming is wrongly focused on the image centre, not the pointer.  This will be fixed at some point.
- It should be able to sort the files by name or group them by extension. This will be fixed at some point.
- It only loads images once when you first switch to them.  It does not free them until it closes, for fast cycling.  Be careful with large folders of images.
- The whole thing could and should be faster.  It's written in Go with [Ebitengine](https://github.com/hajimehoshi/ebiten) because Go has native support for more image formats than Odin right now, and I wanted a single-binary program.  C, Odin, etc. would have gotten me into `libTIFF` and other such external libraries: not today Satan.

[^1]: But not animated ones.  It hates the animated ones.
