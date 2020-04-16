package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

const (
	tiles = "data/tiles"
	dem   = "data/hgt"
)

const SquareSize int = 3601
const NoData int = -9999
const PI float64 = 3.141592653589793
const EarthRadius float32 = 6371.0

type Point [3]interface{}

type Line [2][2]float64

type SRTMFile struct {
	name     string
	lat, lon float64
	contents []byte
}

func LoadSRTMFile(path string) (file *SRTMFile, err error) {
	_, filename := filepath.Split(path)
	lat, lon, err := FilenameToCoordinates(filename[0:7])

	if err != nil {
		return
	}

	file = &SRTMFile{}
	file.name = filename[:7]
	file.lat = lat
	file.lon = lon

	stats, err := os.Stat(path)

	if err != nil {
		return
	}

	if stats.IsDir() {
		err = errors.New("path is a directory instead of a file: " + path)
		return
	}

	file.contents, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return
}

func (file SRTMFile) IsCovered(latitude, longitude float64) bool {
	var latCovered, lonCovered bool

	if file.lat > 0 {
		latCovered = file.lat <= latitude && file.lat+1 > latitude
	} else {
		latCovered = file.lat >= latitude && file.lat+1 < latitude
	}

	if file.lon > 0 {
		lonCovered = file.lon <= longitude && file.lon+1 > longitude
	} else {
		lonCovered = file.lon >= longitude && file.lon+1 < longitude
	}

	return latCovered && lonCovered
}

func (file SRTMFile) GetAltitude(latitude, longitude float64) (int, error) {
	// Check if coordinates are out of bounds
	if !file.IsCovered(latitude, longitude) {
		return 0, fmt.Errorf("(%f, %f) is out of bounds for file %s", latitude, longitude, file.name)
	}

	// Determinate row and column of file
	row := int((file.lat + 1.0 - latitude) * (float64(SquareSize - 1.0)))
	column := int((longitude - file.lon) * (float64(SquareSize - 1.0)))

	// Get the two bytes and return the elevation value
	index := row*SquareSize + column
	return int(file.contents[index*2])*256 + int(file.contents[index*2+1]), nil
}

type FilenameLengthError struct {
	Filename string
	length   int
}

func (e *FilenameLengthError) Error() string {
	return fmt.Sprintf("Filename %s has invalid length! (%d instead of 7)", e.Filename, e.length)
}

// FilenameToCoordinates Returns the coordinates from the starting point of a .hgt filename
func FilenameToCoordinates(filename string) (latitude, longitude float64, err error) {
	if len(filename) != 7 {
		return 0.0, 0.0, &FilenameLengthError{Filename: filename, length: len(filename)}
	}

	latitude, err = strconv.ParseFloat(filename[1:3], 64)

	if err != nil {
		return 0.0, 0.0, err
	}

	if filename[0] == 'S' { // Make negative if south
		latitude *= -1
	}

	longitude, err = strconv.ParseFloat(filename[4:7], 64)

	if err != nil {
		return 0.0, 0.0, err
	}

	if filename[3] == 'W' { // Make negative if west
		longitude *= -1
	}

	return
}

func Radians(deg float64) float64 {
	return deg * (PI / 180)
}

func Degrees(rad float64) float64 {
	return rad * (180 / PI)
}

func MakeRange(start, end, step float64) []float64 {
	var res []float64

	for i := start; i <= end; i += step {
		res = append(res, i)
	}

	return res
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lon1 - lon2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344

	return dist
}

func CoordsToFilename(latitude, longitude float64) string {
	var NorthOrSouth, EastOrWest string

	if latitude >= 0 {
		NorthOrSouth = "N"
	} else {
		NorthOrSouth = "S"
	}

	if longitude >= 0 {
		EastOrWest = "E"
	} else {
		EastOrWest = "W"
	}

	return fmt.Sprintf("%s%d%s%03d", NorthOrSouth, int(latitude), EastOrWest, int(longitude))
}
