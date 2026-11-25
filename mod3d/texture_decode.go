package mod3d

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	imagedraw "golang.org/x/image/draw"

	"github.com/woozymasta/bcn"
)

const MaxTextureFileBytes int64 = 128 * 1024 * 1024

// ConvertTextureToDataURI converts a DDS or supported image file into a
// Base64 data URI for browser display.
func ConvertTextureToDataURI(path string, maxDim uint32) string {
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 || info.Size() > MaxTextureFileBytes {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(path))
	isDDS := ext == ".dds" || (len(data) >= 4 && string(data[:4]) == "DDS ")

	if isDDS {
		if img, ok := decodeDDS(data, maxDim); ok {
			encoded, err := encodePNG(img)
			if err == nil {
				return "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded)
			}
		}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err == nil {
		img = thumbnail(img, maxDim)

		encoded, err := encodePNG(img)
		if err == nil {
			return "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded)
		}
	}

	switch ext {
	case ".png":
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	case ".jpg", ".jpeg":
		return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
	default:
		return ""
	}
}

// decodeDDS decodes a DDS texture using the BCn decoder.
func decodeDDS(data []byte, maxDim uint32) (*image.NRGBA, bool) {
	if len(data) < 128 || string(data[:4]) != "DDS " {
		return nil, false
	}

	img, err := decodeDDSImage(data)
	if err != nil || img == nil {
		return nil, false
	}

	if maxDim > 0 {
		img = thumbnailNRGBA(img, maxDim)
	}

	return img, true
}

func decodeDDSImage(data []byte) (*image.NRGBA, error) {
	_, img, err := bcn.DecodeDDS(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, os.ErrInvalid
	}
	return img, nil
}

// thumbnail scales an image down while preserving its aspect ratio.
func thumbnail(src image.Image, maxDim uint32) image.Image {
	if maxDim == 0 {
		return src
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= int(maxDim) && height <= int(maxDim) {
		return src
	}

	scale := min(
		float64(maxDim)/float64(width),
		float64(maxDim)/float64(height),
	)

	newWidth := max(1, int(float64(width)*scale))
	newHeight := max(1, int(float64(height)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	imagedraw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		src,
		bounds,
		imagedraw.Src,
		nil,
	)

	return dst
}

// thumbnailNRGBA scales an NRGBA image down while preserving its aspect ratio.
func thumbnailNRGBA(src *image.NRGBA, maxDim uint32) *image.NRGBA {
	if maxDim == 0 {
		return src
	}

	width := src.Bounds().Dx()
	height := src.Bounds().Dy()

	if width <= int(maxDim) && height <= int(maxDim) {
		return src
	}

	scale := min(
		float64(maxDim)/float64(width),
		float64(maxDim)/float64(height),
	)

	newWidth := max(1, int(float64(width)*scale))
	newHeight := max(1, int(float64(height)*scale))

	dst := image.NewNRGBA(image.Rect(0, 0, newWidth, newHeight))

	imagedraw.CatmullRom.Scale(
		dst,
		dst.Bounds(),
		src,
		src.Bounds(),
		imagedraw.Src,
		nil,
	)

	return dst
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	encoder := png.Encoder{
		CompressionLevel: png.BestSpeed,
	}

	if err := encoder.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// EncodeJPEG is available if you need JPEG output elsewhere.
func EncodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	err := jpeg.Encode(&buf, img, &jpeg.Options{
		Quality: 90,
	})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
