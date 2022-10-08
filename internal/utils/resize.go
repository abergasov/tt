package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"image/jpeg"

	jpgresize "github.com/nfnt/resize"
)

func ResizeImage(data []byte, width uint, height uint) ([]byte, error) {
	// decode jpeg into image.Image
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to jped decode: %w", err)
	}

	// if either width or height is 0, it will resize respecting the aspect ratio
	newImage := jpgresize.Resize(width, height, img, jpgresize.Lanczos3)

	newData := bytes.Buffer{}
	if err = jpeg.Encode(bufio.NewWriter(&newData), newImage, nil); err != nil {
		return nil, fmt.Errorf("failed to jpeg encode resized image: %w", err)
	}
	return newData.Bytes(), nil
}
