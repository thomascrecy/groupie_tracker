package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//Urls API
const (
	apiURL          = "https://groupietrackers.herokuapp.com/api/artists"
	apiUrlRelations = "https://groupietrackers.herokuapp.com/api/relation"
)

//Structure Artiste
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
}

//Reponse API
type ApiRelations struct {
	ID              int               `json:"id"`
	DatesLocations  map[string][]string `json:"datesLocations"`
}

func fetchData(url string, data interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func getConcert(id string) (*ApiRelations, error) {
	response, err := http.Get(apiUrlRelations + "/" + id)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer response.Body.Close()

	var relation ApiRelations
	if err := json.NewDecoder(response.Body).Decode(&relation); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &relation, nil
}

//Page Index.html
func indexHandler(w http.ResponseWriter, r *http.Request) {
	var artistsData []Artist
	err := fetchData(apiURL, &artistsData)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error1:", err)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error2:", err)
		return
	}

	err = tmpl.Execute(w, artistsData)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error3:", err)
		return
	}
}

//Page ArtistPage.html
func artistPageHandler(w http.ResponseWriter, r *http.Request) {
	var artistsData []Artist
	err := fetchData(apiURL, &artistsData)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error4:", err)
		return
	}

	artistName := r.URL.Query().Get("artist")
	if artistName == "" {
		http.Error(w, "Bad Request: Artist name not provided", http.StatusBadRequest)
		return
	}

	var artist Artist
	for _, a := range artistsData {
		if a.Name == artistName {
			artist = a
			break
		}
	}

	apiRelations, err := getConcert(strconv.Itoa(artist.ID))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error5:", err)
		return
	}

	tmpl, err := template.ParseFiles("templates/artistpage.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error6:", err)
		return
	}

	err = tmpl.Execute(w, struct {
		Artist      Artist
		ApiRelations ApiRelations
	}{
		Artist:      artist,
		ApiRelations: *apiRelations,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error7:", err)
		return
	}
}

//Fonction de recherche
func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	var artistsData []Artist
	err := fetchData(apiURL, &artistsData)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error (searchHandler):", err)
		return
	}

	results := filterArtists(artistsData, query)

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error (searchHandler):", err)
		return
	}

	err = tmpl.Execute(w, results)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		fmt.Println("Internal Server Error (searchHandler):", err)
		return
	}
}

//Filtre artistes pour recherche
func filterArtists(artists []Artist, query string) []Artist {
	var results []Artist
	for _, artist := range artists {
		if containsIgnoreCase(artist.Name, query) {
			results = append(results, artist)
		}
	}
	return results
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func main() {
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artistpage.html", artistPageHandler)
	http.HandleFunc("/search", searchHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}