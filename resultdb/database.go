package resultdb

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

const (
	username = ""
	password = ""
	host     = ""
	port     = ""
	dbname   = ""
)

var DB *sql.DB

func InitDB() error {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return fmt.Errorf("error opening MySQL connection: %v", err)
	}

	// Set the maximum number of open connections to the database
	db.SetMaxOpenConns(20)

	DB = db
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func insertData(query string, args ...interface{}) error {
	// Insert data into the MySQL database
	_, err := DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error inserting into MySQL: %v", err)
	}
	return nil
}

func InsertScanResult(id string, onion string, scannedAt time.Time, result string, portRange string, lastScannedAt time.Time, status string) error {
	return insertData(
		"INSERT INTO scan_results (OnionID, OnionURL, ScannedAt, OpenPorts, PortRange, LastScannedAt, Status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, onion, scannedAt, result, portRange, lastScannedAt, status,
	)
}

func InsertScanHistory(onionID string, lastScannedAt time.Time, openPorts string, portRange string, status string) error {
	return insertData(
		"INSERT INTO scan_history (OnionID, LastScannedAt, OpenPorts, PortRange, Status) VALUES (?, ?, ?, ?, ?)",
		onionID, lastScannedAt, openPorts, portRange, status,
	)
}

func InsertOrUpdate(id string, onion string, scannedAt time.Time, result string, portRange string, lastScannedAt time.Time, status string) error {
	// Check if the onion is already present in scan_results
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM scan_results WHERE OnionID = ?", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking if onion exists in scan_results: %v", err)
	}

	if count == 0 {
		// Onion is not present, insert a new row
		err = InsertScanResult(id, onion, scannedAt, result, portRange, lastScannedAt, status)
		if err != nil {
			return fmt.Errorf("error inserting into scan_results: %v", err)
		}
	} else {
		// Onion is already present, update lastScannedAt column
		_, err := DB.Exec("UPDATE scan_results SET LastScannedAt = ? WHERE OnionID = ?", lastScannedAt, id)
		if err != nil {
			return fmt.Errorf("error updating lastScannedAt in scan_results: %v", err)
		}
	}

	// Insert into scan_history
	err = InsertScanHistory(id, lastScannedAt, result, portRange, status)
	if err != nil {
		return fmt.Errorf("error inserting into scan_history: %v", err)
	}

	return nil
}
