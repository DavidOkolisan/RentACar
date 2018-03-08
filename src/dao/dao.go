package dao

import (
	"database/sql"
	"strconv"
	"sort"
	"log"
	_ "github.com/lib/pq"
)

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

var db *sql.DB

func Init(){
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:super@localhost/carrental?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
}

func MapFromDbRows(query string, args ...int64) (map[int] map[string]interface{}, error){
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
func (car *Car) SetBrand(id int64) {
	err := db.QueryRow("SELECT name FROM brand WHERE id = $1", id).Scan(&car.Brand)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return model name on car's model id input
func (car *Car) SetModel(id int64) {
	err := db.QueryRow("SELECT name FROM brand_model WHERE id = $1", id).Scan(&car.Model)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return car type on car's type id input
func (car *Car) SetCarType(id int64) {
	err := db.QueryRow("SELECT type FROM car_type WHERE id = $1", id).Scan(&car.Type)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}

//Return fuel type on car's fuel id input
func (car *Car) SetFuel(id int64) {
	err := db.QueryRow("SELECT type FROM fuel_type WHERE id = $1", id).Scan(&car.Fuel)
	switch err {
	case sql.ErrNoRows:
		log.Println("No rows were returned from database!")
	}
}


//fill car data by id
func (car *Car) FillCarData(carId int64){
	log.Println("Fill car data from db")
	carRow := CarRow{}
	carRow.CarRow(carId)

	car.CarId = carId
	car.Consumption = carRow.Consumption
	car.Available = carRow.Available

	car.SetBrand(carRow.BrandId)
	car.SetModel(carRow.ModelId)
	car.SetCarType(carRow.CarTypeId)
	car.SetFuel(carRow.FuelTypeId)
}

//get car data from db by car id
func (carRow *CarRow) CarRow(carId int64) {
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


//get all brands from db
func Brands() []Brand{
	brandMap, err := MapFromDbRows("SELECT * FROM brand")
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
func BrandModels(brandId int64) []Model{
	modelMap, err := MapFromDbRows("SELECT m.id, m.name FROM brand_model m WHERE m.brand_id = $1", brandId)
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
func CarTypes() []CarType{
	carTypeMap, err := MapFromDbRows("SELECT * FROM car_type")
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
func FuelTypes() []FuelType{
	fuelTypeMap, err := MapFromDbRows("SELECT * FROM fuel_type")
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