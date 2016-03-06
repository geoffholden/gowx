// Copyright © 2016 Geoff Holden <geoff@geoffholden.com>

package data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Load SQLite DB driver
)

type sqlite_driver struct {
}

func init() {
	RegisterDBDriver("sqlite3", sqlite_driver{})
}

func (sqlite sqlite_driver) OpenDatabase(db *sql.DB) error {
	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS samples (
		timestamp   integer,
		id          text,
		channel     integer,
		serial      string,
		key         string,
		min         real,
		max         real,
		avg         real
	)`); err != nil {
		db.Close()
		return err
	}

	if _, err := db.Exec(`
	CREATE INDEX IF NOT EXISTS i_samples ON samples (
		timestamp,
		key,
		id,
		channel,
		serial
	)`); err != nil {
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

func (sqlite sqlite_driver) QueryWind(db *sql.DB, start int64, key string, col string, id string, channel int) (*sql.Rows, error) {
	var column string
	switch col {
	case "avg":
		column = "avg(d.avg)"
	case "min":
		column = "min(d.min)"
	case "max":
		column = "max(d.max)"
	default:
		column = "avg(d.avg)"
	}
	stmt := `SELECT
			(((dir.avg + 5.125 + 360) % 360) / 11.25) % 32 AS dir,` +
		column +
		` FROM samples dir
		INNER JOIN samples d
			ON d.timestamp = dir.timestamp
		WHERE dir.key = 'WindDir'
			AND d.key = ?
			AND dir.id LIKE ?
			AND d.id LIKE ?
			AND dir.timestamp > ?
		GROUP BY dir;`
	return db.Query(stmt, key, id, id, start)
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
		GROUP BY ts
		ORDER BY ts`
	return db.Query(stmt, interval, interval, key, id, start)
}
