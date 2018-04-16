package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/muesli/smartcrop"
	sclogger "github.com/muesli/smartcrop/logger"
	"github.com/muesli/smartcrop/nfnt"
	//	"github.com/muesli/smartcrop/gocv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please give me an argument")
		os.Exit(1)
	}

	f, _ := os.Open(os.Args[1])
	img, _, _ := image.Decode(f)

	l := sclogger.Logger{
		DebugMode: true,
		Log:       log.New(os.Stderr, "", 0),
	}

	analyzer := smartcrop.NewAnalyzerWithLogger(nfnt.NewDefaultResizer(), l)

	/*
		To replace skin detection with gocv-based face detection:

		analyzer.SetDetectors([]smartcrop.Detector{
			&gocv.FaceDetector{"./cascade.xml", true},
			&smartcrop.SaturationDetector{},
		})
	*/

	topCrop, _ := analyzer.FindBestCrop(img, 300, 200)

	// The crop will have the requested aspect ratio, but you need to copy/scale it yourself
	fmt.Printf("Top crop: %+v\n", topCrop)
}
