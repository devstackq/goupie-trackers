package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	temp = template.Must(template.ParseFiles("templates/notify.html", "templates/error2.html", "templates/filter.html", "templates/index.html", "templates/artist.html"))
)

type Locations struct {
	Index []struct {
		ID        int
		Locations []string
	}
}

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
	SortedCity    []SortCity
}

type SortCity struct {
	cities string
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

func main() {

	artists, _ := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	artistsBytes, _ := ioutil.ReadAll(artists.Body)
	artists.Body.Close()
	json.Unmarshal(artistsBytes, &API.Artist)

	locations, _ := http.Get("https://groupietrackers.herokuapp.com/api/locations")
	locationsBytes, _ := ioutil.ReadAll(locations.Body)
	locations.Body.Close()

	json.Unmarshal(locationsBytes, &API.LocationsHtml)
	//	json.Unmarshal(output, &API.SortedCity)

	relations, _ := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	relationsBytes, _ := ioutil.ReadAll(relations.Body)
	relations.Body.Close()
	json.Unmarshal(relationsBytes, &API.RelationHtml)
	//static data, css, js
	static := http.FileServer(http.Dir("public"))
	//secure, not access another files
	http.Handle("/public/", http.StripPrefix("/public/", static))

	http.HandleFunc("/", getAllArtists)
	http.HandleFunc("/artist", getArtist)
	http.HandleFunc("/filter", handlerFilters)
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

//todo unique city
func sortUniqueCity() {

	var sort SortCity
	var tmp []string
	for _, l := range API.LocationsHtml.Index {
		for _, city := range l.Locations {
			if isUnique(tmp, city) {
				tmp = append(tmp, city)
			}
			sort.cities = city
		}
	}
	API.SortedCity = append(API.SortedCity, sort)

	fmt.Println(API.SortedCity)
}

func isUnique(arr []string, s string) bool {
	flag := false
	for _, c := range arr {
		if c != s {
			flag = true
		}
	}
	return flag
}

func getAllArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if r.URL.Path != "/" {
			errorHandler(w, r, http.StatusNotFound)
			return
		}
		sortUniqueCity()

		temp.ExecuteTemplate(w, "index", API)
	}
}

func getArtist(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/artist" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	if r.Method == "GET" {
		temp.ExecuteTemplate(w, "index", "")
	}
	if r.Method == "POST" {
		ID, _ := strconv.Atoi(r.FormValue("uid"))
		API.ID = ID - 1
		temp.ExecuteTemplate(w, "artist", API)
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	//w.WriteHeader(status)
	temp = template.Must(template.ParseGlob("templates/*.html"))
	if status == 404 {
		temp.ExecuteTemplate(w, "error.html", nil)
		w.WriteHeader(404)
		return
	}
	if status == 500 {
		temp.ExecuteTemplate(w, "error1.html", nil)
		w.WriteHeader(500)
		return
	}
	if status == 400 {
		temp.ExecuteTemplate(w, "error2.html", nil)
		w.WriteHeader(400)
		return
	}
	if status == 666 {
		temp.ExecuteTemplate(w, "notify.html", nil)
		w.WriteHeader(666)
		return
	}
}

func handlerFilters(w http.ResponseWriter, r *http.Request) {

	if API.Artist == nil {
		errorHandler(w, r, 500)
		return
	}
	if r.URL.Path != "/filter" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	//todo show group when filter
	from := r.FormValue("from")
	to := r.FormValue("to")
	key := r.Form["input"]
	mems := r.Form["members"]
	cityWas := r.FormValue("cities")
	cb := checkCity(cityWas)

	if len(key) == 0 && len(mems) <= 0 && cityWas == "" {
		errorHandler(w, r, 666)
		fmt.Print("her")
		return
	}
	if len(key) >= 0 {
		if len(cityWas) == 0 || cb == false {
			errorHandler(w, r, 400)
			return
		}
	}

	f, _ := strconv.Atoi(from)
	t, _ := strconv.Atoi(to)
	//case only 1 key
	if len(key) == 1 && len(mems) == 0 {

		if key[0] == "albumdate" {
			//range, from to value
			for i := f; i <= t; i++ {
				FilterByAlbumDate(w, r, i, false)
			}
		} else if key[0] == "creation" {
			for i := f; i <= t; i++ {
				FilterByCreateDate(w, r, i, false)
			}
		}
		//case only mems && more membs
	} else if cityWas != "" && len(key) == 0 {
		FilterByWasCity(w, r, cityWas, false)
	} else if len(mems) >= 1 && len(key) == 0 {
		FilterCountMember(w, r, false, nil)
		//case : mems && creat date, mems and key
	} else if len(key) == 1 && len(mems) >= 1 {
		if key[0] == "creation" && mems != nil {
			var arrArt []Data
			for i := f; i <= t; i++ {
				arrCrDatRes := FilterByCreateDate(w, r, i, true)
				for _, v := range arrCrDatRes {
					arrArt = append(arrArt, v)
				}
			}
			FilterCountMember(w, r, true, arrArt)
		}
		if cityWas != "" && mems != nil {
			var arrCity []Data
			arrCitiesArtist := FilterByWasCity(w, r, cityWas, true)
			for _, v := range arrCitiesArtist {
				arrCity = append(arrCity, v)
			}
			fmt.Print(arrCity)
			FilterCountMember(w, r, true, arrCity)
		}
		if key[0] == "albumdate" && mems != nil {
			var arrAlbum []Data
			for i := f; i <= t; i++ {
				arrCrDatRes := FilterByAlbumDate(w, r, i, true)
				for _, v := range arrCrDatRes {
					arrAlbum = append(arrAlbum, v)
				}
			}
			FilterCountMember(w, r, true, arrAlbum)
		}
		//case only 2 key
	} else if len(key) == 2 {
		if key[1] == "creation" && key[0] == "albumdate" {
			var arrArt []Data
			for i := f; i <= t; i++ {
				arrCrDatRes := FilterByCreateDate(w, r, i, false)
				for _, v := range arrCrDatRes {
					arrArt = append(arrArt, v)
				}
			}
			//FilterByAlbumDate(w, r, i, true)
			fmt.Print(arrArt)
		}
	}

}

func FilterByWasCity(w http.ResponseWriter, r *http.Request, cityArg string, twice bool) []Data {

	var city Data
	var citiesArtist []Data

	for id, v := range API.LocationsHtml.Index {
		for _, cityApi := range v.Locations {
			if strings.Contains(strings.ToLower(cityApi), strings.ToLower(cityArg)) {
				for idArt, aN := range API.Artist {
					if idArt == id {
						if !twice {
							putData(aN)
						}
						city.Name = aN.Name
						citiesArtist = append(citiesArtist, city)
						break
					}
				}
			}
		}
	}
	if !twice {
		temp.ExecuteTemplate(w, "filter", FilterAPI)
		FilterAPI.Artist = nil
	}
	return citiesArtist
}

func FilterCountMember(w http.ResponseWriter, r *http.Request, twice bool, arts []Data) {

	mems := r.Form["members"]
	fm := ""
	lm := ""
	fm = mems[0]
	memsLenFlag := false
	if len(mems) > 1 {
		memsLenFlag = true
		lm = mems[len(mems)-1]
	}
	fim, _ := strconv.Atoi(fm)
	lim, _ := strconv.Atoi(lm)
	count := 0
	if twice {
		if arts != nil {
			for _, a := range API.Artist {
				for _, cA := range arts {
					if a.Name == cA.Name {
						for range a.Members {
							count++
						}
						if memsLenFlag {
							for i := fim; i <= lim; i++ {
								if i == count {
									putData(a)
									break
								}
							}
						} else {
							if count == fim {
								putData(a)
							}
						}
						count = 0
					}
				}
			}
		}
	} else {
		if memsLenFlag {
			for _, a := range API.Artist {
				for range a.Members {
					count++
				}
				for i := fim; i < lim; i++ {
					if count == i {
						putData(a)
						break
					}
					count = 0
				}
			}
		} else {
			for _, v := range API.Artist {
				for range v.Members {
					count++
				}
				if count == fim {
					putData(v)
					break
				}
				count = 0
			}
		}

	}

	temp.ExecuteTemplate(w, "filter", FilterAPI)
	FilterAPI.Artist = nil
}

func FilterByAlbumDate(w http.ResponseWriter, r *http.Request, i int, twice bool) []Data {

	key := strconv.Itoa(i)
	var album Data
	var albums []Data

	for _, v := range API.Artist {
		if strings.Contains(v.FirstAlbum, key) {
			if !twice {
				putData(v)
			}
			album.Name = v.Name
			albums = append(albums, album)
		}
	}

	if !twice {
		temp.ExecuteTemplate(w, "filter", FilterAPI)
		FilterAPI.Artist = nil
	}
	return albums
}

func FilterByCreateDate(w http.ResponseWriter, r *http.Request, k int, twice bool) []Data {

	var crDate Data
	var createDates []Data

	for _, v := range API.Artist {
		if v.CreationDate == k {
			if !twice {
				putData(v)
			}
			crDate.Name = v.Name
			createDates = append(createDates, crDate)
		}
	}
	if !twice {
		temp.ExecuteTemplate(w, "filter", FilterAPI)
		FilterAPI.Artist = nil
	}
	return createDates
}
func checkCity(s string) bool {
	flag := false
	for _, v := range s {
		if v == '-' {
			flag = true
		}
	}
	return flag
}

func putData(arg Singers) {

	var singer Singers
	singer.ID = arg.ID
	singer.Name = arg.Name
	singer.CreationDate = arg.CreationDate
	singer.FirstAlbum = arg.FirstAlbum
	singer.Image = arg.Image
	singer.Members = arg.Members
	FilterAPI.Artist = append(FilterAPI.Artist, singer)
}

// /todo:
//1 locations datalist - unique city send
//2 func optimize
//3 one to many, city -> state
//4 page show artist, not redirect, || add button back home page
//5 reset btn - show all artist
//6 	 params  - another value -  400 handler

// 7delete html, not use
