smartcrop.go
============

smartcrop implementation in Go

## Installation

Make sure you have a working Go environment. See the [install instructions](http://golang.org/doc/install.html).

To install smartcrop, simply run:

    go get github.com/muesli/smartcrop.go

To compile it from source:

    git clone git://github.com/muesli/smartcrop.go.git
    cd smartcrop.go && go build && go test -v

## Example
```go
package main

import (
	"github.com/muesli/smartcrop.go"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

func main() {
	fi, _ := os.Open("test.png")
	defer fi.Close()

	img, _, err := image.Decode(fi)
	if err != nil {
		panic(err)
	}

	topCrop, err := smartcrop.SmartCrop(&img, 250, 250)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Top crop: %+v\n", topCrop)
}
```

Also see the test-cases in crop_test.go for further working examples.

## Development
API docs can be found [here](http://godoc.org/github.com/muesli/smartcrop.go).

Continous integration: [![Build Status](https://secure.travis-ci.org/muesli/smartcrop.go.png)](http://travis-ci.org/muesli/smartcrop.go)
