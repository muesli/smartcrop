package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/svkoskin/smartcrop"
	"github.com/svkoskin/smartcrop/nfnt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please give me an argument")
		os.Exit(1)
	}

	f, _ := os.Open(os.Args[1])
	img, _, _ := image.Decode(f)

	l := smartcrop.Logger{
		DebugMode: true,
		Log:       log.New(os.Stderr, "", 0),
	}

	analyzer := smartcrop.NewAnalyzerWithLogger(nfnt.NewDefaultResizer(), l)
	topCrop, _ := analyzer.FindBestCrop(img, 300, 200)

	// The crop will have the requested aspect ratio, but you need to copy/scale it yourself
	fmt.Printf("Top crop: %+v\n", topCrop)
}
