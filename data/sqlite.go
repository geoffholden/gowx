package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Load SQLite DB driver
)

type sqlite_driver struct {
}

func init() {
	RegisterDBDriver("sqlite", sqlite_driver{})
}

func (sqlite sqlite_driver) OpenDatabase(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS samples (
		timestamp   integer,
		id          text,
		channel     integer,
		serial      string,
		key         string,
		min         real,
		max         real,
		avg         real
	)`)
	if err != nil {
		db.Close()
		return err
	}

	return nil
}

func (sqlite sqlite_driver) Close(db *sql.DB) {
}

func (sqlite sqlite_driver) InsertRow(db *sql.DB, timestamp int64, id string, channel int, serial string, key string, min float64, max float64, avg float64) error {
	stmt := `INSERT INTO samples (
		timestamp,
		id,
		channel,
		serial,
		key,
		min, max, avg
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(stmt, timestamp, id, channel, serial, key, min, max, avg)
	return err
}

func (sqlite sqlite_driver) QueryWind(db *sql.DB, start int64) (*sql.Rows, error) {
	stmt := `SELECT
			(((dir.avg + 5.125 + 360) % 360) / 11.25) % 32 AS dir,
			max(gust.max),
			avg(avg.avg)
		FROM samples dir
		INNER JOIN samples gust
			ON gust.timestamp = dir.timestamp
		INNER JOIN samples avg
			ON avg.timestamp = gust.timestamp
		WHERE dir.key = 'WindDir'
			AND gust.key = 'CurrentWind'
			AND avg.key = 'AverageWind'
			AND dir.timestamp > ?
		GROUP BY dir;`
	return db.Query(stmt, start)
}

func (sqlite sqlite_driver) QueryFirst(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key = ? AND
			id LIKE ? AND
			channel = ? AND
			timestamp > ?
		ORDER BY timestamp
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (sqlite sqlite_driver) QueryLast(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key = ? AND
			id LIKE ? AND
			channel = ? AND
			timestamp > ?
		ORDER BY timestamp DESC
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (sqlite sqlite_driver) QueryRows(db *sql.DB, start int64, key string, id string) (*sql.Rows, error) {
	stmt := `SELECT timestamp,min,max,avg FROM samples
		WHERE
			key = ? AND
			id LIKE ? AND
			timestamp > ?
		ORDER BY timestamp`
	return db.Query(stmt, key, id, start)
}

func (sqlite sqlite_driver) QueryRowsInterval(db *sql.DB, start int64, key string, id string, interval int64) (*sql.Rows, error) {
	stmt := `SELECT
			CAST(timestamp/? as INTEGER) * ? as ts,
			MIN(min),
			MAX(max),
			AVG(avg)
		FROM samples
		WHERE
			key = ? AND
			id LIKE ? AND
			timestamp > ?
		ORDER BY timestamp`
	return db.Query(stmt, interval, interval, key, id, start)
}
