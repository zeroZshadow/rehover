package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/adinfinit/texpack/maxrect"
)

type TexPackerOptions struct {
	MaxBounds   image.Point
	StripPrefix string
}

type TexPacker struct {
	outfile string
	options TexPackerOptions
	images  []imageInfo
}

type imageInfo struct {
	Image  image.Image
	Path   string
	Coords Rect
}

func NewTexPacker(outfile string, options TexPackerOptions) *TexPacker {
	return &TexPacker{
		outfile: outfile,
		options: options,
	}
}

// Add specifies a new texture to be packed.
func (packer *TexPacker) Add(texpath string) error {
	input, err := os.Open(texpath)
	checkErr(err, "Cannot open input file: "+texpath)
	defer input.Close()

	img, fmt, err := image.Decode(input)
	if err != nil {
		return err
	}
	if fmt != "png" && fmt != "jpeg" && fmt != "gif" {
		return errors.New("Unsupported input format: " + fmt)
	}
	// Strip path prefix
	relpath, err := filepath.Rel(packer.options.StripPrefix, texpath)
	checkErr(err, "Error converting absolute path %s to relative (with base %s)", texpath, packer.options.StripPrefix)
	packer.images = append(packer.images, imageInfo{
		Image: img,
		Path:  relpath,
	})
	return nil
}

// Save packs all the given textures into one and writes the result into `output`
func (packer *TexPacker) Save() error {
	imageOut, headerOut, err := packer.getOutputs()
	if err != nil {
		return err
	}
	defer imageOut.Close()
	defer headerOut.Close()

	outtex, err := packer.pack()
	if err != nil {
		return err
	}
	if err = packer.writeHeader(headerOut); err != nil {
		return err
	}
	return png.Encode(imageOut, outtex)
}

// getOutput opens the output files and returns their handle
func (packer *TexPacker) getOutputs() (imageOut io.WriteCloser, headerOut io.WriteCloser, err error) {
	// Get output writer
	imageOut, err = os.Create(packer.outfile)
	if err != nil {
		return
	}
	headerOut, err = os.Create(packer.outfile + ".atlas")
	if err != nil {
		imageOut.Close()
		return
	}
	return
}

// pack runs the maxrect algorithms on the input textures, then writes the result
// in the output texture, which is returned
func (packer *TexPacker) pack() (image.Image, error) {

	points := getImageSizes(packer.images)

	outsize, bounds, ok := minimizeFit(packer.options.MaxBounds, points)
	if !ok {
		return nil, errors.New("Couldn't pack all images in given bounds!")
	}

	// Create and write packed texture
	outtex := image.NewRGBA(image.Rect(0, 0, outsize.X, outsize.Y))
	for i, imginfo := range packer.images {
		srcbounds := bounds[i]
		r := image.Rectangle{srcbounds.Min, srcbounds.Min.Add(imginfo.Image.Bounds().Size())}

		// Copy whole imginfo.Image to rectangle `r` in outtex
		draw.Draw(outtex, r, imginfo.Image, image.Point{0, 0}, draw.Src)

		packer.images[i].Coords = Rect{
			Start: Vector2{uint16(srcbounds.Min.X), uint16(srcbounds.Min.Y)},
			Size:  Vector2{uint16(srcbounds.Dx()), uint16(srcbounds.Dy())},
		}

		// No need to keep this anymore
		packer.images[i].Image = nil
	}

	return outtex, nil
}

// writeHeader outputs binary metadata on the packed textures to the given Writer.
func (packer *TexPacker) writeHeader(output io.Writer) error {

	nEntries := len(packer.images)

	// Output file format is:
	// Entry Count        [4B]
	// Entry0             [12B]
	// ...

	// Used to check hash collisions
	hashes := make(map[FileHash]string, nEntries)

	// Entry Count
	countbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(countbuf, uint32(nEntries))
	if _, err := output.Write(countbuf); err != nil {
		return err
	}

	// Entries
	for _, imginfo := range packer.images {
		hash := ToFileHash(imginfo.Path)
		if orig, collides := hashes[hash]; collides {
			return fmt.Errorf("Hash conflict detected between the following files:\n    [%8x] %s\n    [%8x] %s", hash, orig, hash, imginfo.Path)
		}
		hashes[hash] = imginfo.Path
		entry := Entry{
			TexPath: hash,
			Coords:  imginfo.Coords,
		}
		if _, err := output.Write(entry.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func getImageSizes(images []imageInfo) []image.Point {
	points := make([]image.Point, len(images))
	for i, imginfo := range images {
		points[i] = imginfo.Image.Bounds().Size()
	}
	return points
}

// Taken from https://github.com/adinfinit/texpack/blob/master/pack/fit.go
func minimizeFit(maxContextSize image.Point, sizes []image.Point) (contextSize image.Point, rects []image.Rectangle, ok bool) {

	try := func(size image.Point) ([]image.Rectangle, bool) {
		context := maxrect.New(size)
		return context.Adds(sizes...)
	}

	contextSize = maxContextSize
	rects, ok = try(contextSize)
	if !ok {
		return
	}

	shrunk, shrinkX, shrinkY := true, true, true
	for shrunk {
		shrunk = false
		if shrinkX {
			trySize := image.Point{contextSize.X - 128, contextSize.Y}
			tryRects, tryOk := try(trySize)
			if tryOk {
				contextSize = trySize
				rects = tryRects
				shrunk = true
			} else {
				shrinkX = false
			}
		}

		if shrinkY {
			trySize := image.Point{contextSize.X, contextSize.Y - 128}
			tryRects, tryOk := try(trySize)
			if tryOk {
				contextSize = trySize
				rects = tryRects
				shrunk = true
			} else {
				shrinkY = false
			}
		}
	}

	return
}
