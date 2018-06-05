/*
 * Copyright (c) 2014 Christian Muehlhaeuser
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *	Authors:
 *		Christian Muehlhaeuser <muesli@gmail.com>
 *		Michael Wendland <michael@michiwend.com>
 *		Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>
 */

/*
Package smartcrop implements a content aware image cropping library based on
Jonas Wagner's smartcrop.js https://github.com/jwagner/smartcrop.js
*/
package smartcrop

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// debugImage carries debug output image and has methods for updating and writing it
type DebugImage struct {
	img          *image.RGBA
	colors       []color.RGBA
	nextColorIdx int
}

func NewDebugImage(bounds image.Rectangle) *DebugImage {
	di := DebugImage{}

	// Set up the actual image
	di.img = image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			di.img.Set(x, y, color.Black)
		}
	}

	// Set up an array of colors used for debug outputs
	di.colors = []color.RGBA{
		{0, 255, 0, 255},   // default edges
		{255, 0, 0, 255},   // default skin
		{0, 0, 255, 255},   // default saturation
		{255, 128, 0, 255}, // a few extra...
		{128, 0, 128, 255},
		{64, 255, 255, 255},
		{255, 64, 255, 255},
		{255, 255, 64, 255},
		{255, 255, 255, 255},
	}
	di.nextColorIdx = 0
	return &di
}

func (di *DebugImage) popNextColor() color.RGBA {
	c := di.colors[di.nextColorIdx]
	di.nextColorIdx++

	// Wrap around if necessary (if someone ever implements and sets a tenth detector)
	if di.nextColorIdx >= len(di.colors) {
		di.nextColorIdx = 0
	}
	return c
}

func scaledColorComponent(factor uint8, oldComponent uint8, newComponent uint8) uint8 {
	if factor < 1 {
		return oldComponent
	}

	return uint8(bounds(float64(factor) / 255.0 * float64(newComponent)))
}

func (di *DebugImage) AddDetected(d [][]uint8) {
	baseColor := di.popNextColor()

	minX := di.img.Bounds().Min.X
	minY := di.img.Bounds().Min.Y

	maxX := di.img.Bounds().Max.X
	maxY := di.img.Bounds().Max.Y
	if maxX > len(d) {
		maxX = len(d)
	}
	if maxY > len(d[0]) {
		maxY = len(d[0])
	}

	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			if d[x][y] > 0 {
				c := di.img.RGBAAt(x, y)
				nc := color.RGBA{}
				nc.R = scaledColorComponent(d[x][y], c.R, baseColor.R)
				nc.G = scaledColorComponent(d[x][y], c.G, baseColor.G)
				nc.B = scaledColorComponent(d[x][y], c.B, baseColor.B)
				nc.A = 255

				di.img.SetRGBA(x, y, nc)
			}
		}
	}
}

func (di *DebugImage) DebugOutput(debugType string) {
	writeImage("png", di.img, "./smartcrop_"+debugType+".png")
}

func writeImage(imgtype string, img image.Image, name string) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		panic(err)
	}

	switch imgtype {
	case "png":
		return writeImageToPng(img, name)
	case "jpeg":
		return writeImageToJpeg(img, name)
	}

	return errors.New("Unknown image type")
}

func writeImageToJpeg(img image.Image, name string) error {
	fso, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fso.Close()

	return jpeg.Encode(fso, img, &jpeg.Options{Quality: 100})
}

func writeImageToPng(img image.Image, name string) error {
	fso, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fso.Close()

	return png.Encode(fso, img)
}

func (di *DebugImage) DrawDebugCrop(topCrop Crop) {
	o := di.img

	width := o.Bounds().Dx()
	height := o.Bounds().Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := o.At(x, y).RGBA()
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := uint8(b >> 8)

			imp := importance(topCrop, x, y)

			if imp > 0 {
				g8 += imp * 32
			} else if imp < 0 {
				r8 += imp * -64
			}

			nc := color.RGBA{uint8(bounds(r8)), uint8(bounds(g8)), b8, 255}
			o.SetRGBA(x, y, nc)
		}
	}
}
