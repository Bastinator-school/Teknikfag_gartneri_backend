package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const MaxReasonableTemperature = 80
const MaxReasonableHumidity = 100

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

type ResponseType struct {
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

type TemperatureIntakeType struct {
	Temperature float32   `json:"temperature"`
	Timestamp   time.Time `json:"timestamp"`
	DeviceId    string    `json:"deviceId"`
}

type HumidityIntakeType struct {
	Humidity  float32   `json:"humidity"`
	Timestamp time.Time `json:"timestamp"`
	DeviceId  string    `json:"deviceId"`
}

type TemperatureHumidityIntakeType struct {
	Temperature float32   `json:"temperature"`
	Humidity    float32   `json:"humidity"`
	Timestamp   time.Time `json:"timestamp"`
	DeviceId    string    `json:"deviceId"`
}

func main() {
	log.Println("starting server")
	db := connectDB()
	defer func() { _ = db.Close() }()
	HandleRequests(db)
}

func Greetings(w http.ResponseWriter, _ *http.Request) {
	writeJSONResponse(w, ResponseType{
		Status:  "ok",
		Code:    http.StatusOK,
		Message: "hello world",
	})
}

func writeJSONResponse(w http.ResponseWriter, response ResponseType) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Code)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		return
	}
	_, _ = w.Write(jsonResponse)
}

func decodeJSONBody(v *http.Request, target interface{}) error {
	decoder := json.NewDecoder(v.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func parseInterval(intervalParam string) (time.Duration, error) {
	intervalParam = strings.TrimSpace(intervalParam)
	if intervalParam == "" {
		return 5 * time.Minute, nil
	}

	if intervalParam == "0" || intervalParam == "0s" || intervalParam == "0m" || intervalParam == "0h" || intervalParam == "0d" {
		return 0, nil
	}

	if strings.HasSuffix(intervalParam, "d") {
		daysPart := strings.TrimSuffix(intervalParam, "d")
		days, err := strconv.Atoi(daysPart)
		if err != nil || days < 0 {
			return 0, fmt.Errorf("invalid day interval")
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(intervalParam)
}

func processTemperatureIntake(ctx context.Context, db *sql.DB, intake TemperatureIntakeType) ResponseType {
	if intake.Temperature > MaxReasonableTemperature {
		return ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Temperature over " + strconv.Itoa(MaxReasonableTemperature) + " assuming error",
			Error:   "temperature too high to believe",
		}
	}

	if err := insertTemperature(ctx, db, intake); err != nil {
		return ResponseType{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "failed to store temperature intake",
			Error:   err.Error(),
		}
	}

	return ResponseType{
		Status:  "ok",
		Code:    http.StatusCreated,
		Message: "temperature intake stored",
		Data:    intake,
	}
}

func processHumidityIntake(ctx context.Context, db *sql.DB, intake HumidityIntakeType) ResponseType {
	log.Println("processing humidty + temp intake")
	if intake.Humidity > MaxReasonableHumidity {
		return ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Humidity over " + strconv.Itoa(MaxReasonableHumidity) + " assuming error",
			Error:   "humidity too high to believe",
		}
	}

	if err := insertHumidity(ctx, db, intake); err != nil {
		return ResponseType{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "failed to store humidity intake",
			Error:   err.Error(),
		}
	}

	return ResponseType{
		Status:  "ok",
		Code:    http.StatusCreated,
		Message: "humidity intake stored",
		Data:    intake,
	}
}

func TemperatureIntake(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	if v.Method != http.MethodPost {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use POST",
		})
		return
	}

	var intake TemperatureIntakeType
	if err := decodeJSONBody(v, &intake); err != nil {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "invalid request body",
			Error:   err.Error(),
		})
		return
	}

	intake.Timestamp = time.Now().UTC()
	log.Println("Received temperature intake:", intake)
	writeJSONResponse(w, processTemperatureIntake(context.Background(), db, intake))
}

func HumidityIntake(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	if v.Method != http.MethodPost {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use POST",
		})
		return
	}

	var intake HumidityIntakeType
	if err := decodeJSONBody(v, &intake); err != nil {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "invalid request body",
			Error:   err.Error(),
		})
		return
	}

	intake.Timestamp = time.Now().UTC()
	log.Println("Received humidity intake:", intake)
	writeJSONResponse(w, processHumidityIntake(context.Background(), db, intake))
}

func TemperatureHumidityIntake(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	if v.Method != http.MethodPost {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use POST",
		})
		return
	}
	log.Println("Received temperature and humidity intake request")
	var intake TemperatureHumidityIntakeType
	if err := decodeJSONBody(v, &intake); err != nil {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "invalid request body",
			Error:   err.Error(),
		})
		return
	}

	timestamp := time.Now().UTC()
	temperatureIntake := TemperatureIntakeType{
		Temperature: intake.Temperature,
		Timestamp:   timestamp,
		DeviceId:    intake.DeviceId,
	}
	humidityIntake := HumidityIntakeType{
		Humidity:  intake.Humidity,
		Timestamp: timestamp,
		DeviceId:  intake.DeviceId,
	}

	temperatureResponse := processTemperatureIntake(context.Background(), db, temperatureIntake)
	if temperatureResponse.Status == "error" {
		writeJSONResponse(w, temperatureResponse)
		return
	}

	humidityResponse := processHumidityIntake(context.Background(), db, humidityIntake)
	if humidityResponse.Status == "error" {
		writeJSONResponse(w, humidityResponse)
		return
	}

	intake.Timestamp = timestamp
	writeJSONResponse(w, ResponseType{
		Status:  "ok",
		Code:    http.StatusCreated,
		Message: "temperature and humidity intake stored",
		Data:    intake,
	})
}

func GetTemperature(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	if v.Method != http.MethodGet {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use GET",
		})
		return
	}

	intervalParam := v.URL.Query().Get("interval")
	if intervalParam == "" {
		intervalParam = "5m"
	}

	interval, err := parseInterval(intervalParam)
	if err != nil || interval < 0 {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "invalid interval",
			Error:   "interval must be a valid duration like 0, 1m, 5m, 1h, 1d; 0 means all-time",
		})
		return
	}

	readings, err := pullTemperatureByInterval(context.Background(), db, interval)
	if err != nil {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "failed to fetch temperature readings",
			Error:   err.Error(),
		})
		return
	}

	writeJSONResponse(w, ResponseType{
		Status:  "ok",
		Code:    http.StatusOK,
		Message: "temperature readings fetched",
		Data:    readings,
	})
}

func GetHumidity(w http.ResponseWriter, v *http.Request, db *sql.DB) {
	if v.Method != http.MethodGet {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusMethodNotAllowed,
			Message: "invalid method",
			Error:   "use GET",
		})
		return
	}

	intervalParam := v.URL.Query().Get("interval")
	if intervalParam == "" {
		intervalParam = "5m"
	}

	interval, err := parseInterval(intervalParam)
	if err != nil || interval < 0 {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "invalid interval",
			Error:   "interval must be a valid duration like 0, 1m, 5m, 1h, 1d; 0 means all-time",
		})
		return
	}

	readings, err := pullHumidityByInterval(context.Background(), db, interval)
	if err != nil {
		writeJSONResponse(w, ResponseType{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "failed to fetch humidity readings",
			Error:   err.Error(),
		})
		return
	}

	writeJSONResponse(w, ResponseType{
		Status:  "ok",
		Code:    http.StatusOK,
		Message: "humidity readings fetched",
		Data:    readings,
	})
}

func HandleRequests(db *sql.DB) {
	http.HandleFunc("/greet", Greetings)

	// Keep separate paths for backward compatibility.
	http.HandleFunc("/api/post/temperature", func(w http.ResponseWriter, r *http.Request) {
		TemperatureIntake(w, r, db)
	})
	http.HandleFunc("/api/post/humidity", func(w http.ResponseWriter, r *http.Request) {
		HumidityIntake(w, r, db)
	})

	// Combined path for devices sending both values in one payload.
	http.HandleFunc("/api/post/temperaturehumidity", func(w http.ResponseWriter, r *http.Request) {
		TemperatureHumidityIntake(w, r, db)
	})

	http.HandleFunc("/api/get/temperature", func(w http.ResponseWriter, r *http.Request) {
		GetTemperature(w, r, db)
	})
	http.HandleFunc("/api/get/humidity", func(w http.ResponseWriter, r *http.Request) {
		GetHumidity(w, r, db)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
