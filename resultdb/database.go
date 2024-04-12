package resultdb

import (
	"database/sql"
	"fmt"
	"strings"
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

func InitDB() (*sql.DB, error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return db, fmt.Errorf("error opening MySQL connection: %v", err)
	}

	// Set the maximum number of open connections to the database
	db.SetMaxOpenConns(20)

	return db, nil
}

func CloseDB(conn *sql.DB) {
	if conn != nil {
		conn.Close()
	}
}

func insertData(conn *sql.DB, query string, args ...interface{}) error {
	// Insert data into the MySQL database
	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error inserting into MySQL: %v", err)
	}
	return nil
}

func InsertResponseTime(conn *sql.DB, id string, onion string, torProxyAddress string, scannedAt time.Time, status string, timeDiff float64, errMessage string) error {
	return insertData(conn,
		"INSERT INTO response_time (OnionID, OnionURL, TorProxyAddress, ScannedAt, Status, Time, ErrorMessage) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, onion, torProxyAddress, status, timeDiff, errMessage,
	)
}

func InsertScanResult(conn *sql.DB, id string, onion string, scannedAt time.Time, result string, portRange string, lastScannedAt time.Time, status string) error {
	return insertData(conn,
		"INSERT INTO scan_results (OnionID, OnionURL, ScannedAt, OpenPorts, PortRange, LastScannedAt, Status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, onion, scannedAt, result, portRange, lastScannedAt, status,
	)
}

func InsertScanHistory(conn *sql.DB, onionID string, lastScannedAt time.Time, openPorts string, portRange string, status string) error {
	return insertData(conn,
		"INSERT INTO scan_history (OnionID, LastScannedAt, OpenPorts, PortRange, Status) VALUES (?, ?, ?, ?, ?)",
		onionID, lastScannedAt, openPorts, portRange, status,
	)
}

func InsertOrUpdate(conn *sql.DB, id string, onion string, scannedAt time.Time, result string, portRange string, lastScannedAt time.Time, status string) error {
	// Check if the onion is already present in scan_results
	var count int
	err := conn.QueryRow("SELECT COUNT(*) FROM scan_results WHERE OnionID = ?", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking if onion exists in scan_results: %v", err)
	}

	if count == 0 {
		// Onion is not present, insert a new row
		err = InsertScanResult(conn, id, onion, scannedAt, result, portRange, lastScannedAt, status)
		if err != nil {
			return fmt.Errorf("error inserting into scan_results: %v", err)
		}
	} else {
		// Get the current OpenPorts value from scan_results
        var currentOpenPorts string
        err = conn.QueryRow("SELECT OpenPorts FROM scan_results WHERE OnionID = ?", id).Scan(&currentOpenPorts)
        if err != nil {
            return fmt.Errorf("error getting current OpenPorts from scan_results: %v", err)
        }

		updatedOpenPorts := unionPorts(currentOpenPorts, result)

		// Onion is already present, update lastScannedAt and OpenPorts column
		_, err := conn.Exec("UPDATE scan_results SET LastScannedAt = ?, OpenPorts = ? WHERE OnionID = ?", lastScannedAt, updatedOpenPorts, id)
		if err != nil {
			return fmt.Errorf("error updating lastScannedAt and openPorts in scan_results: %v", err)
		}

		// Insert into scan_history
		err = InsertScanHistory(conn, id, lastScannedAt, result, portRange, status)
		if err != nil {
			return fmt.Errorf("error inserting into scan_history: %v", err)
		}
	}

	return nil
}

func unionPorts(existingPorts, newPorts string) string {
    // Convert existing and new port ranges to sets
    existingSet := make(map[string]struct{})
    newSet := make(map[string]struct{})

	if existingPorts != "" {
		// Add existing ports to set
		for _, port := range strings.Split(existingPorts, ",") {
			existingSet[port] = struct{}{}
		}
	}
    
	if newPorts != "" {
		// Add new ports to set
		for _, port := range strings.Split(newPorts, ",") {
			newSet[port] = struct{}{}
		}
	}
	
    // Perform union operation
    for port := range newSet {
        existingSet[port] = struct{}{}
    }

    // Construct the result
    var result []string
    for port := range existingSet {
        result = append(result, port)
    }

    return strings.Join(result, ",")
}

// TODO: function for merging port range

// TODO: function for handling status when open port detected for a status = "temporarily down"

