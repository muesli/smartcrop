package smartcrop

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"testing"
)

var (
	testFile = "/tmp/foo.png"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func TestCrop(t *testing.T) {
	fi, _ := os.Open(testFile)
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		panic(err)
	}

	topCrop, err := SmartCrop(&img, 250, 250)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Top crop: %+v\n", topCrop)

/*	cropImage := img.(SubImager).SubImage(image.Rect(topCrop.X, topCrop.Y, topCrop.Width+topCrop.X, topCrop.Height+topCrop.Y))
	writeImage(&cropImage, "/tmp/foo_topcrop.jpg")*/
}
