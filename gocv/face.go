package gocv

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
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

func (d *FaceDetector) Detect(i *image.RGBA, o *image.RGBA) error {
	if i == nil {
		return fmt.Errorf("i can't be nil")
	}
	if o == nil {
		return fmt.Errorf("o can't be nil")
	}
	if d.FaceDetectionHaarCascadeFilepath == "" {
		return fmt.Errorf("FaceDetector's FaceDetectionHaarCascadeFilepath not specified")
	}

	_, err := os.Stat(d.FaceDetectionHaarCascadeFilepath)
	if err != nil {
		return err
	}

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()
	if !classifier.Load(d.FaceDetectionHaarCascadeFilepath) {
		return fmt.Errorf("FaceDetector failed loading cascade file")
	}

	// image.NRGBA-compatible params
	cvMat := gocv.NewMatFromBytes(i.Rect.Dy(), i.Rect.Dx(), gocv.MatTypeCV8UC4, i.Pix)
	defer cvMat.Close()

	faces := classifier.DetectMultiScale(cvMat)

	gc := draw2dimg.NewGraphicContext(o)

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

		// Draw a filled circle where the face is
		draw2dkit.Ellipse(
			gc,
			float64(x+(width/2)),
			float64(y+(height/2)),
			float64(width/2),
			float64(height)/2)
		gc.SetFillColor(color.RGBA{255, 0, 0, 255})
		gc.Fill()
	}
	return nil
}
