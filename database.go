package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/user"
	"strconv"
	"time"
)

var dataDir string
var dbPath string

var dbTypeMap = map[string]string{
	"INTEGER": "int64",
	"REAL":    "float64",
	"TEXT":    "string",
	"BLOB":    "[]byte",
}

// Pragma documentation can be found here: https://www.sqlite.org/pragma.html
var pragmaOrder = []string{"journal_mode", "synchronous", "encoding", "temp_store", "mmap_size"}
var pragmaValues = map[string]string{
	"journal_mode": "WAL",
	"synchronous":  "NORMAL",
	"encoding":     "UTF-8",
	"temp_store":   "memory",
	"mmap_size":    "30000000000",
}

// ColumnInfo metadata about a single column in a database table
type ColumnInfo struct {
	Cid          *int
	Name         *string
	Type         *string
	NotNull      *bool
	DefaultValue *string
	PrimaryKey   *bool
}

var (
	DatabaseHandle *sql.DB
)

func InitDb() {
	usr, err := user.Current()

	if err != nil {
		panic("Failed to find current user: " + err.Error())
	}

	dataDir = usr.HomeDir + "/.go4ignition"
	dbPath = dataDir + "/go4ignition.db"

	if !DirExists(dataDir) {
		CreateDir(dataDir)
		log.Println("Created data directory: " + dataDir)
	}

	// Open a database connection
	d, err := sql.Open("sqlite3", dbPath)
	DatabaseHandle = d

	if err != nil {
		panic("Error opening database:" + err.Error())
	}

	// Execute pragmas
	for _, key := range pragmaOrder {
		_, err = DatabaseHandle.Exec("PRAGMA " + key + "=\"" + pragmaValues[key] + "\"")
		if err != nil {
			panic("Failed to set pragma " + key + " => " + pragmaValues[key] + ": " + err.Error())
		}
		log.Println("sqlite3: " + key + " => " + pragmaValues[key])
	}
}

func CloseDb() {
	log.Println("sqlite3: Running pragma optimize...")

	// https://www.sqlite.org/pragma.html#pragma_optimize
	_, err := DatabaseHandle.Exec("pragma optimize")
	if err != nil {
		println("sqlite3: Failed to run pragma optimize")
	}

	log.Println("sqlite3: Closing database...")
	err = DatabaseHandle.Close()
	if err != nil {
		println("sqlite3: Failed to close database: " + err.Error())
	}
}

func VacuumDb(dbPath *string) {
	log.Println("Attempting to vacuum database: " + *dbPath)

	_, err := DatabaseHandle.Exec("VACUUM")

	if err != nil {
		panic("Failed to vacuum database: " + err.Error())
	}

	log.Println("OK")
	os.Exit(0)
}

func RunMigrations(migrations []string) {
	_, err := DatabaseHandle.Exec(migrations[0])

	if err != nil {
		panic("Error creating table: " + err.Error())
	}

	for migrationID := 1; migrationID < len(migrations); migrationID++ {
		actualMd5SumText := Md5sum(migrations[migrationID])

		timestamp := time.Now().Unix()

		// Query the database
		rows, err := DatabaseHandle.Query(`SELECT MigrationID, Md5Sum FROM Migration where MigrationID = ?`, migrationID)

		if err != nil {
			panic("Error querying database: " + err.Error())
		}

		hasNext := rows.Next()

		if hasNext {
			var id int
			var expectedMd5SumText string
			_ = rows.Scan(&id, &expectedMd5SumText)

			if expectedMd5SumText != actualMd5SumText {
				println("Md5Sum for migration " + strconv.Itoa(migrationID) + " Md5Sum has changed from " + expectedMd5SumText + " to " + actualMd5SumText)
				panic(1)
			}
		} else {
			log.Println("Running migration " + strconv.Itoa(migrationID) + " with checksum " + actualMd5SumText)
			_, err = DatabaseHandle.Exec(migrations[migrationID])
			if err != nil {
				println("Error running migration: " + err.Error())
				panic(1)
			}

			if migrationID > 0 {
				insertQuery := "insert into Migration(MigrationID, Md5Sum, UnixTimestamp) values(?, ?, ?)"
				result, err := DatabaseHandle.Exec(insertQuery, migrationID, actualMd5SumText, timestamp)

				if err != nil {
					panic("Error inserting record: " + err.Error())
				}

				_, err = result.LastInsertId()

				if err != nil {
					panic("Error getting last insert id: " + err.Error())
				}
			}
		}

		if err := rows.Close(); err != nil {
			println("Failed to close rows object: " + err.Error())
		}
	}
}
