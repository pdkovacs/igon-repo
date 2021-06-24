package services

import (
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"io"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var name = "svg"
var magicString = "<svg "

func decodeSVG(reader io.Reader) (image.Image, error) {
	// logger := log.WithField("prefix", "SVG decoder::decodeSVG")
	// bytes, readError := io.ReadAll(reader)
	// if readError != nil {
	// 	return nil, fmt.Errorf("failed to read image content: %w", readError)
	// }
	return nil, errors.New("Unsupported operation: decode")
}

type SVG struct {
	XMLName xml.Name `xml:"svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
}

func decodeSVGConfig(reader io.Reader) (image.Config, error) {
	logger := log.WithField("prefix", "SVG decoder::decodeSVGConfig")

	byteValue, readError := io.ReadAll(reader)
	if readError != nil {
		logger.Errorf("failed to read image content: %v", readError)
		return image.Config{}, fmt.Errorf("failed to read image content: %w", readError)
	}
	svg := SVG{}
	xml.Unmarshal(byteValue, &svg)

	width, widthParseError := strconv.Atoi(svg.Width)
	if widthParseError != nil {
		logger.Errorf("failed to parse image width: %v", readError)
		return image.Config{}, fmt.Errorf("failed to parse image width: %w", readError)
	}

	height, heightParseError := strconv.Atoi(svg.Height)
	if heightParseError != nil {
		logger.Errorf("failed to parse image height: %v", readError)
		return image.Config{}, fmt.Errorf("failed to parse image height: %w", readError)
	}

	return image.Config{
		ColorModel: nil,
		Width:      width,
		Height:     height,
	}, nil
}

func registerSVGDecoder() {
	image.RegisterFormat(name, magicString, decodeSVG, decodeSVGConfig)
}
