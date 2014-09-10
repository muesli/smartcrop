package smartcrop

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
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
}
