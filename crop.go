/*
 * Copyright (c) 2014 Christian Muehlhaeuser
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *	Authors:
 *		Christian Muehlhaeuser <muesli@gmail.com>
 *		Michael Wendland <michael@michiwend.com>
 */

/*
Package smartcrop implements a content aware image cropping library based on
Jonas Wagner's smartcrop.js https://github.com/jwagner/smartcrop.js
*/
package smartcrop

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"sync"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/nfnt/resize"
)

var skinColor = [3]float64{0.78, 0.57, 0.44}

const (
	detailWeight            = 0.2
	skinBias                = 0.9
	skinBrightnessMin       = 0.2
	skinBrightnessMax       = 1.0
	skinThreshold           = 0.8
	skinWeight              = 1.8
	saturationBrightnessMin = 0.05
	saturationBrightnessMax = 0.9
	saturationThreshold     = 0.4
	saturationBias          = 0.2
	saturationWeight        = 0.3
	scoreDownSample         = 8
	// step * minscale rounded down to the next power of two should be good
	step              = 8
	scaleStep         = 0.1
	minScale          = 0.9
	maxScale          = 1.0
	edgeRadius        = 0.4
	edgeWeight        = -20.0
	outsideImportance = -0.5
	ruleOfThirds      = true
	prescale          = true
	prescaleMin       = 400.00
)

// Score contains values that classify matches
type Score struct {
	Detail     float64
	Saturation float64
	Skin       float64
	Total      float64
}

// Crop contains results
type Crop struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (crop *Crop) importance(x, y int) float64 {
	if crop.X > x || x >= crop.X+crop.Width || crop.Y > y || y >= crop.Y+crop.Height {
		return outsideImportance
	}

	xf := float64(x-crop.X) / float64(crop.Width)
	yf := float64(y-crop.Y) / float64(crop.Height)

	px := math.Abs(0.5-xf) * 2.0
	py := math.Abs(0.5-yf) * 2.0

	dx := math.Max(px-1.0+edgeRadius, 0.0)
	dy := math.Max(py-1.0+edgeRadius, 0.0)
	d := (dx*dx + dy*dy) * edgeWeight

	s := 1.41 - math.Sqrt(px*px+py*py)
	if ruleOfThirds {
		s += (math.Max(0.0, s+d+0.5) * 1.2) * (thirds(px) + thirds(py))
	}

	return s + d
}

func (crop *Crop) score(i *image.Image) Score {
	height := (*i).Bounds().Size().Y
	width := (*i).Bounds().Size().X
	score := Score{}

	// same loops but with downsampling
	for y := 0; y <= height-scoreDownSample; y += scoreDownSample {
		for x := 0; x <= width-scoreDownSample; x += scoreDownSample {

			//for y := 0; y < height; y++ {
			//for x := 0; x < width; x++ {

			r, g, b, _ := (*i).At(x, y).RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			imp := crop.importance(int(x), int(y))
			det := g8 / 255.0

			score.Skin += r8 / 255.0 * (det + skinBias) * imp
			score.Detail += det * imp
			score.Saturation += b8 / 255.0 * (det + saturationBias) * imp
		}
	}

	score.Total = (score.Detail*detailWeight + score.Skin*skinWeight + score.Saturation*saturationWeight) /
		float64(crop.Width) / float64(crop.Height)

	return score
}

func (crop *Crop) drawDebugCrop(o *image.Image) {
	w := (*o).Bounds().Size().X
	h := (*o).Bounds().Size().Y

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {

			r, g, b, _ := (*o).At(x, y).RGBA()
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := uint8(b >> 8)

			imp := crop.importance(x, y)

			if imp > 0 {
				g8 += imp * 32
			} else if imp < 0 {
				r8 += imp * -64
			}

			nc := color.RGBA{uint8(bounds(r8)), uint8(bounds(g8)), b8, 255}
			(*o).(*image.RGBA).Set(x, y, nc)
		}
	}
}

func crops(i image.Image, cropWidth, cropHeight, realMinScale float64) []Crop {
	res := []Crop{}
	width := i.Bounds().Size().X
	height := i.Bounds().Size().Y

	minDimension := math.Min(float64(width), float64(height))
	var cropW, cropH float64

	if cropWidth != 0.0 {
		cropW = cropWidth
	} else {
		cropW = minDimension
	}
	if cropHeight != 0.0 {
		cropH = cropHeight
	} else {
		cropH = minDimension
	}

	for scale := maxScale; scale >= realMinScale; scale -= scaleStep {
		for y := 0; float64(y)+cropH*scale <= float64(height); y += step {
			for x := 0; float64(x)+cropW*scale <= float64(width); x += step {
				res = append(res, Crop{
					X:      x,
					Y:      y,
					Width:  int(cropW * scale),
					Height: int(cropH * scale),
				})
			}
		}
	}

	return res
}

//CropSettings contains options to
//change cropping behaviour
type CropSettings struct {
	FaceDetection                    bool
	FaceDetectionHaarCascadeFilepath string
	InterpolationType                resize.InterpolationFunction
	DebugMode                        bool
}

//Analyzer interface analyzes its struct
//and returns the best possible crop with the given
//width and height
//returns an error if invalid
type Analyzer interface {
	FindBestCrop(img image.Image, width, height int) (Crop, Score, error)
}

type openCVAnalyzer struct {
	cropSettings CropSettings
	logger       logger
	logLevel     LogLevel
	logGuard     sync.RWMutex
}

// SmartCrop applies the smartcrop algorithms on the the given image and returns
// the top crop or an error if somthing went wrong.
func SmartCrop(img image.Image, width, height int) (Crop, Score, error) {
	analyzer := NewAnalyzer()
	return analyzer.FindBestCrop(img, width, height)
}

//NewAnalyzer returns a new analyzer with default settings
func NewAnalyzer() Analyzer {
	faceDetectionHaarCascade := "/usr/share/opencv/haarcascades/haarcascade_frontalface_alt.xml"

	cropSettings := CropSettings{
		FaceDetection:                    true,
		FaceDetectionHaarCascadeFilepath: faceDetectionHaarCascade,
		InterpolationType:                resize.Bicubic,
		DebugMode:                        false,
	}

	return NewAnalyzerWithCropSettings(cropSettings)
}

//NewAnalyzerWithCropSettings returns a new analyzer with the given settings
func NewAnalyzerWithCropSettings(cropSettings CropSettings) Analyzer {
	return &openCVAnalyzer{
		cropSettings: cropSettings,
		logger:       log.New(os.Stderr, "", log.Flags()),
		logLevel:     LogLevelInfo,
	}
}

func (analyzer openCVAnalyzer) FindBestCrop(img image.Image, width, height int) (Crop, Score, error) {
	if width == 0 && height == 0 {
		return Crop{}, Score{}, errors.New("Expect either a height or width")
	}

	scale := math.Min(float64(img.Bounds().Size().X)/float64(width), float64(img.Bounds().Size().Y)/float64(height))

	// resize image for faster processing
	var lowimg image.Image
	var prescalefactor = 1.0

	if prescale {

		//if f := 1.0 / scale / minScale; f < 1.0 {
		//	prescalefactor = f
		//}
		if f := prescaleMin / math.Min(float64(img.Bounds().Size().X), float64(img.Bounds().Size().Y)); f < 1.0 {
			prescalefactor = f
		}
		analyzer.log(LogLevelInfo, "prescale factor %f", prescalefactor)

		lowimg = resize.Resize(
			uint(float64(img.Bounds().Size().X)*prescalefactor),
			0,
			img,
			analyzer.cropSettings.InterpolationType)
	} else {
		lowimg = img
	}

	if analyzer.cropSettings.DebugMode {
		writeImageToPng(&lowimg, "./smartcrop_prescale.png")
	}

	cropWidth, cropHeight := chop(float64(width)*scale*prescalefactor), chop(float64(height)*scale*prescalefactor)
	realMinScale := math.Min(maxScale, math.Max(1.0/scale, minScale))

	analyzer.log(LogLevelInfo, "original resolution: %dx%d\n", img.Bounds().Size().X, img.Bounds().Size().Y)
	analyzer.log(LogLevelInfo, "scale: %f, cropw: %f, croph: %f, minscale: %f\n", scale, cropWidth, cropHeight, realMinScale)

	topCrop, topScore, err := analyzer.analyse(lowimg, cropWidth, cropHeight, realMinScale)
	if err != nil {
		return topCrop, topScore, err
	}

	if prescale == true {
		topCrop.X = int(chop(float64(topCrop.X) / prescalefactor))
		topCrop.Y = int(chop(float64(topCrop.Y) / prescalefactor))
		topCrop.Width = int(chop(float64(topCrop.Width) / prescalefactor))
		topCrop.Height = int(chop(float64(topCrop.Height) / prescalefactor))
	}

	return topCrop, topScore, nil
}

func (analyzer openCVAnalyzer) analyse(img image.Image, cropWidth, cropHeight, realMinScale float64) (Crop, Score, error) {
	o := image.Image(image.NewRGBA(img.Bounds()))

	analyzer.edgeDetect(img, o)
	debugOutput(analyzer.cropSettings.DebugMode, &o, "edge")

	if analyzer.cropSettings.FaceDetection {
		err := analyzer.faceDetect(img, o)

		if err != nil {
			return Crop{}, Score{}, err
		}

		debugOutput(analyzer.cropSettings.DebugMode, &o, "face")
	} else {
		analyzer.skinDetect(img, o)
		debugOutput(analyzer.cropSettings.DebugMode, &o, "skin")
	}

	analyzer.saturationDetect(img, o)
	debugOutput(analyzer.cropSettings.DebugMode, &o, "saturation")

	var (
		topCrop  Crop
		topScore = Score{Total: -1.0}
	)
	cs := crops(o, cropWidth, cropHeight, realMinScale)

	for _, crop := range cs {
		score := crop.score(&o)
		if score.Total > topScore.Total {
			topCrop = crop
			topScore = score
		}
	}

	if analyzer.cropSettings.DebugMode {
		topCrop.drawDebugCrop(&o)
		debugOutput(true, &o, "final")
	}

	return topCrop, topScore, nil
}

func (analyzer openCVAnalyzer) edgeDetect(i image.Image, o image.Image) {
	w := i.Bounds().Size().X
	h := i.Bounds().Size().Y
	cies := makeCies(i)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var lightness float64

			if x == 0 || x >= w-1 || y == 0 || y >= h-1 {
				//lightness = cie((*i).At(x, y))
				lightness = 0
			} else {
				lightness = cies[y*w+x]*4.0 -
					cies[x+(y-1)*w] -
					cies[x-1+y*w] -
					cies[x+1+y*w] -
					cies[x+(y+1)*w]
			}

			nc := color.RGBA{0, uint8(bounds(lightness)), 0, 255}
			o.(*image.RGBA).Set(x, y, nc)
		}
	}
}

func (analyzer openCVAnalyzer) faceDetect(i image.Image, o image.Image) error {

	_, err := os.Stat(analyzer.cropSettings.FaceDetectionHaarCascadeFilepath)
	if err != nil {
		return err
	}
	cascade := opencv.LoadHaarClassifierCascade(analyzer.cropSettings.FaceDetectionHaarCascadeFilepath)
	defer cascade.Release()

	cvImage := opencv.FromImage(i)
	defer cvImage.Release()

	faces := cascade.DetectObjects(cvImage)

	gc := draw2dimg.NewGraphicContext((o).(*image.RGBA))

	analyzer.log(LogLevelInfo, "Faces detected: %d", len(faces))

	for _, face := range faces {
		analyzer.log(LogLevelInfo, "Face: x: %d y: %d w: %d h: %d\n", face.X(), face.Y(), face.Width(), face.Height())
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

func (analyzer openCVAnalyzer) skinDetect(i image.Image, o image.Image) {

	w := i.Bounds().Size().X
	h := i.Bounds().Size().Y

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			lightness := cie(i.At(x, y)) / 255.0
			skin := skinCol(i.At(x, y))

			if skin > skinThreshold && lightness >= skinBrightnessMin && lightness <= skinBrightnessMax {
				r := (skin - skinThreshold) * (255.0 / (1.0 - skinThreshold))
				_, g, b, _ := o.At(x, y).RGBA()
				nc := color.RGBA{uint8(bounds(r)), uint8(g >> 8), uint8(b >> 8), 255}
				o.(*image.RGBA).Set(x, y, nc)
			} else {
				_, g, b, _ := o.At(x, y).RGBA()
				nc := color.RGBA{0, uint8(g >> 8), uint8(b >> 8), 255}
				o.(*image.RGBA).Set(x, y, nc)
			}
		}
	}
}

func (analyzer openCVAnalyzer) saturationDetect(i image.Image, o image.Image) {

	w := i.Bounds().Size().X
	h := i.Bounds().Size().Y

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			lightness := cie(i.At(x, y)) / 255.0
			saturation := saturation(i.At(x, y))

			if saturation > saturationThreshold && lightness >= saturationBrightnessMin && lightness <= saturationBrightnessMax {
				b := (saturation - saturationThreshold) * (255.0 / (1.0 - saturationThreshold))
				r, g, _, _ := o.At(x, y).RGBA()
				nc := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bounds(b)), 255}
				o.(*image.RGBA).Set(x, y, nc)
			} else {
				r, g, _, _ := o.At(x, y).RGBA()
				nc := color.RGBA{uint8(r >> 8), uint8(g >> 8), 0, 255}
				o.(*image.RGBA).Set(x, y, nc)
			}
		}
	}
}

func (analyzer *openCVAnalyzer) SetLogger(l logger, logLevel LogLevel) {
	analyzer.logGuard.Lock()
	defer analyzer.logGuard.Unlock()

	analyzer.logger = l
	analyzer.logLevel = logLevel
}

func (analyzer *openCVAnalyzer) getLogger() (logger, LogLevel) {
	analyzer.logGuard.RLock()
	defer analyzer.logGuard.RUnlock()

	return analyzer.logger, analyzer.logLevel
}

func (analyzer *openCVAnalyzer) log(lvl LogLevel, line string, args ...interface{}) {
	logger, logLevel := analyzer.getLogger()

	if logger == nil {
		return
	}

	if logLevel > lvl {
		return
	}

	logger.Output(2, fmt.Sprintf("%-4s %s", lvl, fmt.Sprintf(line, args...)))
}

func chop(x float64) float64 {
	if x < 0 {
		return math.Ceil(x)
	}
	return math.Floor(x)
}

func thirds(x float64) float64 {
	x = (math.Mod(x-(1.0/3.0)+1.0, 2.0)*0.5 - 0.5) * 16.0
	return math.Max(1.0-x*x, 0.0)
}

func bounds(l float64) float64 {
	return math.Min(math.Max(l, 0.0), 255)
}

func saturation(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := float64(r >> 8)
	g8 := float64(g >> 8)
	b8 := float64(b >> 8)

	maximum := math.Max(math.Max(r8/255.0, g8/255.0), b8/255.0)
	minimum := math.Min(math.Min(r8/255.0, g8/255.0), b8/255.0)

	if maximum == minimum {
		return 0
	}

	l := (maximum + minimum) / 2.0
	d := maximum - minimum

	if l > 0.5 {
		return d / (2.0 - maximum - minimum)
	}

	return d / (maximum + minimum)
}

func cie(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := float64(r >> 8)
	g8 := float64(g >> 8)
	b8 := float64(b >> 8)

	return 0.5126*b8 + 0.7152*g8 + 0.0722*r8
}

func skinCol(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r8 := float64(r >> 8)
	g8 := float64(g >> 8)
	b8 := float64(b >> 8)

	mag := math.Sqrt(r8*r8 + g8*g8 + b8*b8)
	rd := r8/mag - skinColor[0]
	gd := g8/mag - skinColor[1]
	bd := b8/mag - skinColor[2]

	d := math.Sqrt(rd*rd + gd*gd + bd*bd)
	return 1.0 - d
}

func makeCies(img image.Image) []float64 {
	w := img.Bounds().Size().X
	h := img.Bounds().Size().Y
	cies := make([]float64, h*w, h*w)
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cies[i] = cie(img.At(x, y))
			i++
		}
	}

	return cies
}
