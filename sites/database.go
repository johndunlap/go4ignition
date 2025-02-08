package sites

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

// ColumnInfo metadata about a single column in a database table
type ColumnInfo struct {
	Cid          *int
	Name         *string
	Type         *string
	NotNull      *bool
	DefaultValue *string
	PrimaryKey   *bool
}

type Pragma struct {
	Name  string
	Value string
}

// DefaultPragmas documentation can be found here: https://www.sqlite.org/pragma.html
var defaultPragmas = []Pragma{
	{
		Name:  "journal_mode",
		Value: "WAL",
	}, {
		Name:  "synchronous",
		Value: "NORMAL",
	}, {
		Name:  "encoding",
		Value: "UTF-8",
	}, {
		Name:  "temp_store",
		Value: "memory",
	}, {
		Name:  "mmap_size",
		Value: "30000000000",
	},
}

func InitDb(dbPath string) (*sql.DB, error) {
	// Open a database connection
	d, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Execute pragmas
	for _, pragma := range defaultPragmas {
		_, err = d.Exec("PRAGMA " + pragma.Name + "=\"" + pragma.Value + "\"")
		if err != nil {
			return nil, fmt.Errorf("failed to set pragma %s => %s: %w", pragma.Name, pragma.Value, err)
		}
		log.Printf("sqlite3: %s => %s\n", pragma.Name, pragma.Value)
	}

	return d, nil
}

func CloseDb(db *sql.DB) error {
	log.Println("sqlite3: Running pragma optimize...")

	// https://www.sqlite.org/pragma.html#pragma_optimize
	_, err := db.Exec("pragma optimize")
	if err != nil {
		println("sqlite3: Failed to run pragma optimize")
	}

	log.Println("sqlite3: Closing database...")
	err = db.Close()

	if err != nil {
		return fmt.Errorf("sqlite3: Failed to close database: %w", err)
	}

	return nil
}

func VacuumDb(db *sql.DB, dbPath *string) error {
	log.Println("Attempting to vacuum database: " + *dbPath)

	_, err := db.Exec("VACUUM")

	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	log.Println("OK")
	return nil
}

func RunMigrations(db *sql.DB, migrations []string) error {
	_, err := db.Exec(migrations[0])

	if err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	for migrationID := 1; migrationID < len(migrations); migrationID++ {
		actualMd5SumText := Md5sum(migrations[migrationID])

		timestamp := time.Now().Unix()

		// Query the database
		rows, err := db.Query(`SELECT MigrationID, Md5Sum FROM Migration where MigrationID = ?`, migrationID)

		if err != nil {
			return fmt.Errorf("error querying database: %w", err)
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
			_, err = db.Exec(migrations[migrationID])
			if err != nil {
				return fmt.Errorf("error running migration: %w", err)
			}

			if migrationID > 0 {
				insertQuery := "insert into Migration(MigrationID, Md5Sum, UnixTimestamp) values(?, ?, ?)"
				result, err := db.Exec(insertQuery, migrationID, actualMd5SumText, timestamp)

				if err != nil {
					return fmt.Errorf("error inserting record: %w", err)
				}

				_, err = result.LastInsertId()

				if err != nil {
					return fmt.Errorf("error getting last insert id: %w", err)
				}
			}
		}

		if err := rows.Close(); err != nil {
			return fmt.Errorf("failed to close rows object: %w", err)
		}
	}

	return nil
}
