package main

import (
	"encoding/json"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	inFacesDir  = "/pfs/identify"
	inImagesDir = "/pfs/unidentified"
	inTagsDir   = "/pfs/tags"
	outDir      = "/pfs/out"
)

// IdentifiedFaces includes information about the faces
// identified in an image.
type IdentifiedFaces struct {
	Success    bool   `json:"success"`
	FacesCount int    `json:"facesCount"`
	Faces      []Face `json:"faces"`
}

// Face includes information about where a face is in
// an image.
type Face struct {
	Rect    FaceRectangle `json:"rect"`
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Matched bool          `json:"matched"`
}

// FaceRectangle includes the dimensions of a detected face.
type FaceRectangle struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func main() {

	// Walk over files in the input.
	if err := filepath.Walk(inFacesDir, func(path string, info os.FileInfo, err error) error {

		// Skip any directories.
		if info.IsDir() {
			return nil
		}

		// Decode the identified face information.
		f, err := ioutil.ReadFile(filepath.Join(inFacesDir, info.Name()))
		if err != nil {
			return err
		}

		var identifiedFaces IdentifiedFaces
		if err := json.Unmarshal(f, &identifiedFaces); err != nil {
			return err
		}

		// Import the background image to be tagged.
		imageName := strings.Split(info.Name(), ".")[0] + ".jpg"
		backgroundImage, err := os.Open(filepath.Join(inImagesDir, imageName))
		if err != nil {
			return err
		}
		defer backgroundImage.Close()

		// Decode the image to be tagged.
		background, _, err := image.Decode(backgroundImage)
		if err != nil {
			return err
		}

		// Initialize the canvas.
		canvas := image.NewRGBA(background.Bounds())
		draw.Draw(canvas, canvas.Bounds(), background, image.ZP, draw.Src)

		// Loop over identified faces.
		for _, face := range identifiedFaces.Faces {

			if face.Matched {

				// Open the appropriate tag image.
				tagImage, err := os.Open(filepath.Join(inTagsDir, face.Name+".jpg"))
				if err != nil {
					return err
				}
				defer tagImage.Close()

				// Decode the tag.
				tag, _, err := image.Decode(tagImage)
				if err != nil {
					return err
				}

				// Create the rectangles for the tag.
				dp := image.Point{face.Rect.Left, face.Rect.Top}
				sr := tag.Bounds()
				r := image.Rectangle{dp, dp.Add(sr.Size())}

				// Tag the image.
				draw.Draw(canvas, r, tag, sr.Min, draw.Src)
			}
		}

		// Output the tagged image.
		outputImage, err := os.Create(filepath.Join(outDir, "tagged_"+imageName))
		if err != nil {
			return err
		}
		defer outputImage.Close()

		if err := jpeg.Encode(outputImage, canvas, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
