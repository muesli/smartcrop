smartcrop
=========

[![Latest Release](https://img.shields.io/github/release/muesli/smartcrop.svg)](https://github.com/muesli/smartcrop/releases)
[![Build Status](https://github.com/muesli/smartcrop/workflows/build/badge.svg)](https://github.com/muesli/smartcrop/actions)
[![Coverage Status](https://coveralls.io/repos/github/muesli/smartcrop/badge.svg?branch=master)](https://coveralls.io/github/muesli/smartcrop?branch=master)
[![Go ReportCard](https://goreportcard.com/badge/muesli/smartcrop)](https://goreportcard.com/report/muesli/smartcrop)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/muesli/smartcrop)

smartcrop finds good image crops for arbitrary sizes. It is a pure Go implementation, based on Jonas Wagner's [smartcrop.js](https://github.com/jwagner/smartcrop.js)

![Example](./examples/gopher.jpg)
Image: [https://www.flickr.com/photos/usfwspacific/8182486789](https://www.flickr.com/photos/usfwspacific/8182486789) by Washington Dept of Fish and Wildlife, originally licensed under [CC-BY-2.0](https://creativecommons.org/licenses/by/2.0/) when the image was imported back in September 2014

![Example](./examples/goodtimes.jpg)
Image: [https://www.flickr.com/photos/endogamia/5682480447](https://www.flickr.com/photos/endogamia/5682480447) by Leon F. Cabeiro (N. Feans), licensed under [CC-BY-2.0](https://creativecommons.org/licenses/by/2.0/)

## Installation

Make sure you have a working Go environment (Go 1.12 or higher is required).
See the [install instructions](https://golang.org/doc/install.html).

To install smartcrop, simply run:

    go get github.com/muesli/smartcrop

To compile it from source:

    git clone https://github.com/muesli/smartcrop.git
    cd smartcrop
    go build

## Example
```go
package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
)

func main() {
	f, _ := os.Open("image.png")
	img, _, _ := image.Decode(f)

	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	topCrop, _ := analyzer.FindBestCrop(img, 250, 250)

	// The crop will have the requested aspect ratio, but you need to copy/scale it yourself
	fmt.Printf("Top crop: %+v\n", topCrop)

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	croppedimg := img.(SubImager).SubImage(topCrop)
	// ...
}
```

Also see the test cases in smartcrop_test.go and cli application in cmd/smartcrop/ for further working examples.

## Simple CLI application

    go install github.com/muesli/smartcrop/cmd/smartcrop

    Usage of smartcrop:
      -height int
            crop height
      -input string
            input filename
      -output string
            output filename
      -quality int
            jpeg quality (default 85)
      -resize
            resize after cropping (default true)
      -width int
            crop width

Example:
    smartcrop -input examples/gopher.jpg -output gopher_cropped.jpg -width 300 -height 150

## Sample Data

You can find a bunch of test images for the algorithm [here](https://github.com/muesli/smartcrop-samples).

## Feedback

Got some feedback or suggestions? Please open an issue or drop me a note!

* [Twitter](https://twitter.com/mueslix)
* [The Fediverse](https://mastodon.social/@fribbledom)
