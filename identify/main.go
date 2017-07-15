package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/machinebox/sdk-go/facebox"
)

// IdentifiedFaces includes information about the faces
// identified in an image.
type IdentifiedFaces struct {
	Success    bool           `json:"success"`
	FacesCount int            `json:"facesCount"`
	Faces      []facebox.Face `json:"faces"`
}

func main() {

	// Declare the input and output directory flags.
	inModelDirPtr := flag.String("inModelDir", "", "The directory containing the state file.")
	inImageDirPtr := flag.String("inImageDir", "", "The directory containing input images for tagging.")
	outDirPtr := flag.String("outDir", "", "The output directory")

	// Parse the command line flags.
	flag.Parse()

	// Connect to FaceBox.
	faceboxClient := facebox.New("http://localhost:8080")

	// Open the state file for facebox.
	stateFile, err := os.Open(filepath.Join(*inModelDirPtr, "state.facebox"))
	if err != nil {
		log.Fatal(err)
	}
	defer stateFile.Close()

	// Load the facebox state.
	if err := faceboxClient.PostState(stateFile); err != nil {
		log.Fatal(err)
	}

	// Walk over images in the input directory.
	if err := filepath.Walk(*inImageDirPtr, func(path string, info os.FileInfo, err error) error {

		// Skip any directories.
		if info.IsDir() {
			return nil
		}

		// Open the input image.
		f, err := os.Open(filepath.Join(*inImageDirPtr, info.Name()))
		if err != nil {
			return err
		}
		defer f.Close()

		// Teach FaceBox the input image.
		faces, err := faceboxClient.Check(f)
		if err != nil {
			return err
		}

		// Prepare the output.
		output := IdentifiedFaces{
			Success:    true,
			FacesCount: len(faces),
			Faces:      faces,
		}

		// Prepare an output file name for the tagged information.
		outName := strings.Split(info.Name(), ".")[0]
		outName += ".json"

		// Marshal the output.
		outputData, err := json.Marshal(output)
		if err != nil {
			return err
		}

		// Save the marshalled data to a file.
		if err := ioutil.WriteFile(filepath.Join(*outDirPtr, outName), outputData, 0644); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
