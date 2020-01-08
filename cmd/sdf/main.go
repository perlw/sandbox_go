package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"image"
	"image/png"

	"github.com/perlw/sandbox_go/pkg/sdf"
)

func loadPNG(filepath string) (image.Image, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("file \"%s\" could not be opened: %w", filepath, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("file \"%s\" could not be decoded: %w", filepath, err)
	}

	return img, nil
}

func savePNG(img image.Image, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("file \"%s\" could not be created: %w", filepath, err)
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("file \"%s\" could not be encoded: %w", filepath, err)
	}
	if err := buf.Flush(); err != nil {
		return fmt.Errorf("file \"%s\" could not be flushed: %w", filepath, err)
	}
	return nil
}

func main() {
	var inFile, outFile string
	flag.StringVar(&inFile, "in", "", "the png to calculate the sdf for")
	flag.StringVar(&outFile, "out", "", "the png to output sdf to")
	flag.Parse()

	if inFile == "" || outFile == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	if filepath.Ext(inFile) != ".png" || filepath.Ext(outFile) != ".png" {
		fmt.Println("in/out file should be png")
		os.Exit(-1)
	}

	srcPNG, err := loadPNG(inFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	outPNG, err := sdf.Generate(srcPNG)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	if err := savePNG(outPNG, outFile); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
