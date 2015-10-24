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
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/nfnt/resize"
)

var skinColor = [3]float64{0.78, 0.57, 0.44}

const (
	detailWeight             = 0.2
	faceDetectionHaarCascade = "./files/haarcascade_frontalface_alt.xml"
	//skinBias          = 0.01
	useFaceDetection  = true // if true, opencv face detection is used instead of skin detection.
	skinBias          = 0.9
	skinBrightnessMin = 0.2
	skinBrightnessMax = 1.0
	skinThreshold     = 0.8
	//skinWeight              = 1.8
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
	debug             = true
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
	Score  Score
}

//Analyzer interface analyzes its struct
//and returns the best possible crop with the given
//width and height
//returns an error if invalid
type Analyzer interface {
	FindBestCrop(img image.Image, width, height int) (Crop, error)
}

type openCVAnalyzer struct {
	cropSettings CropSettings
}

func (o openCVAnalyzer) FindBestCrop(img image.Image, width, height int) (Crop, error) {
	if width == 0 && height == 0 {
		return Crop{}, errors.New("Expect either a height or width")
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
		log.Println(prescalefactor)

		lowimg = resize.Resize(
			uint(float64(img.Bounds().Size().X)*prescalefactor),
			0,
			img,
			o.cropSettings.InterpolationType)
	} else {
		lowimg = img
	}

	if debug {
		writeImageToPng(&lowimg, "./smartcrop_prescale.png")
	}

	cropWidth, cropHeight := chop(float64(width)*scale*prescalefactor), chop(float64(height)*scale*prescalefactor)
	realMinScale := math.Min(maxScale, math.Max(1.0/scale, minScale))

	log.Printf("original resolution: %dx%d\n", img.Bounds().Size().X, img.Bounds().Size().Y)
	log.Printf("scale: %f, cropw: %f, croph: %f, minscale: %f\n", scale, cropWidth, cropHeight, realMinScale)

	topCrop, err := analyse(lowimg, cropWidth, cropHeight, realMinScale)
	if err != nil {
		return topCrop, err
	}

	if prescale == true {
		topCrop.X = int(chop(float64(topCrop.X) / prescalefactor))
		topCrop.Y = int(chop(float64(topCrop.Y) / prescalefactor))
		topCrop.Width = int(chop(float64(topCrop.Width) / prescalefactor))
		topCrop.Height = int(chop(float64(topCrop.Height) / prescalefactor))
	}

	return topCrop, nil
}

//CropSettings contains options to
//change cropping behaviour
type CropSettings struct {
	FaceDetectionHaarCascadeFilepath string
	InterpolationType                resize.InterpolationFunction
}

//NewAnalyzer returns a new analyzer with default settings
func NewAnalyzer() Analyzer {
	cropSettings := CropSettings{
		FaceDetectionHaarCascadeFilepath: faceDetectionHaarCascade,
		InterpolationType:                resize.Bicubic,
	}

	return &openCVAnalyzer{cropSettings: cropSettings}
}

//NewAnalyzerWithCropSettings returns a new analyzer with the given settings
func NewAnalyzerWithCropSettings(cropSettings CropSettings) Analyzer {
	return &openCVAnalyzer{cropSettings: cropSettings}
}

// SmartCrop applies the smartcrop algorithms on the the given image and returns
// the top crop or an error if somthing went wrong.
func SmartCrop(img image.Image, width, height int) (Crop, error) {
	analyzer := NewAnalyzer()
	return analyzer.FindBestCrop(img, width, height)
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

func importance(crop *Crop, x, y int) float64 {
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

func score(output *image.Image, crop *Crop) Score {
	height := (*output).Bounds().Size().Y
	width := (*output).Bounds().Size().X
	score := Score{}

	// same loops but with downsampling
	for y := 0; y <= height-scoreDownSample; y += scoreDownSample {
		for x := 0; x <= width-scoreDownSample; x += scoreDownSample {

			//for y := 0; y < height; y++ {
			//for x := 0; x < width; x++ {

			r, g, b, _ := (*output).At(x, y).RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			imp := importance(crop, int(x), int(y))
			det := g8 / 255.0

			score.Skin += r8 / 255.0 * (det + skinBias) * imp
			score.Detail += det * imp
			score.Saturation += b8 / 255.0 * (det + saturationBias) * imp
		}
	}

	score.Total = (score.Detail*detailWeight + score.Skin*skinWeight + score.Saturation*saturationWeight) / float64(crop.Width) / float64(crop.Height)
	return score
}

func debugOutput(img *image.Image, debugType string) {
	if debug {
		writeImageToPng(img, "./smartcrop_"+debugType+".png")
	}
}

func writeImageToJpeg(img *image.Image, name string) {
	fso, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer fso.Close()

	jpeg.Encode(fso, (*img), &jpeg.Options{Quality: 100})
}

func writeImageToPng(img *image.Image, name string) {
	fso, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer fso.Close()

	png.Encode(fso, (*img))
}

func drawDebugCrop(topCrop *Crop, o *image.Image) {
	w := (*o).Bounds().Size().X
	h := (*o).Bounds().Size().Y

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {

			r, g, b, _ := (*o).At(x, y).RGBA()
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := uint8(b >> 8)

			imp := importance(topCrop, x, y)

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

func analyse(img image.Image, cropWidth, cropHeight, realMinScale float64) (Crop, error) {
	o := image.Image(image.NewRGBA(img.Bounds()))

	now := time.Now()
	edgeDetect(img, o)
	log.Println("Time elapsed edge:", time.Since(now))
	debugOutput(&o, "edge")

	now = time.Now()
	if useFaceDetection {
		err := faceDetect(img, o)

		if err != nil {
			return Crop{}, err
		}

		log.Println("Time elapsed face:", time.Since(now))
		debugOutput(&o, "face")
	} else {
		skinDetect(img, o)
		log.Println("Time elapsed skin:", time.Since(now))
		debugOutput(&o, "skin")
	}

	now = time.Now()
	saturationDetect(img, o)
	log.Println("Time elapsed sat:", time.Since(now))
	debugOutput(&o, "saturation")

	now = time.Now()
	var topCrop Crop
	topScore := -1.0
	cs := crops(o, cropWidth, cropHeight, realMinScale)
	log.Println("Time elapsed crops:", time.Since(now), len(cs))

	now = time.Now()
	for _, crop := range cs {
		nowIn := time.Now()
		crop.Score = score(&o, &crop)
		log.Println("Time elapsed single-score:", time.Since(nowIn))
		if crop.Score.Total > topScore {
			topCrop = crop
			topScore = crop.Score.Total
		}
	}
	log.Println("Time elapsed score:", time.Since(now))

	if debug {
		drawDebugCrop(&topCrop, &o)
	}
	debugOutput(&o, "final")

	return topCrop, nil
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

func edgeDetect(i image.Image, o image.Image) {
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

func faceDetect(i image.Image, o image.Image) error {

	cvImage := opencv.FromImage(i)
	_, err := os.Stat(faceDetectionHaarCascade)
	if err != nil {
		return err
	}
	cascade := opencv.LoadHaarClassifierCascade(faceDetectionHaarCascade)
	faces := cascade.DetectObjects(cvImage)

	gc := draw2dimg.NewGraphicContext((o).(*image.RGBA))

	if debug == true {
		log.Println("Faces detected:", len(faces))
	}

	for _, face := range faces {
		if debug == true {
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

func skinDetect(i image.Image, o image.Image) {
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

func saturationDetect(i image.Image, o image.Image) {
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
