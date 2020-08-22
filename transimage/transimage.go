package transimage

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/anthonynsimon/bild/blur"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/paint"
	"github.com/anthonynsimon/bild/transform"
	"github.com/google/uuid"
)

type coordinates struct {
	x int
	y int
}

func RemoveBG(imageName string, savePath string) string {
	img, err := imgio.Open(imageName)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	img = blur.Gaussian(img, 5.0)
	img = paint.FloodFill(img, image.Point{0, 0}, color.RGBA{0, 0, 0, 0}, 150)

	// Crop the image
	minTransX := -1
	maxTransX := -1
	for x := 0; x < w; x++ {
		foundNonTransparentX := 0
		transparencyLevel := uint32(0)
		for y := 0; y < h; y++ {
			_, _, _, transparencyLevel = img.At(x, y).RGBA()
			if !isTransparent(transparencyLevel) {
				foundNonTransparentX = 1
			}
		}
		if minTransX == -1 && foundNonTransparentX == 1 {
			minTransX = x
		}
		if minTransX != -1 && foundNonTransparentX == 1 {
			maxTransX = x
		}
	}

	minTransY := -1
	maxTransY := -1
	for y := 0; y < h; y++ {
		foundNonTransparentY := 0
		transparencyLevel := uint32(0)
		for x := 0; x < w; x++ {
			_, _, _, transparencyLevel = img.At(x, y).RGBA()
			if !isTransparent(transparencyLevel) {
				foundNonTransparentY = 1
			}
		}
		if minTransY == -1 && foundNonTransparentY == 1 {
			minTransY = y
		}
		if minTransY != -1 && foundNonTransparentY == 1 {
			maxTransY = y
		}
	}
	if minTransY == -1 || minTransX == -1 || maxTransY == -1 || maxTransX == -1 {
		fmt.Println("Could not crop the image")
		return ""
	}

	img = transform.Crop(img, image.Rect(minTransX, minTransY, maxTransX, maxTransY))

	bounds = img.Bounds()
	croppedW, croppedH := bounds.Max.X, bounds.Max.Y
	resizeW := int(float64(croppedW) / float64(w) * 200)
	resizeH := int(float64(croppedH) / float64(h) * 200)
	img = transform.Resize(img, resizeW, resizeH, transform.Gaussian)

	imgFinalName := uuid.New().String()
	if err := imgio.Save(savePath+imgFinalName+".png", img, imgio.PNGEncoder()); err != nil {
		fmt.Println(err)
		return ""
	}
	return savePath + imgFinalName + ".png"
}

func isTransparent(transparencyLevel uint32) bool {
	if transparencyLevel == 0 {
		return true
	}
	return false
}

func SaveBytesToImageFile(byteArr []byte, path string) error {
	img, _, err := image.Decode(bytes.NewReader(byteArr))
	if err != nil {
		return err
	}
	out, _ := os.Create(path)
	defer out.Close()

	var opts jpeg.Options
	opts.Quality = 100

	err = jpeg.Encode(out, img, &opts)
	if err != nil {
		return err
	}
	return nil
}
