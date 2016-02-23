package data

import (
	"database/sql"
	"github.com/spf13/viper"
)

type Database struct {
	db     *sql.DB
	driver DBdriver
}

type Row struct {
	Timestamp     int64
	Min, Max, Avg float64
}

type WindRow struct {
	Dir  float64
	Gust float64
	Avg  float64
}

var drivers map[string]DBdriver

type DBdriver interface {
	OpenDatabase(db *sql.DB) error
	Close(db *sql.DB)
	InsertRow(db *sql.DB, timestamp int64, id string, channel int, serial string, key string, min float64, max float64, avg float64) error
	QueryWind(db *sql.DB, start int64) (*sql.Rows, error)
	QueryFirst(db *sql.DB, start int64, key string, id string, channel int) (float64, error)
	QueryLast(db *sql.DB, start int64, key string, id string, channel int) (float64, error)
	QueryRows(db *sql.DB, start int64, key string, id string) (*sql.Rows, error)
	QueryRowsInterval(db *sql.DB, start int64, key string, id string, interval int64) (*sql.Rows, error)
}

func init() {
	drivers = make(map[string]DBdriver)
}

func RegisterDBDriver(name string, driver DBdriver) {
	drivers[name] = driver
}

func DBDrivers() []string {
	names := make([]string, len(drivers))
	i := 0
	for name := range drivers {
		names[i] = name
		i++
	}
	return names
}

func OpenDatabase() (*Database, error) {
	db, err := sql.Open(viper.GetString("dbDriver"), viper.GetString("database"))
	if err != nil {
		return nil, err
	}

	driver := drivers[viper.GetString("dbDriver")]
	driver.OpenDatabase(db)

	return &Database{db, driver}, nil
}

func (database *Database) Close() {
	database.driver.Close(database.db)
	database.db.Close()
}

func (database *Database) InsertRow(timestamp int64, id string, channel int, serial string, key string, min float64, max float64, avg float64) error {
	return database.driver.InsertRow(database.db, timestamp, id, channel, serial, key, min, max, avg)
}

func (database *Database) QueryWind(start int64) <-chan WindRow {
	rows, err := database.driver.QueryWind(database.db, start)
	if err != nil {
		return nil
	}

	ch := make(chan WindRow, 16)
	go func() {
		defer rows.Close()
		for rows.Next() {
			var dir, gust, avg float64
			err := rows.Scan(&dir, &gust, &avg)
			if err != nil {
				continue
			}
			ch <- WindRow{dir, gust, avg}
		}
		close(ch)
	}()
	return ch
}

func (database *Database) QueryFirst(start int64, key string, id string, channel int) (float64, error) {
	return database.driver.QueryFirst(database.db, start, key, id, channel)
}

func (database *Database) QueryLast(start int64, key string, id string, channel int) (float64, error) {
	return database.driver.QueryLast(database.db, start, key, id, channel)
}

func (database *Database) QueryRows(start int64, key string, id string) <-chan Row {
	rows, err := database.driver.QueryRows(database.db, start, key, id)
	if err != nil {
		return nil
	}

	ch := make(chan Row, 64)
	go func() {
		defer rows.Close()
		for rows.Next() {
			var timestamp int64
			var min, max, avg float64

			err := rows.Scan(&timestamp, &min, &max, &avg)
			if err != nil {
				continue
			}
			ch <- Row{timestamp, min, max, avg}
		}
		close(ch)
	}()

	return ch
}

func (database *Database) QueryRowsInterval(start int64, key string, id string, interval int64) <-chan Row {
	rows, err := database.driver.QueryRowsInterval(database.db, start, key, id, interval)
	if err != nil {
		return nil
	}

	ch := make(chan Row, 64)
	go func() {
		for rows.Next() {
			defer rows.Close()
			var timestamp int64
			var min, max, avg float64

			err := rows.Scan(&timestamp, &min, &max, &avg)
			if err != nil {
				continue
			}
			ch <- Row{timestamp, min, max, avg}
		}
		close(ch)
	}()

	return ch
}
