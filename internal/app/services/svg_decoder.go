package services

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iconrepo/internal/logging"
	"image"
	"io"
	"strconv"
)

var name = "svg"
var magicString = "<svg "

func decodeSVG(reader io.Reader) (image.Image, error) {
	// logger := log.WithField("prefix", "SVG decoder::decodeSVG")
	// bytes, readError := io.ReadAll(reader)
	// if readError != nil {
	// 	return nil, fmt.Errorf("failed to read image content: %w", readError)
	// }
	return nil, errors.New("unsupported operation: decode")
}

type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
}

func decodeSVGConfig() func(reader io.Reader) (image.Config, error) {
	return func(reader io.Reader) (image.Config, error) {
		logger := logging.CreateMethodLogger(logging.Get(), "SVG decoder::decodeSVGConfig")

		byteValue, readError := io.ReadAll(reader)
		if readError != nil {
			logger.Error().Err(readError).Msg("failed to read image content")
			return image.Config{}, fmt.Errorf("failed to read image content: %w", readError)
		}
		svg := SVG{}
		xml.Unmarshal(byteValue, &svg)

		width, widthParseError := strconv.Atoi(svg.Width)
		if widthParseError != nil {
			logger.Error().Err(readError).Msg("failed to parse image width")
			return image.Config{}, fmt.Errorf("failed to parse image width: %w", readError)
		}

		height, heightParseError := strconv.Atoi(svg.Height)
		if heightParseError != nil {
			logger.Error().Err(readError).Msg("failed to parse image height")
			return image.Config{}, fmt.Errorf("failed to parse image height: %w", readError)
		}

		return image.Config{
			ColorModel: nil,
			Width:      width,
			Height:     height,
		}, nil
	}
}

func RegisterSVGDecoder() {
	image.RegisterFormat(name, magicString, decodeSVG, decodeSVGConfig())
}
