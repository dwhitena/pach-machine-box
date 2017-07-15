package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/machinebox/sdk-go/facebox"
)

func main() {

	// Declare the input and output directory flags.
	inDirPtr := flag.String("inDir", "", "The directory containing the training data.")
	outDirPtr := flag.String("outDir", "", "The output directory")

	// Parse the command line flags.
	flag.Parse()

	// Connect to FaceBox.
	faceboxClient := facebox.New("http://localhost:8080")

	// Walk over images in the training directory.
	if err := filepath.Walk(*inDirPtr, func(path string, info os.FileInfo, err error) error {

		// Skip any directories.
		if info.IsDir() {
			return nil
		}

		// Open the training image file.
		f, err := os.Open(filepath.Join(*inDirPtr, info.Name()))
		if err != nil {
			return err
		}
		defer f.Close()

		// Extract the name of the person corresponding to the image
		// (from the file name).
		person := strings.Split(info.Name(), ".")[0]
		person = digitPrefix(person)

		// Teach FaceBox the input image.
		if err := faceboxClient.Teach(f, info.Name(), person); err != nil {
			return err
		}

		// Wait for the training.
		time.Sleep(time.Second * 2)

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	// Export the state of facebox.
	state, err := faceboxClient.OpenState()
	if err != nil {
		log.Fatal(err)
	}
	defer state.Close()

	stateData, err := ioutil.ReadAll(state)
	if err != nil {
		log.Fatal(err)
	}

	// Write out the MachineBox state.
	if err := ioutil.WriteFile(filepath.Join(*outDirPtr, "state.facebox"), stateData, 0644); err != nil {
		log.Fatal(err)
	}
}

// digitPrefix returns the characters in a string before the
// first digit.
func digitPrefix(s string) string {
	for i, r := range s {
		if unicode.IsDigit(r) {
			return s[:i]
		}
	}
	return s
}
