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
	testFile = "./samples/test.png"
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

	topCrop, scaledImg, err := SmartCrop(&img, 300, 300)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Top crop: %+v\n", topCrop)

	sub, ok := scaledImg.(SubImager)
	if ok {
		cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
		WriteImageToJpeg(&cropImage, "/tmp/smartcrop.jpg")

	} else {
		t.Error(errors.New("No SubImage support"))
	}

}

func BenchmarkImageDir(b *testing.B) {

	b.SetParallelism(4)
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

			topCrop, scaledImg, err := SmartCrop(&img, 300, 300)
			if err != nil {
				b.Error(err)
			}

			sub, ok := scaledImg.(SubImager)
			if ok {
				cropImage := sub.SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
				WriteImageToJpeg(&cropImage, "/tmp/smartcrop/smartcrop-"+file.Name())
			} else {
				b.Error(errors.New("No SubImage support"))
			}
		}
	}

	//fmt.Println("average time/image:", b.t

}
