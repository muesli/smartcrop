// +build !ci

package gocv

import (
	"fmt"
	"image"
	"os"

	"gocv.io/x/gocv"

	sclogger "github.com/muesli/smartcrop/logger"
)

type FaceDetector struct {
	FaceDetectionHaarCascadeFilepath string
	Logger                           *sclogger.Logger
}

func (d *FaceDetector) Name() string {
	return "face"
}

func (d *FaceDetector) Bias() float64 {
	return 0.9
}

func (d *FaceDetector) Weight() float64 {
	return 1.8
}

func (d *FaceDetector) Detect(img *image.RGBA) ([][]uint8, error) {
	res := make([][]uint8, img.Bounds().Dx())
	for x := range res {
		res[x] = make([]uint8, img.Bounds().Dy())
	}

	if img == nil {
		return res, fmt.Errorf("img can't be nil")
	}
	if d.FaceDetectionHaarCascadeFilepath == "" {
		return res, fmt.Errorf("FaceDetector's FaceDetectionHaarCascadeFilepath not specified")
	}

	_, err := os.Stat(d.FaceDetectionHaarCascadeFilepath)
	if err != nil {
		return res, err
	}

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()
	if !classifier.Load(d.FaceDetectionHaarCascadeFilepath) {
		return res, fmt.Errorf("FaceDetector failed loading cascade file")
	}

	// image.NRGBA-compatible params
	cvMat := gocv.NewMatFromBytes(img.Rect.Dy(), img.Rect.Dx(), gocv.MatTypeCV8UC4, img.Pix)
	defer cvMat.Close()

	faces := classifier.DetectMultiScale(cvMat)

	if d.Logger.DebugMode == true {
		d.Logger.Log.Printf("Number of faces detected: %d\n", len(faces))
	}

	for _, face := range faces {
		// Upper left corner of detected face-rectangle
		x := face.Min.X
		y := face.Min.Y

		width := face.Dx()
		height := face.Dy()

		if d.Logger.DebugMode == true {
			d.Logger.Log.Printf("Face: x: %d y: %d w: %d h: %d\n", x, y, width, height)
		}

		drawAFilledCircle(res, x+(width/2), y+(height/2), width/2)
	}
	return res, nil
}

func drawAFilledCircle(pix [][]uint8, x0, y0, r int) {
	x := r - 1
	y := 0
	dx := 1
	dy := 1
	err := dx - (r << 1)

	for {
		if x < y {
			return
		}

		for i := -x; i <= x; i++ {
			putPixel(pix, x0+i, y0+y)
			putPixel(pix, x0+i, y0-y)
			putPixel(pix, x0+y, y0+i)
			putPixel(pix, x0-y, y0+i)
		}

		if err <= 0 {
			y++
			err += dy
			dy += 2
		} else {
			x--
			dx += 2
			err += dx - (r << 1)
		}
	}
}

func putPixel(pix [][]uint8, x, y int) {
	if x >= len(pix) {
		return
	}
	if y >= len(pix[x]) {
		return
	}
	pix[x][y] = uint8(255)
}
