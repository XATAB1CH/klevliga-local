package avatar

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
)

const (
	MaxUploadBytes  = 5 << 20
	MaxSide         = 512
	maxSourcePixels = 12_000_000
)

var ErrInvalidImage = errors.New("invalid avatar image")

// Process keeps only decoded pixels. Re-encoding removes EXIF, including GPS
// coordinates and camera identifiers, before the image becomes public.
func Process(data []byte, contentType string) ([]byte, error) {
	if len(data) == 0 || len(data) > MaxUploadBytes {
		return nil, ErrInvalidImage
	}
	var expected string
	switch contentType {
	case "image/jpeg":
		expected = "jpeg"
	case "image/png":
		expected = "png"
	default:
		return nil, ErrInvalidImage
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != expected || config.Width < 1 || config.Height < 1 ||
		config.Width > maxSourcePixels/config.Height {
		return nil, ErrInvalidImage
	}
	src, format, err := image.Decode(bytes.NewReader(data))
	if err != nil || format != expected {
		return nil, ErrInvalidImage
	}
	bounds := src.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	outputSide := min(side, MaxSide)
	startX := bounds.Min.X + (bounds.Dx()-side)/2
	startY := bounds.Min.Y + (bounds.Dy()-side)/2
	dst := image.NewRGBA(image.Rect(0, 0, outputSide, outputSide))
	background := color.RGBA{R: 13, G: 43, B: 55, A: 255}
	for y := 0; y < outputSide; y++ {
		for x := 0; x < outputSide; x++ {
			sourceX := startX + x*side/outputSide
			sourceY := startY + y*side/outputSide
			pixel := color.NRGBAModel.Convert(src.At(sourceX, sourceY)).(color.NRGBA)
			alpha := uint16(pixel.A)
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8((uint16(pixel.R)*alpha + uint16(background.R)*(255-alpha)) / 255),
				G: uint8((uint16(pixel.G)*alpha + uint16(background.G)*(255-alpha)) / 255),
				B: uint8((uint16(pixel.B)*alpha + uint16(background.B)*(255-alpha)) / 255),
				A: 255,
			})
		}
	}
	var result bytes.Buffer
	if err := jpeg.Encode(&result, dst, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
