// Package ui provides the user interface components.
package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"

	"golang.org/x/image/draw"
)

//go:embed assets/kef.png
var kefLogoPNG []byte

// Icon size for macOS menu bar.
const iconSize = 22

// GenerateVolumeIcon creates the KEF K logo that fills based on volume level.
// volumePercent should be 0-100.
// At 0%: just the outline of the logo
// At 100%: fully filled logo
func GenerateVolumeIcon(volumePercent int) []byte {
	// Clamp volume to valid range
	if volumePercent < 0 {
		volumePercent = 0
	}
	if volumePercent > 100 {
		volumePercent = 100
	}

	// Decode the embedded logo
	srcImg, _, err := image.Decode(bytes.NewReader(kefLogoPNG))
	if err != nil {
		return getDefaultIcon()
	}

	// Create output image at icon size
	img := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))

	// Scale the source logo to icon size
	scaledLogo := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))
	draw.CatmullRom.Scale(scaledLogo, scaledLogo.Bounds(), srcImg, srcImg.Bounds(), draw.Src, nil)

	// Calculate fill threshold (from bottom up)
	fillY := int(float64(iconSize) * (1.0 - float64(volumePercent)/100.0))

	// Colors
	fillColor := color.RGBA{0, 0, 0, 255}         // Black fill
	borderColor := color.RGBA{100, 100, 100, 255} // Gray for outline

	// Process each pixel
	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			r, g, b, a := scaledLogo.At(x, y).RGBA()

			// Check if this pixel is part of the logo (non-transparent and dark)
			isLogo := a > 32768 && (r+g+b)/3 < 32768

			if isLogo {
				if y >= fillY {
					// Below fill line - show filled color
					img.SetRGBA(x, y, fillColor)
				} else {
					// Above fill line - show as outline/border
					if isEdgePixel(scaledLogo, x, y, iconSize) {
						img.SetRGBA(x, y, borderColor)
					}
				}
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return getDefaultIcon()
	}

	return buf.Bytes()
}

// isEdgePixel checks if a pixel is on the edge of the logo.
func isEdgePixel(img *image.RGBA, x, y, size int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < size && ny >= 0 && ny < size {
				_, _, _, na := img.At(nx, ny).RGBA()
				nr, ng, nb, _ := img.At(nx, ny).RGBA()
				neighborIsLogo := na > 32768 && (nr+ng+nb)/3 < 32768
				if !neighborIsLogo {
					return true
				}
			} else {
				return true
			}
		}
	}
	return false
}

// getDefaultIcon returns a minimal valid PNG icon as fallback.
func getDefaultIcon() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00,
		0x1F, 0x15, 0xC4, 0x89,
		0x00, 0x00, 0x00, 0x0A,
		0x49, 0x44, 0x41, 0x54,
		0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01,
		0x0D, 0x0A, 0x2D, 0xB4,
		0x00, 0x00, 0x00, 0x00,
		0x49, 0x45, 0x4E, 0x44,
		0xAE, 0x42, 0x60, 0x82,
	}
}


