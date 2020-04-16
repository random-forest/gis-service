package main

import (
	"log"
	"os"
)

type Tuple [2]interface{}

func Range(start, end, step int) []int {
	var out []int

	for i := start; i <= end; i += step {
		out = append(out, i)
	}

	return out
}

func FRange(start, end, step float64) []float64 {
	var out []float64

	for i := start; i <= end; i += step {
		out = append(out, i)
	}

	return out
}

func Zip(lists ...[]float64) func() []float64 {
	zip := make([]float64, len(lists))
	i := 0

	return func() []float64 {
		for j := range lists {
			if i >= len(lists[j]) {
				return nil
			}
			zip[j] = lists[j][i]
		}
		i++

		return zip
	}
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func Check(e error) {
	if e != nil {
		log.Println(e)
	}
}
