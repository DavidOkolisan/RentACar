package templates

import (
	"html/template"
	"net/http"
	"strconv"
	"log"
	"../dao"
)

var Tpl *template.Template

func Index(w http.ResponseWriter, r *http.Request) {
	err := Tpl.ExecuteTemplate(w, "home.gohtml", nil)

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

func CarsOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	carMap, err := dao.MapFromDbRows("SELECT * FROM car")
	cars := make([]dao.Car, 0)
	for _, carRow := range carMap{

		car := dao.Car{
			CarId: carRow["id"].(int64),
			Consumption: carRow["fuel_consumption"].(float64),
			Available: carRow["rental_free"].(bool),
		}
		car.SetBrand(carRow["brand_id"].(int64))
		car.SetModel(carRow["model_id"].(int64))
		car.SetCarType(carRow["car_type_id"].(int64))
		car.SetFuel(carRow["fuel_type_id"].(int64))
		cars = append(cars, car)
	}
	log.Println("Render car list overview")
	err = Tpl.ExecuteTemplate(w, "cars.gohtml", cars)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		panic(err)
		return
	}
}



//Delete car
func DeleteCar(w http.ResponseWriter, r *http.Request) {
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
func CarDetailsView(w http.ResponseWriter, r *http.Request) {
	carId := r.FormValue("id")
	if carId == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	var car = dao.Car{}
	id, err := strconv.ParseInt(carId, 10, 64)
	if err != nil {
		log.Print("Error during converting carId to type int64")
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	car.FillCarData(id)

	err = Tpl.ExecuteTemplate(w, "viewCar.gohtml", car)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		panic(err)
		return
	}
}


//car edit-create func
func CarEditView(w http.ResponseWriter, r *http.Request)  {
	carId := r.FormValue("id")
	switch carId {
	case "":
		log.Println("Create new car")
		//createCar()
		return
	default:
		id, _ := strconv.ParseInt(carId, 10, 64)
		carRow := dao.CarRow{}
		carRow.CarRow(id)

		carEdit := dao.CarEdit{
			CarData: carRow,
			BrandList: dao.Brands(),
			ModelList: dao.BrandModels(carRow.BrandId),
			TypeList: dao.CarTypes(),
			FuelList: dao.FuelTypes(),
			AvailableList: []bool{true, false},
			EditPageType: "Edit",
		}
		log.Println(carEdit)
		err := Tpl.ExecuteTemplate(w, "addEditCars.gohtml", carEdit)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			panic(err)
			return
		}
	}
}