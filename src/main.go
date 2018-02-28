package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"html/template"
	"net/http"

	"log"
	"path/filepath"
	"os"
	"strings"
	"strconv"
)

var db *sql.DB
var tpl *template.Template

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:super@localhost/carrental?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	tpl = ParseTemplates()
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

type Brand struct {
	Brand_Id   int
	Brand_Name  string
}

type Car struct {
	Car_Id   int64
	Brand  string
	Model string
	Type string
	Fuel string
	Consumption float64
	Available bool
}

type CarEdit struct {
	Car_Id int64
	BrandList map[int64] string
	ModelList map[int64] string
	TypeList map[int64] string
	FuelList map[int64] string
	Consumption float64
	Available bool
}

func main() {

	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("src/static/"))))
	//http.HandleFunc("/brands", brandsOverview) // to be implemented
	http.HandleFunc("/cars", carsOverview)
	http.HandleFunc("/cars/delete", deleteCar)
	http.HandleFunc("/cars/details", carDetailsView)
	http.HandleFunc("/cars/edit", carEditView)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "home.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	//http.ServeFile(w,r,"src/static/app/html/index.gohtml")
}

/* //TO BE IMPLEMENTED
func brandsOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM brand")
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	brnds := make([]Brand, 0)
	for rows.Next() {
		brnd := Brand{}
		err := rows.Scan(&brnd.Brand_Id, &brnd.Brand_Name) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		brnds = append(brnds, brnd)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "brands.gohtml", brnds)
}*/

func carsOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	carMap, err := mapFromDbRows("SELECT * FROM car")
	cars := make([]Car, 0)
	for _, carRow := range carMap{

		car := Car{
			Car_Id: carRow["id"].(int64),
			Consumption: carRow["fuel_consumption"].(float64),
			Available: carRow["rental_free"].(bool),
		}
		car.setBrand(carRow["brand_id"].(int64))
		car.setModel(carRow["model_id"].(int64))
		car.setCarType(carRow["car_type_id"].(int64))
		car.setFuel(carRow["fuel_type_id"].(int64))
		cars = append(cars, car)
	}
	log.Println("Render car list overview")
	err = tpl.ExecuteTemplate(w, "cars.gohtml", cars)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		panic(err)
		return
	}
}

func mapFromDbRows(query string) (map[int] map[string]interface{}, error){
	rows, err := db.Query(query) // Note: Ignoring errors for brevity
	cols, err := rows.Columns()

	rowNr := 0;
	dbMap := make(map[int] map[string]interface{})

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		rowNr++
		dbMap[rowNr] = m
	}

	return dbMap, err
}

//Return brand name on car's brand id input
func (car *Car) setBrand(id int64) {
	err := db.QueryRow("SELECT name FROM brand WHERE id = $1", id).Scan(&car.Brand)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return model name on car's model id input
func (car *Car) setModel(id int64) {
	err := db.QueryRow("SELECT name FROM brand_model WHERE id = $1", id).Scan(&car.Model)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return car type on car's type id input
func (car *Car) setCarType(id int64) {
	err := db.QueryRow("SELECT type FROM car_type WHERE id = $1", id).Scan(&car.Type)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return fuel type on car's fuel id input
func (car *Car) setFuel(id int64) {
	err := db.QueryRow("SELECT type FROM fuel_type WHERE id = $1", id).Scan(&car.Fuel)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Delete car
func deleteCar(w http.ResponseWriter, r *http.Request) {
	carId := r.FormValue("car_id")
	if carId == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// delete car from db
	/*_, err := db.Exec("DELETE FROM car WHERE id=$1;", carId)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
*/
	http.Redirect(w, r, "/cars", http.StatusSeeOther)
}

//view specific car by id
func carDetailsView(w http.ResponseWriter, r *http.Request) {
	carId := r.FormValue("id")
	if carId == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	var car = Car{}
	id, err := strconv.ParseInt(carId, 10, 64)
	if err != nil {
		log.Print("Error during converting carId to type int64")
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	car.fillCarData(id)

	err = tpl.ExecuteTemplate(w, "viewCar.gohtml", car)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		panic(err)
		return
	}
}

//fill car data by id
func (car *Car) fillCarData(carId int64){
	log.Println("Fill car data from db")
	var modelId int64
	var brandId int64
	var carTypeId int64
	var fuelTypeId int64
	var consumption float64
	var available bool
	carRow := db.QueryRow(`SELECT
				c.model_id,
				c.brand_id,
				c.car_type_id,
				c.fuel_type_id,
				c.fuel_consumption,
				c.rental_free
			       FROM car c WHERE id = $1`, carId).Scan(&modelId,&brandId,&carTypeId,&fuelTypeId,&consumption,&available)
	switch carRow {
	case sql.ErrNoRows:
		log.Println("No car with id " + strconv.FormatInt(carId, 10) + "found in database!")
	}

	car.Car_Id = carId
	car.Consumption = consumption
	car.Available = available

	car.setBrand(brandId)
	car.setModel(modelId)
	car.setCarType(carTypeId)
	car.setFuel(fuelTypeId)
}

func carEditView(w http.ResponseWriter, r *http.Request)  {
	carId := r.FormValue("id")
	switch carId {
	case "":
		log.Println("Create new car")
		//createCar()
		return
	default:
		id, _ := strconv.ParseInt(carId, 10, 64)
		car := Car{}
		car.fillCarData(id)

		err := tpl.ExecuteTemplate(w, "addEditCars.gohtml", car)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			panic(err)
			return
		}
	}
}

/*


func booksCreateForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func booksCreateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	bk := Book{}
	bk.Isbn = r.FormValue("isbn")
	bk.Title = r.FormValue("title")
	bk.Author = r.FormValue("author")
	p := r.FormValue("price")

	// validate form values
	if bk.Isbn == "" || bk.Title == "" || bk.Author == "" || p == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// convert form values
	f64, err := strconv.ParseFloat(p, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Please hit back and enter a number for the price", http.StatusNotAcceptable)
		return
	}
	bk.Price = float32(f64)

	// insert values
	_, err = db.Exec("INSERT INTO books (isbn, title, author, price) VALUES ($1, $2, $3, $4)", bk.Isbn, bk.Title, bk.Author, bk.Price)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "created.gohtml", bk)
}

func booksUpdateForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	isbn := r.FormValue("isbn")
	if isbn == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM books WHERE isbn = $1", isbn)

	bk := Book{}
	err := row.Scan(&bk.Isbn, &bk.Title, &bk.Author, &bk.Price)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", bk)
}

func booksUpdateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	bk := Book{}
	bk.Isbn = r.FormValue("isbn")
	bk.Title = r.FormValue("title")
	bk.Author = r.FormValue("author")
	p := r.FormValue("price")

	// validate form values
	if bk.Isbn == "" || bk.Title == "" || bk.Author == "" || p == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// convert form values
	f64, err := strconv.ParseFloat(p, 32)
	if err != nil {
		http.Error(w, http.StatusText(406)+"Please hit back and enter a number for the price", http.StatusNotAcceptable)
		return
	}
	bk.Price = float32(f64)

	// insert values
	_, err = db.Exec("UPDATE books SET isbn = $1, title=$2, author=$3, price=$4 WHERE isbn=$1;", bk.Isbn, bk.Title, bk.Author, bk.Price)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "updated.gohtml", bk)
}

*/
