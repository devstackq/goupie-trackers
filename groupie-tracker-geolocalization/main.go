package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

var (
	temp = template.Must(template.ParseFiles("templates/index.html"))
)

type Locations struct {
	Index []struct {
		ID        int
		Locations []string
	}
}

//give locations and date artist
type Relations struct {
	Index []struct {
		ID             int
		DatesLocations map[string][]string
	}
}

var API struct {
	ID            int
	IDS           int
	GroupWasCity  []int
	IDWC          int
	Artist        []Singers
	LocationsHtml Locations
	RelationHtml  Relations
	Location      []LocationsHtml
	SortedCity    []SortCity
}

type GroupWasCity struct {
	Name string
}
type AlbumDate struct {
	Name string
}
type Data struct {
	Name string
}

var FilterAPI struct {
	ID     int
	Artist []Singers
}

type Singers struct {
	ID           int
	Image        string
	Name         string
	Members      []string
	CreationDate int
	FirstAlbum   string
}

type LocationsHtml struct {
	Lat float32 `json:"latitude"`
	Lon float32 `json:"longitude"`
}
type SortCity struct {
	cities string
}

func main() {
	//AnSKQmEZC2RbKmnr3UsEqmxo7yDmN4ATG1B3bL9mG9g7j_BJV4od1Dex2bdJL7Wf
	artists, _ := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	artistsBytes, _ := ioutil.ReadAll(artists.Body)
	artists.Body.Close()
	json.Unmarshal(artistsBytes, &API.Artist)

	locations, _ := http.Get("https://groupietrackers.herokuapp.com/api/locations")
	locationsBytes, _ := ioutil.ReadAll(locations.Body)
	locations.Body.Close()
	json.Unmarshal(locationsBytes, &API.LocationsHtml)

	relations, _ := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	relationsBytes, _ := ioutil.ReadAll(relations.Body)
	relations.Body.Close()
	json.Unmarshal(relationsBytes, &API.RelationHtml)

	http.HandleFunc("/", index)
	http.HandleFunc("/map", jsonData)
	err := http.ListenAndServe(":6969", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func jsonData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		sendFindCityClient(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var art Singers
		for _, a := range API.Artist {
			art.Name = a.Name
			API.Artist = append(API.Artist, art)
		}
		temp.ExecuteTemplate(w, "index", API.Artist)
	}
}

// notification -> when artist not found

func sendFindCityClient(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Print(string(body))
	citiesArtist := FindCityArtist(w, r, strings.ToLower(string(body)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(citiesArtist)
}

//find city, city convert to coords
func FindCityArtist(w http.ResponseWriter, r *http.Request, artist string) []Data {

	var city Data
	var citiesArtist []Data

	for idA, a := range API.Artist {
		if strings.ToLower(a.Name) == artist {
			for idL, l := range API.LocationsHtml.Index {
				if idA == idL {
					for _, cityAPI := range l.Locations {
						city.Name = cityAPI
						citiesArtist = append(citiesArtist, city)
					}
				}
			}
		}
	}
	FilterAPI.Artist = nil
	return citiesArtist
}
