package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// TODO GET NOT WORKING

const Max_reasonable_temperature = 80
const Max_reasonable_humidity = 100

func connectDB() *sql.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://bastian:haha@localhost:5432/drivhus_db?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

type Response_type struct {
	Status  string      `json:"status"`  // ok:error
	Code    int         `json:"code"`    // http response code
	Message string      `json:"message"` //	 human-readable message
	Data    interface{} `json:"data"`    // DATA if relavent
	Error   string      `json:"error"`   // error message
}
type Temperature_Intake_type struct {
	Temperature float32   `json:"temperature"`
	Timestamp   time.Time `json:"timestamp"`
	DeviceId    string    `json:"deviceId"`
}

type Humidity_Intake_type struct {
	Humidity  float32   `json:"humidity"`
	Timestamp time.Time `json:"timestamp"`
	DeviceId  string    `json:"deviceId"`
}

func main() {
	log.Println("starting server")
	db := connectDB()
	defer db.Close()
	HandleRequests(db)
}
func Greetings(w http.ResponseWriter, v *http.Request) {
	Response := Response_type{
		Status:  "ok",
		Code:    200,
		Message: "hello world",
		Data:    nil,
		Error:   "",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		log.Println(err)
	}
	w.Write(jsonResponse)
}

func Temperature_Intake(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	Response := Response_type{}
	if v.Method != http.MethodPost {
		Response = Response_type{
			Status:  "error",
			Code:    400,
			Message: "invalid method",
			Data:    nil,
			Error:   "",
		}
	} else {
		var intake Temperature_Intake_type
		if err := json.NewDecoder(v.Body).Decode(&intake); err != nil {
			Response = Response_type{
				Status:  "error",
				Code:    400,
				Message: "invalid request body",
				Data:    nil,
				Error:   err.Error(),
			}
		} else {
			log.Println("Received temperature intake: ", intake)

			if intake.Temperature > Max_reasonable_temperature {
				Response = Response_type{
					Status:  "error",
					Code:    400,
					Message: "Temperature over " + strconv.Itoa(Max_reasonable_temperature) + " assuming error",
					Data:    nil,
					Error:   "temperature too high to believe",
				}
			} else {
				intake.Timestamp = time.Now().UTC()
				log.Println("trying to write to db")
				if err := insertTemperature(context.Background(), db, intake); err != nil {
					Response = Response_type{
						Status:  "error",
						Code:    http.StatusInternalServerError,
						Message: "failed to store temperature intake",
						Error:   err.Error(),
					}
				} else {
					Response = Response_type{
						Status:  "ok",
						Code:    http.StatusCreated,
						Message: "temperature intake stored",
						Data:    intake,
						Error:   "",
					}
				}
			}

		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Response.Code)
	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		log.Println(err)
	}
	_, _ = w.Write(jsonResponse)
}

func Humidity_Intake(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	Response := Response_type{}
	if v.Method != http.MethodPost {
		Response = Response_type{
			Status:  "error",
			Code:    400,
			Message: "invalid method",
			Data:    nil,
			Error:   "",
		}
	} else {
		var intake Humidity_Intake_type
		if err := json.NewDecoder(v.Body).Decode(&intake); err != nil {
			Response = Response_type{
				Status:  "error",
				Code:    400,
				Message: "invalid request body",
				Data:    nil,
				Error:   err.Error(),
			}
		} else {
			intake.Timestamp = time.Now().UTC()
			log.Println("Received humidity intake: ", intake)

			if intake.Humidity > Max_reasonable_humidity {
				Response = Response_type{
					Status:  "error",
					Code:    400,
					Message: "Humidity over " + strconv.Itoa(Max_reasonable_temperature) + " assuming error",
					Data:    nil,
					Error:   "temperature too high to believe",
				}
			} else {
				log.Println("trying to write to db")
				if err := insertHumidity(context.Background(), db, intake); err != nil {
					Response = Response_type{
						Status:  "error",
						Code:    http.StatusInternalServerError,
						Message: "failed to store temperature intake",
						Error:   err.Error(),
					}
				} else {
					Response = Response_type{
						Status:  "ok",
						Code:    http.StatusCreated,
						Message: "temperature intake stored",
						Data:    intake,
						Error:   "",
					}
				}
			}

		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Response.Code)
	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		log.Println(err)
	}
	_, _ = w.Write(jsonResponse)
}

func GetTemperature(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	Response := Response_type{}
	if v.Method != http.MethodGet {
		Response = Response_type{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use GET",
		}
	} else {
		intervalParam := v.URL.Query().Get("interval")
		if intervalParam == "" {
			intervalParam = "5m"
		}

		interval, err := time.ParseDuration(intervalParam)
		if err != nil || interval <= 0 {
			Response = Response_type{
				Status:  "error",
				Code:    http.StatusBadRequest,
				Message: "invalid interval",
				Error:   "interval must be a valid duration like 1m, 5m, 1h",
			}
		} else {
			readings, err := pullTemperatureByInterval(context.Background(), db, interval)
			if err != nil {
				Response = Response_type{
					Status:  "error",
					Code:    http.StatusInternalServerError,
					Message: "failed to fetch temperature readings",
					Error:   err.Error(),
				}
			} else {
				Response = Response_type{
					Status:  "ok",
					Code:    http.StatusOK,
					Message: "temperature readings fetched",
					Data:    readings,
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(Response.Code)
	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		log.Println(err)
	}
	_, _ = w.Write(jsonResponse)
}

func HandleRequests(db *sql.DB) {
	http.HandleFunc("/greet", Greetings)
	http.HandleFunc("/api/post/temperature", func(w http.ResponseWriter, r *http.Request) {
		Temperature_Intake(w, r, db)
	})
	http.HandleFunc("/api/post/humidity", func(w http.ResponseWriter, r *http.Request) {
		Humidity_Intake(w, r, db)
	})

	http.HandleFunc("/api/get/temperature", func(w http.ResponseWriter, r *http.Request) {
		GetTemperature(w, r, db)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
