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
)

var (
	testFile = "./samples/gopher.jpg"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func TestCrop(t *testing.T) {
	fi, _ := os.Open(testFile)
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		t.Error(err)
	}

	topCrop, err := SmartCrop(&img, 250, 0)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Top crop: %+v\n", topCrop)

	sub, ok := img.(SubImager)
	if ok {
		cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
		writeImageToJpeg(&cropImage, "./smartcrop.jpg")

	} else {
		t.Error(errors.New("No SubImage support"))
	}

}

func BenchmarkImageDir(b *testing.B) {

	b.StopTimer()

	files, err := ioutil.ReadDir("./samples")
	if err != nil {
		b.Error(err)
	}

	b.StartTimer()
	for _, file := range files {
		if strings.Contains(file.Name(), ".jpg") {

			fi, _ := os.Open("./samples/" + file.Name())
			defer fi.Close()

			img, _, err := image.Decode(fi)
			if err != nil {
				b.Error(err)
			}

			topCrop, err := SmartCrop(&img, 900, 500)
			if err != nil {
				b.Error(err)
			}
			fmt.Printf("Top crop: %+v\n", topCrop)

			sub, ok := img.(SubImager)
			//sub, ok := img.(SubImager)
			if ok {
				cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
				writeImageToJpeg(&cropImage, "/tmp/smartcrop/smartcrop-"+file.Name())
			} else {
				b.Error(errors.New("No SubImage support"))
			}
		}
	}
	//fmt.Println("average time/image:", b.t

}
