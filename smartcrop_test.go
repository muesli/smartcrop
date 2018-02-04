/*
 * Copyright (c) 2014-2017 Christian Muehlhaeuser
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

package smartcrop

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/muesli/smartcrop/nfnt"
)

var (
	testFile = "./examples/gopher.jpg"
)

// Moved here and unexported to decouple the resizer implementation.
func smartCrop(img image.Image, width, height int) (image.Rectangle, error) {
	analyzer := NewAnalyzer(nfnt.NewDefaultResizer())
	return analyzer.FindBestCrop(img, width, height)
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func TestCrop(t *testing.T) {
	fi, _ := os.Open(testFile)
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		t.Fatal(err)
	}

	topCrop, err := smartCrop(img, 250, 250)
	if err != nil {
		t.Fatal(err)
	}
	expected := image.Rect(464, 24, 719, 279)
	if topCrop != expected {
		t.Fatalf("expected %v, got %v", expected, topCrop)
	}

	sub, ok := img.(SubImager)
	if ok {
		cropImage := sub.SubImage(topCrop)
		// cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
		writeImage("jpeg", cropImage, "./smartcrop.jpg")
	} else {
		t.Error(errors.New("No SubImage support"))
	}
}

func BenchmarkCrop(b *testing.B) {
	fi, err := os.Open(testFile)
	if err != nil {
		b.Fatal(err)
	}
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := smartCrop(img, 250, 250); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEdge(b *testing.B) {
	fi, err := os.Open(testFile)
	if err != nil {
		b.Fatal(err)
	}
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		b.Fatal(err)
	}

	rgbaImg := toRGBA(img)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o := image.NewRGBA(img.Bounds())
		edgeDetect(rgbaImg, o)
	}
}

func BenchmarkImageDir(b *testing.B) {
	files, err := ioutil.ReadDir("./examples")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for _, file := range files {
		if strings.Contains(file.Name(), ".jpg") {
			fi, _ := os.Open("./examples/" + file.Name())
			defer fi.Close()

			img, _, err := image.Decode(fi)
			if err != nil {
				b.Error(err)
				continue
			}

			topCrop, err := smartCrop(img, 220, 220)
			if err != nil {
				b.Error(err)
				continue
			}
			fmt.Printf("Top crop: %+v\n", topCrop)

			sub, ok := img.(SubImager)
			if ok {
				cropImage := sub.SubImage(topCrop)
				// cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
				writeImage("jpeg", cropImage, "/tmp/smartcrop/smartcrop-"+file.Name())
			} else {
				b.Error(errors.New("No SubImage support"))
			}
		}
	}
	// fmt.Println("average time/image:", b.t)
}
