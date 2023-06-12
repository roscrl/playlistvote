package domain

import (
	"fmt"
	"image"
	"sync"

	"github.com/EdlinOrg/prominentcolor"
)

// ProminentFourColorsMosaic returns the four most prominent hex colors in a playlist sampler.
func ProminentFourColorsMosaic(img image.Image) []string {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Split the input image into four quadrants.
	quadWidth := width / 2   //nolint:gomnd
	quadHeight := height / 2 //nolint:gomnd

	quad1 := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(0, 0, quadWidth, quadHeight))

	quad2 := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(quadWidth, 0, width, quadHeight))

	quad3 := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(0, quadHeight, quadWidth, height))

	quad4 := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(quadWidth, quadHeight, width, height))

	var (
		quad1Color, quad2Color, quad3Color, quad4Color []prominentcolor.ColorItem
		wg                                             sync.WaitGroup
	)

	wg.Add(4) //nolint:gomnd

	go func() {
		defer wg.Done()

		quad1Color = getProminentColors(1, quad1)
	}()
	go func() {
		defer wg.Done()

		quad2Color = getProminentColors(1, quad2)
	}()
	go func() {
		defer wg.Done()

		quad3Color = getProminentColors(1, quad3)
	}()
	go func() {
		defer wg.Done()

		quad4Color = getProminentColors(1, quad4)
	}()

	wg.Wait()

	quad1Hex := rgbToHex(quad1Color[0].Color.R, quad1Color[0].Color.G, quad1Color[0].Color.B)
	quad2Hex := rgbToHex(quad2Color[0].Color.R, quad2Color[0].Color.G, quad2Color[0].Color.B)
	quad3Hex := rgbToHex(quad3Color[0].Color.R, quad3Color[0].Color.G, quad3Color[0].Color.B)
	quad4Hex := rgbToHex(quad4Color[0].Color.R, quad4Color[0].Color.G, quad4Color[0].Color.B)

	return []string{quad1Hex, quad2Hex, quad3Hex, quad4Hex}
}

// ProminentFourColors returns the four most prominent hex colors in a user uploaded playlist sampler.
func ProminentFourColors(img image.Image) []string {
	colors := getProminentColors(4, img) //nolint:gomnd

	color1Hex := rgbToHex(colors[0].Color.R, colors[0].Color.G, colors[0].Color.B)
	color2Hex := rgbToHex(colors[1].Color.R, colors[1].Color.G, colors[1].Color.B)
	color3Hex := rgbToHex(colors[2].Color.R, colors[2].Color.G, colors[2].Color.B)
	color4Hex := rgbToHex(colors[3].Color.R, colors[3].Color.G, colors[3].Color.B)

	return []string{color1Hex, color2Hex, color3Hex, color4Hex}
}

func getProminentColors(numColors int, img image.Image) []prominentcolor.ColorItem {
	ignoreColors := []prominentcolor.ColorBackgroundMask{prominentcolor.MaskWhite} // ignore white
	colors, _ := prominentcolor.KmeansWithAll(numColors, img, prominentcolor.ArgumentNoCropping, prominentcolor.DefaultSize, ignoreColors)

	return colors
}

func rgbToHex(r, g, b uint32) string {
	hex := fmt.Sprintf("#%02x%02x%02x", r, g, b)

	return hex
}
