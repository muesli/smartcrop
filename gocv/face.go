package gocv

import (
	"fmt"
	"image"
	"log"
	"os"

	"gocv.io/x/gocv"
)

type FaceDetector struct {
	FaceDetectionHaarCascadeFilepath string
	DebugMode                        bool
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

	if d.DebugMode == true {
		log.Println("Faces detected:", len(faces))
	}

	for _, face := range faces {
		// Upper left corner of detected face-rectangle
		x := face.Min.X
		y := face.Min.Y

		width := face.Dx()
		height := face.Dy()

		if d.DebugMode == true {
			log.Printf("Face: x: %d y: %d w: %d h: %d\n", x, y, width, height)
		}

		// Mark the rectangle in our [][]uint8 result
		for i := 0; i < width; i++ {
			for j := 0; j < height; j++ {
				res[x+i][y+j] = 255
			}
		}
	}
	return res, nil
}
