package avatar

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestProcessCropsAndRemovesMetadata(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1200, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 1200; x++ {
			source.Set(x, y, color.RGBA{R: 100, G: 150, B: 180, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, source, nil); err != nil {
		t.Fatal(err)
	}
	// A JPEG APP1 block may carry camera and GPS metadata. It must disappear.
	marker := []byte("Exif-GPS-private-marker")
	segment := append([]byte{0xff, 0xe1, 0, byte(len(marker) + 2)}, marker...)
	input := append(append([]byte{}, encoded.Bytes()[:2]...), append(segment, encoded.Bytes()[2:]...)...)
	output, err := Process(input, "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(output, marker) {
		t.Fatal("EXIF data survived processing")
	}
	result, err := jpeg.Decode(bytes.NewReader(output))
	if err != nil {
		t.Fatal(err)
	}
	if result.Bounds().Dx() != MaxSide || result.Bounds().Dy() != MaxSide {
		t.Fatalf("unexpected avatar dimensions: %v", result.Bounds())
	}
}

func TestProcessRejectsInvalidOrMislabeledImage(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 16, 16))); err != nil {
		t.Fatal(err)
	}
	if _, err := Process(encoded.Bytes(), "image/png"); err != nil {
		t.Fatalf("valid PNG: %v", err)
	}
	for _, tc := range []struct {
		data []byte
		kind string
	}{
		{encoded.Bytes(), "image/jpeg"},
		{[]byte("not an image"), "image/png"},
		{make([]byte, MaxUploadBytes+1), "image/png"},
	} {
		if _, err := Process(tc.data, tc.kind); err == nil {
			t.Fatalf("accepted invalid %s image", tc.kind)
		}
	}
}
