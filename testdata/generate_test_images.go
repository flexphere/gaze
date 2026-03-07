//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
)

func main() {
	// 100x100 PNG with gradient
	createPNG("testdata/test_100x100.png", 100, 100)

	// 1920x1080 PNG
	createPNG("testdata/test_1920x1080.png", 1920, 1080)

	// JPEG
	createJPEG("testdata/test_200x150.jpg", 200, 150)

	// GIF
	createGIF("testdata/test_50x50.gif", 50, 50)
}

func createPNG(path string, w, h int) {
	img := createGradient(w, h)
	f, _ := os.Create(path)
	defer f.Close()
	_ = png.Encode(f, img)
}

func createJPEG(path string, w, h int) {
	img := createGradient(w, h)
	f, _ := os.Create(path)
	defer f.Close()
	_ = jpeg.Encode(f, img, nil)
}

func createGIF(path string, w, h int) {
	img := createGradient(w, h)
	f, _ := os.Create(path)
	defer f.Close()
	_ = gif.Encode(f, img, nil)
}

func createGradient(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			r := uint8(x * 255 / w)
			g := uint8(y * 255 / h)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}
