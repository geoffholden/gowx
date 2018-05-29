// Copyright © 2018 Geoff Holden <geoff@geoffholden.com>

package data

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type postgres_driver struct {
}

func init() {
	RegisterDBDriver("postgres", postgres_driver{})
}

func (postgres postgres_driver) OpenDatabase(db *sql.DB) error {
	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS samples (
		timestamp   timestamp,
		id          text,
		channel     integer,
		serial      text,
		key         text,
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

func (postgres postgres_driver) Close(db *sql.DB) {
}

func (postgres postgres_driver) InsertRow(db *sql.DB, timestamp int64, id string, channel int, serial string, key string, min float64, max float64, avg float64) error {
	stmt := `INSERT INTO samples (
		timestamp,
		id,
		channel,
		serial,
		key,
		min, max, avg
	) VALUES (to_timestamp($1), $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(stmt, timestamp, id, channel, serial, key, min, max, avg)
	return err
}

func (postgres postgres_driver) QueryWind(db *sql.DB, start int64, key string, col string, id string, channel int) (*sql.Rows, error) {
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
			cast((cast(dir.avg + 5.125 + 360 as numeric) % 360) / 11.25 as numeric) % 32 AS dir,` +
		column +
		` FROM samples dir
		INNER JOIN samples d
			ON d.timestamp = dir.timestamp
		WHERE dir.key = 'WindDir'
			AND d.key = $1
			AND dir.id LIKE $2
			AND d.id LIKE $3
			AND dir.timestamp > to_timestamp($4)
		GROUP BY dir;`
	return db.Query(stmt, key, id, id, start)
}

func (postgres postgres_driver) QueryFirst(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key = $1 AND
			id LIKE $2 AND
			channel = $3 AND
			timestamp > to_timestamp($4)
		ORDER BY timestamp
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (postgres postgres_driver) QueryLast(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key = $1 AND
			id LIKE $2 AND
			channel = $3 AND
			timestamp > to_timestamp($4)
		ORDER BY timestamp DESC
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (postgres postgres_driver) QueryRows(db *sql.DB, start int64, key string, id string) (*sql.Rows, error) {
	stmt := `SELECT cast(extract(epoch from timestamp) as bigint),min,max,avg FROM samples
		WHERE
			key = $1 AND
			id LIKE $2 AND
			timestamp > to_timestamp($3)
		ORDER BY timestamp`
	return db.Query(stmt, key, id, start)
}

func (postgres postgres_driver) QueryRowsInterval(db *sql.DB, start int64, key string, id string, interval int64) (*sql.Rows, error) {
	stmt := `SELECT
			CAST(extract(epoch from timestamp)/$1 as bigint) * $2 as ts,
			MIN(min),
			MAX(max),
			AVG(avg)
		FROM samples
		WHERE
			key = $3 AND
			id LIKE $4 AND
			timestamp > to_timestamp($5)
		GROUP BY ts
		ORDER BY ts`
	return db.Query(stmt, interval, interval, key, id, start)
}
