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
	"sort"
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
	Id   int64
	Name  string
}

type Model struct {
	Id   int64
	Name  string
}

type CarType struct {
	Id   int64
	Type  string
}

type FuelType struct {
	Id   int64
	Type  string
}

type Car struct {
	CarId   int64
	Brand  string
	Model string
	Type string
	Fuel string
	Consumption float64
	Available bool
}

type CarRow struct  {
	Id int64
	ModelId int64
	BrandId int64
	CarTypeId int64
 	FuelTypeId int64
	Consumption float64
	Available bool

}
type CarEdit struct {
	CarData CarRow
	BrandList []Brand
	ModelList []Model
	TypeList []CarType
	FuelList []FuelType
	AvailableList []bool
	EditPageType string
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
		err := rows.Scan(&brnd.BrandId, &brnd.Brand_Name) // order matters
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
			CarId: carRow["id"].(int64),
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

func mapFromDbRows(query string, args ...int64) (map[int] map[string]interface{}, error){
	var rows *sql.Rows
	var err error
	if len(args)!= 0 {
		log.Println(args[0])
		rows, err = db.Query(query, args[0])
		if(err != nil) {
			log.Println("mapFromDbRows(): Error occured while extracting rows from database.")
			panic(err)
		}
	} else {
		rows, err = db.Query(query)
		if(err != nil) {
			log.Println("mapFromDbRows(): Error occured while extracting rows from database.")
			panic(err)
		}
	}
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
	carRow := CarRow{}
	carRow.carRow(carId)

	car.CarId = carId
	car.Consumption = carRow.Consumption
	car.Available = carRow.Available

	car.setBrand(carRow.BrandId)
	car.setModel(carRow.ModelId)
	car.setCarType(carRow.CarTypeId)
	car.setFuel(carRow.FuelTypeId)
}

//get car data from db by car id
func (carRow *CarRow) carRow(carId int64) {
	carRow.Id = carId
	row := db.QueryRow(`SELECT
				c.model_id,
				c.brand_id,
				c.car_type_id,
				c.fuel_type_id,
				c.fuel_consumption,
				c.rental_free
			       FROM car c WHERE id = $1`, carId).Scan(&carRow.ModelId,&carRow.BrandId,&carRow.CarTypeId,
		&carRow.FuelTypeId,&carRow.Consumption,&carRow.Available)

	switch row {
	case sql.ErrNoRows:
		log.Println("No car with id " + strconv.FormatInt(carId, 10) + "found in database!")
	}
	log.Println("Extract car row from db")

}

//car edit-create func
func carEditView(w http.ResponseWriter, r *http.Request)  {
	carId := r.FormValue("id")
	switch carId {
	case "":
		log.Println("Create new car")
		//createCar()
		return
	default:
		id, _ := strconv.ParseInt(carId, 10, 64)
		carRow := CarRow{}
		carRow.carRow(id)

		carEdit := CarEdit{
			CarData: carRow,
			BrandList: brands(),
			ModelList: brandModels(carRow.BrandId),
			TypeList: carTypes(),
			FuelList: fuelTypes(),
			AvailableList: []bool{true, false},
			EditPageType: "Edit",
		}
		log.Println(carEdit)
		err := tpl.ExecuteTemplate(w, "addEditCars.gohtml", carEdit)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			panic(err)
			return
		}
	}
}

//get all brands from db
func brands() []Brand{
	brandMap, err := mapFromDbRows("SELECT * FROM brand")
	brands := make([]Brand, 0)
	for _, brandRow := range brandMap{

		brand := Brand{
			Id: brandRow["id"].(int64),
			Name: brandRow["name"].(string),
		}
		brands = append(brands, brand)
	}

	//order by names
	sort.Slice(brands, func(i, j int) bool {
		return brands[i].Name < brands[j].Name
	})

	if err != nil {
		panic(err)
	}
	log.Println("Retrieve brand list from db")
	return brands
}

//get all models for specific brand from db
func brandModels(brandId int64) []Model{
	modelMap, err := mapFromDbRows("SELECT m.id, m.name FROM brand_model m WHERE m.brand_id = $1", brandId)
	models := make([]Model, 0)
	for _, modelRow := range modelMap{

		model := Model{
			Id: modelRow["id"].(int64),
			Name: modelRow["name"].(string),
		}
		models = append(models, model)
	}

	//order by names
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	if err != nil {
		panic(err)
	}
	log.Println("Retrieve model list from db")
	return models
}

//get all car types from db
func carTypes() []CarType{
	carTypeMap, err := mapFromDbRows("SELECT * FROM car_type")
	carTypes := make([]CarType, 0)
	for _, carTypeRow := range carTypeMap{

		carType := CarType{
			Id: carTypeRow["id"].(int64),
			Type: carTypeRow["type"].(string),
		}
		carTypes = append(carTypes, carType)
	}

	//order by types
	sort.Slice(carTypes, func(i, j int) bool {
		return carTypes[i].Type < carTypes[j].Type
	})

	if err != nil {
		panic(err)
	}
	log.Println("Retrieve carTypes from db")
	return carTypes
}

//get all fuel types from db
func fuelTypes() []FuelType{
	fuelTypeMap, err := mapFromDbRows("SELECT * FROM fuel_type")
	fuelTypes := make([]FuelType, 0)
	for _, fuelTypeRow := range fuelTypeMap{

		fuelType := FuelType{
			Id: fuelTypeRow["id"].(int64),
			Type: fuelTypeRow["type"].(string),
		}
		fuelTypes = append(fuelTypes, fuelType)
	}

	//order by types
	sort.Slice(fuelTypes, func(i, j int) bool {
		return fuelTypes[i].Type < fuelTypes[j].Type
	})

	if err != nil {
		panic(err)
	}
	log.Println("Retrieve fuelTypes from db")
	return fuelTypes
}