// Copyright © 2018 Geoff Holden <geoff@geoffholden.com>

package data

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type mysql_driver struct {
}

func init() {
	RegisterDBDriver("mysql", mysql_driver{})
}

func (mysql mysql_driver) OpenDatabase(db *sql.DB) error {
	if _, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS samples (
		timestamp   timestamp,
		id          varchar(128),
		channel     integer,
		serial      varchar(128),
		key_        varchar(128),
		min         double,
		max         double,
		avg         double
	)`); err != nil {
		db.Close()
		return err
	}

	row := db.QueryRow(`
	SELECT COUNT(1) IndexIsThere FROM INFORMATION_SCHEMA.STATISTICS WHERE
		table_schema=DATABASE() AND
		table_name='samples' AND
		index_name='i_samples';
	`)
	var result int
	err := row.Scan(&result)
	if err != nil {
		db.Close()
		return err
	}

	if result == 0 {
		if _, err := db.Exec(`
		CREATE INDEX i_samples ON samples (
			timestamp,
			key_,
			id,
			channel,
			serial
		)`); err != nil {
			db.Close()
			return err
		}
	}

	return nil
}

func (mysql mysql_driver) Close(db *sql.DB) {
}

func (mysql mysql_driver) InsertRow(db *sql.DB, timestamp int64, id string, channel int, serial string, key string, min float64, max float64, avg float64) error {
	stmt := `INSERT INTO samples (
		timestamp,
		id,
		channel,
		serial,
		key_,
		min, max, avg
	) VALUES (FROM_UNIXTIME(?), ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(stmt, timestamp, id, channel, serial, key, min, max, avg)
	return err
}

func (mysql mysql_driver) QueryWind(db *sql.DB, start int64, key string, col string, id string, channel int) (*sql.Rows, error) {
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
		WHERE dir.key_ = 'WindDir'
			AND d.key_ = ?
			AND dir.id LIKE ?
			AND d.id LIKE ?
			AND dir.timestamp > FROM_UNIXTIME(?)
		GROUP BY dir;`
	return db.Query(stmt, key, id, id, start)
}

func (mysql mysql_driver) QueryFirst(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key_ = ? AND
			id LIKE ? AND
			channel = ? AND
			timestamp > FROM_UNIXTIME(?)
		ORDER BY timestamp
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (mysql mysql_driver) QueryLast(db *sql.DB, start int64, key string, id string, channel int) (float64, error) {
	stmt := `SELECT
		avg FROM samples
		WHERE
			key_ = ? AND
			id LIKE ? AND
			channel = ? AND
			timestamp > FROM_UNIXTIME(?)
		ORDER BY timestamp DESC
		LIMIT 1`
	row := db.QueryRow(stmt, key, id, channel, start)
	var result float64
	err := row.Scan(&result)

	return result, err
}

func (mysql mysql_driver) QueryRows(db *sql.DB, start int64, key string, id string) (*sql.Rows, error) {
	stmt := `SELECT UNIX_TIMESTAMP(timestamp),min,max,avg FROM samples
		WHERE
			key_ = ? AND
			id LIKE ? AND
			timestamp > FROM_UNIXTIME(?)
		ORDER BY timestamp`
	return db.Query(stmt, key, id, start)
}

func (mysql mysql_driver) QueryRowsInterval(db *sql.DB, start int64, key string, id string, interval int64) (*sql.Rows, error) {
	stmt := `SELECT
			CAST(UNIX_TIMESTAMP(timestamp)/? as INTEGER) * ? as ts,
			MIN(min),
			MAX(max),
			AVG(avg)
		FROM samples
		WHERE
			key_ = ? AND
			id LIKE ? AND
			timestamp > FROM_UNIXTIME(?)
		GROUP BY ts
		ORDER BY ts`
	return db.Query(stmt, interval, interval, key, id, start)
}
