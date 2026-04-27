package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // To register the driver.
)

func insertTemperature(ctx context.Context, db *sql.DB, intake TemperatureIntakeType) error {
	if db == nil {
		return errors.New("database is nil")
	}

	const query = `
		INSERT INTO temperature_readings (temperature, timestamp, device_id)
		VALUES ($1, $2, $3)
	`
	log.Println("inserting temperature")
	_, err := db.ExecContext(ctx, query, intake.Temperature, intake.Timestamp, intake.DeviceId)
	if err != nil {
		return fmt.Errorf("insert temperature: %w", err)
	}

	return nil
}

func insertHumidity(ctx context.Context, db *sql.DB, intake HumidityIntakeType) error {
	if db == nil {
		return errors.New("database is nil")
	}

	const query = `
		INSERT INTO humidity_readings (humidity, timestamp, device_id)
		VALUES ($1, $2, $3)
	`
	log.Println("inserting humidity")
	_, err := db.ExecContext(ctx, query, intake.Humidity, intake.Timestamp, intake.DeviceId)
	if err != nil {
		return fmt.Errorf("insert humidity: %w", err)
	}

	return nil
}

func pullTemperatureByInterval(ctx context.Context, db *sql.DB, interval time.Duration) ([]TemperatureIntakeType, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	if interval < 0 {
		return nil, errors.New("interval must not be negative")
	}

	const queryWithInterval = `
		SELECT temperature, timestamp, device_id
		FROM temperature_readings
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
	`
	const queryAllTime = `
		SELECT temperature, timestamp, device_id
		FROM temperature_readings
		ORDER BY timestamp ASC
	`

	query := queryWithInterval
	var rows *sql.Rows
	var err error

	if interval == 0 {
		query = queryAllTime
		rows, err = db.QueryContext(ctx, query)
	} else {
		cutoff := time.Now().UTC().Add(-interval)
		rows, err = db.QueryContext(ctx, query, cutoff)
	}

	if err != nil {
		return nil, fmt.Errorf("pull temperature by interval: %w", err)
	}
	defer func() { _ = rows.Close() }()

	readings := make([]TemperatureIntakeType, 0)
	for rows.Next() {
		var intake TemperatureIntakeType
		if err := rows.Scan(&intake.Temperature, &intake.Timestamp, &intake.DeviceId); err != nil {
			return nil, fmt.Errorf("scan temperature row: %w", err)
		}
		readings = append(readings, intake)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate temperature rows: %w", err)
	}

	return readings, nil
}

func pullHumidityByInterval(ctx context.Context, db *sql.DB, interval time.Duration) ([]HumidityIntakeType, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	if interval < 0 {
		return nil, errors.New("interval must not be negative")
	}

	const queryWithInterval = `
		SELECT humidity, timestamp, device_id
		FROM humidity_readings
		WHERE timestamp >= $1
		ORDER BY timestamp ASC
	`
	const queryAllTime = `
		SELECT humidity, timestamp, device_id
		FROM humidity_readings
		ORDER BY timestamp ASC
	`

	query := queryWithInterval
	var rows *sql.Rows
	var err error

	if interval == 0 {
		query = queryAllTime
		rows, err = db.QueryContext(ctx, query)
	} else {
		cutoff := time.Now().UTC().Add(-interval)
		rows, err = db.QueryContext(ctx, query, cutoff)
	}

	if err != nil {
		return nil, fmt.Errorf("pull humidity by interval: %w", err)
	}
	defer func() { _ = rows.Close() }()

	readings := make([]HumidityIntakeType, 0)
	for rows.Next() {
		var intake HumidityIntakeType
		if err := rows.Scan(&intake.Humidity, &intake.Timestamp, &intake.DeviceId); err != nil {
			return nil, fmt.Errorf("scan humidity row: %w", err)
		}
		readings = append(readings, intake)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate temperature rows: %w", err)
	}

	return readings, nil
}
