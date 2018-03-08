package main

import (
	_ "github.com/lib/pq"
	"html/template"
	"net/http"

	"log"
	"path/filepath"
	"os"
	"strings"
	"./templates"
	"./dao"

)

func init() {
	dao.Init()

	templates.Tpl = ParseTemplates()
}

func ParseTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("src/static/app/html", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".gohtml") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				log.Println(err)
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return templ
}

func main() {

	http.HandleFunc("/", templates.Index)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("src/static/"))))
	//http.HandleFunc("/brands", brandsOverview) // to be implemented
	http.HandleFunc("/cars", templates.CarsOverview)
	http.HandleFunc("/cars/delete", templates.DeleteCar)
	http.HandleFunc("/cars/details", templates.CarDetailsView)
	http.HandleFunc("/cars/edit", templates.CarEditView)
	http.ListenAndServe(":8080", nil)
}