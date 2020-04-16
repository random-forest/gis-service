package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func TileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	imgPath := strings.Split(strings.Replace(r.URL.Path, "/tiles/", "", 1), "/")

	mapstyle := imgPath[0]
	z := imgPath[1]
	x := imgPath[2]
	y := imgPath[3]
	filepath := tiles + "/" + mapstyle + ".db"

	if FileExists(filepath) {
		db, _ := sql.Open("sqlite3", filepath)
		var blob []byte

		sqlStatement := `SELECT data FROM tiles WHERE z=$1 and x=$2 and y=$3;`
		row := db.QueryRow(sqlStatement, z, x, y)

		switch err := row.Scan(&blob); err {
		case sql.ErrNoRows:
			fmt.Println("\nNo rows were returned!")
			w.WriteHeader(404)
		case nil:
			w.WriteHeader(200)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", http.DetectContentType(blob)+"; charset=UTF8")
			w.Write(blob)
		default:
			w.WriteHeader(404)
		}

		defer db.Close()
	} else {
		w.WriteHeader(404)
	}
}

func DemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	demPath := strings.Split(strings.Replace(r.URL.Path, "/height/", "", 1), "/")

	lat, _ := strconv.ParseFloat(demPath[0], 64)
	lon, _ := strconv.ParseFloat(demPath[1], 64)
	filepath := dem + "/" + CoordsToFilename(lat, lon) + ".hgt"

	if FileExists(filepath) {
		file, err := LoadSRTMFile(filepath)

		if err != nil {
			return
		}

		w.WriteHeader(200)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=UTF8")

		alt, _ := file.GetAltitude(lat, lon)

		fmt.Fprintf(w, "%d", alt)
	} else {
		w.WriteHeader(404)
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	type OutProfileJson struct {
		Result []Point
	}

	var results []string
	var iPath []Point
	var objmap map[string]interface{}
	var lat1 []float64
	var lon1 []float64

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	results = append(results, string(body))

	json.Unmarshal([]byte(results[0]), &objmap)

	step := int(objmap["step"].(float64))
	path := objmap["path"].([]interface{})

	for _, v := range path {
		p1 := v.([]interface{})[0]
		p2 := v.([]interface{})[1]

		v1 := p1.([]interface{})[0].(float64)
		v2 := p1.([]interface{})[1].(float64)

		v3 := p2.([]interface{})[0].(float64)
		v4 := p2.([]interface{})[1].(float64)

		lat1 = MakeRange(v1, v3, float64(step)/1000000)
		lon1 = MakeRange(v2, v4, float64(step)/1000000)

		iter := Zip(lat1, lon1)

		for tuple := iter(); tuple != nil; tuple = iter() {
			lat := tuple[0]
			lon := tuple[1]

			filepath := dem + "/" + CoordsToFilename(lat, lon) + ".hgt"

			if FileExists(filepath) {
				file, err := LoadSRTMFile(filepath)

				if err != nil {
					return
				}

				alt, _ := file.GetAltitude(lat, lon)
				iPath = append(iPath, Point{lat, lon, alt})
			} else {
				w.WriteHeader(404)
			}
		}
	}

	o := OutProfileJson{
		Result: iPath,
	}

	w.WriteHeader(200)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF8")

	b, _ := json.Marshal(o)

	fmt.Fprintf(w, "%s", string(b))
}
