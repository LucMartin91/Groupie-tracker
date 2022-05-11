package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type APIpassage struct {
	Id             int
	Image          string
	Name           string
	Listename      []string
	Members        []string
	CreationDate   string
	FirstAlbum     string
	Locations      string // Lien de nos locations dans l'API
	Locationsrécup struct {
		Id         int
		Locations  []string
		Dates      string // Lien de nos dates qui correspondent à nos locations
		Datesrécup struct {
			Id    int
			Dates []string
		}
	}
	Locationséparées []string
}

func SplitWithEtoile(s string) []string {
	var listemots []string        // La liste, vide au début, qui va contenir nos dates bien rangées par la suite
	mot := ""                     // On déclare une variable vide dans laquelle les dates vont passer une à une
	for i := 0; i < len(s); i++ { // Ici on boucle sur la longueur de notre string comprenant toutes les dates afin de tout traiter
		if s[i] != '*' && s[i] != '\t' && s[i] != '\n' && i != len(s)-1 { // tant que le caractère sur equel on passe n'est pas "*" , un retour à la ligne ou un tab, alors on continue de remplir notre mot
			mot = mot + string(s[i])
		} else {
			if i == len(s)-1 {
				mot = mot + string(s[i]) // Si les caractères donnés au dessus apparaissent, alors on vide notre string après l'avoir ajoutée à notre tableau liste de mots
			}
			if mot != "" {
				listemots = append(listemots, mot)
			}
			mot = ""
		}
	}
	return listemots
}

func main() {
	var templatesDir = os.Getenv("TEMPLATES_DIR")
	var APIstockage []APIpassage
	for i := 1; i <= 51; i++ {
		var API APIpassage
		artists, err := http.Get("https://groupietrackers.herokuapp.com/api/artists/" + strconv.Itoa(i))
		if err != nil { // Ici on va aller chercher le contenu de l'API correspondant à chaque artiste pour les mettre dans des structures stockées à chaque tout de boucle dans un tableau de struct
			fmt.Print(err.Error())
			os.Exit(1)
		}
		artistsliste, err := ioutil.ReadAll(artists.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(artistsliste, &API)

		loc, err := http.Get(API.Locations) // Ici on récupère les infos de l'API correspondant aux locations et au dates correspondantes à l'artiste en question
		if err != nil {
			fmt.Println(err.Error())
		}
		locliste, err := ioutil.ReadAll(loc.Body)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = json.Unmarshal(locliste, &API.Locationsrécup)

		datesloc, err := http.Get(API.Locationsrécup.Dates) // Ici on récupère les infos correspondant aux dates des locations récupérées plus haut
		if err != nil {
			fmt.Println(err.Error())
		}
		dateslocliste, err := ioutil.ReadAll(datesloc.Body)
		if err != nil {
			fmt.Println(err.Error())
		}
		err = json.Unmarshal(dateslocliste, &API.Locationsrécup.Datesrécup)
		var datestring string
		for i := 0; i < len(API.Locationsrécup.Datesrécup.Dates); i++ {
			datestring = datestring + string(API.Locationsrécup.Datesrécup.Dates[i]) + " "
		}
		API.Locationséparées = SplitWithEtoile(datestring)

		APIstockage = append(APIstockage, API)
		API.Listename = append(API.Listename, API.Name)
		datestring = ""
		fmt.Println(API.Locationséparées)
	}
	tmpl, _ := template.ParseGlob("./html/*.html")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		scanner_input := r.FormValue("scan")
		fmt.Println(scanner_input)
		for i := range APIstockage {
			for _, o := range APIstockage[i].Members {
				if scanner_input == APIstockage[i].Name || scanner_input == APIstockage[i].FirstAlbum || scanner_input == o {
					fmt.Println(APIstockage[i].Id)
					http.Redirect(w, r, "/artiste/"+strconv.Itoa(APIstockage[i].Id), http.StatusSeeOther)
				}
			}
		}

		tmpl.ExecuteTemplate(w, "index", APIstockage)
	})
	http.HandleFunc("/artiste/", func(w http.ResponseWriter, r *http.Request) {
		Artistechoisi := r.URL.RequestURI()[9:]
		ArtisteID, err := strconv.Atoi(Artistechoisi)
		if err != nil {
			fmt.Println(err.Error())
		}
		tmpl.ExecuteTemplate(w, "artiste", APIstockage[ArtisteID-1])
	})
	http.HandleFunc("/ContactUs", func(w http.ResponseWriter, r *http.Request) {
		(template.Must(template.ParseFiles(filepath.Join(templatesDir, "./html/ContactUs.html")))).Execute(w, "")
	})

	http.ListenAndServe("localhost:555", nil)
}
