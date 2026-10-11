package offers

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/HugoSmits86/nativewebp"
	_ "golang.org/x/image/webp"
)

const maxUploadBytes = 8 << 20

type imageVariant struct {
	Size int
	Data []byte
}

func processImage(data []byte) ([]imageVariant, int, int, error) {
	if len(data) == 0 || len(data) > maxUploadBytes {
		return nil, 0, 0, errors.New("image must be between 1 byte and 8MB")
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return nil, 0, 0, errors.New("only JPG, PNG, and WebP images are allowed")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, errors.New("image cannot be decoded")
	}
	width, height := decoded.Bounds().Dx(), decoded.Bounds().Dy()
	if width < 400 || height < 400 {
		return nil, width, height, errors.New("image dimensions must be at least 400x400")
	}
	variants := make([]imageVariant, 0, 3)
	for _, size := range []int{400, 800, 1600} {
		resized := resizeWithin(decoded, size)
		var output bytes.Buffer
		if err := nativewebp.Encode(&output, resized, nil); err != nil {
			return nil, width, height, errors.New("image conversion failed")
		}
		variants = append(variants, imageVariant{Size: size, Data: output.Bytes()})
	}
	return variants, width, height, nil
}

func resizeWithin(source image.Image, maximum int) image.Image {
	bounds := source.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maximum && height <= maximum {
		return source
	}
	ratio := float64(maximum) / float64(width)
	if height > width {
		ratio = float64(maximum) / float64(height)
	}
	targetWidth, targetHeight := int(float64(width)*ratio), int(float64(height)*ratio)
	if targetWidth < 1 {
		targetWidth = 1
	}
	if targetHeight < 1 {
		targetHeight = 1
	}
	destination := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			sourceX := bounds.Min.X + x*width/targetWidth
			sourceY := bounds.Min.Y + y*height/targetHeight
			value := color.NRGBAModel.Convert(source.At(sourceX, sourceY)).(color.NRGBA)
			destination.SetNRGBA(x, y, value)
		}
	}
	return destination
}
