package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
)

// This is an adaptation of the following library into a very simple binary tool:
// https://github.com/Kodeworks/golang-image-ico

type icondir struct {
	reserved  uint16
	imageType uint16
	numImages uint16
}

type icondirentry struct {
	imageWidth   uint8
	imageHeight  uint8
	numColors    uint8
	reserved     uint8
	colorPlanes  uint16
	bitsPerPixel uint16
	sizeInBytes  uint32
	offset       uint32
}

func newIcondir() icondir {
	var id icondir
	id.imageType = 1
	id.numImages = 1
	return id
}

func newIcondirentry() icondirentry {
	var ide icondirentry
	ide.colorPlanes = 1   // windows is supposed to not mind 0 or 1, but other icon files seem to have 1 here
	ide.bitsPerPixel = 32 // can be 24 for bitmap or 24/32 for png. Set to 32 for now
	ide.offset = 22       //6 icondir + 16 icondirentry, next image will be this image size + 16 icondirentry, etc
	return ide
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	inputFileName := flag.String("i", "", "input png file")
	outputFileName := flag.String("o", "", "output ico file")
	flag.Parse()
	if inputFileName == nil || *inputFileName == "" ||
		outputFileName == nil || *outputFileName == "" {
		flag.PrintDefaults()
		return fmt.Errorf("please provide an input and output file")
	}
	f, err := os.Open(*inputFileName)
	if err != nil {
		return fmt.Errorf("could not open input file: %w", err)
	}
	defer f.Close()
	pngFile, err := png.Decode(f)
	if err != nil {
		return fmt.Errorf("could not decode input png file: %w", err)
	}
	f.Close()
	out, err := os.Create(*outputFileName)
	if err != nil {
		return fmt.Errorf("could not open output file: %w", err)
	}
	defer out.Close()
	err = encode(out, pngFile)
	if err != nil {
		return fmt.Errorf("could not encode output ico file: %w", err)
	}
	out.Close()
	return nil
}

func encode(w io.Writer, im image.Image) error {
	b := im.Bounds()
	m := image.NewRGBA(b)
	draw.Draw(m, b, im, b.Min, draw.Src)

	id := newIcondir()
	ide := newIcondirentry()

	pngbb := new(bytes.Buffer)
	pngwriter := bufio.NewWriter(pngbb)
	png.Encode(pngwriter, m)
	pngwriter.Flush()
	ide.sizeInBytes = uint32(len(pngbb.Bytes()))

	bounds := m.Bounds()
	ide.imageWidth = uint8(bounds.Dx())
	ide.imageHeight = uint8(bounds.Dy())
	bb := new(bytes.Buffer)

	var e error
	binary.Write(bb, binary.LittleEndian, id)
	binary.Write(bb, binary.LittleEndian, ide)

	w.Write(bb.Bytes())
	w.Write(pngbb.Bytes())

	return e
}
