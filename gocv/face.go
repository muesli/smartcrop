package gocv

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
)

type FaceDetector struct {
	FaceDetectionHaarCascadeFilepath string
	DebugMode                        bool
}

func (d *FaceDetector) Name() string {
	return "face"
}

func (d *FaceDetector) Detect(i *image.RGBA, o *image.RGBA) error {
	// TODO: Fix to use gocv

	if d.FaceDetectionHaarCascadeFilepath == "" {
		return fmt.Errorf("FaceDetector's FaceDetectionHaarCascadeFilepath not specified")
	}

	_, err := os.Stat(d.FaceDetectionHaarCascadeFilepath)
	if err != nil {
		return err
	}
	cascade := opencv.LoadHaarClassifierCascade(d.FaceDetectionHaarCascadeFilepath)
	defer cascade.Release()

	cvImage := opencv.FromImage(i)
	defer cvImage.Release()

	faces := cascade.DetectObjects(cvImage)

	gc := draw2dimg.NewGraphicContext(o)

	if d.DebugMode == true {
		log.Println("Faces detected:", len(faces))
	}

	for _, face := range faces {
		if d.DebugMode == true {
			log.Printf("Face: x: %d y: %d w: %d h: %d\n", face.X(), face.Y(), face.Width(), face.Height())
		}
		draw2dkit.Ellipse(
			gc,
			float64(face.X()+(face.Width()/2)),
			float64(face.Y()+(face.Height()/2)),
			float64(face.Width()/2),
			float64(face.Height())/2)
		gc.SetFillColor(color.RGBA{255, 0, 0, 255})
		gc.Fill()
	}
	return nil
}
