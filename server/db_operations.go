package main

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
)

type Implant struct {
	ID        uuid.UUID
	IpAddress string
	LastSeen  time.Time
	CreatedAt time.Time
}

//go:embed schema.sql
var createTableSQL string

func createTables(db *sql.DB) error {
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

func insertImplant(db *sql.DB, id uuid.UUID, ipAddress string, lastSeen, createdAt time.Time) error {
	_, err := db.Exec(
		"INSERT INTO implants VALUES (?, ?, ?, ?)",
		id, ipAddress, lastSeen, createdAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func listImplants(db *sql.DB) ([]Implant, error) {
	implants := make([]Implant, 0)

	rows, err := db.Query("SELECT id, ip_address, last_seen, created_at FROM implants")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var implant Implant

		if err := rows.Scan(&implant.ID, &implant.IpAddress, &implant.LastSeen, &implant.CreatedAt); err != nil {
			return nil, err
		}
		implants = append(implants, implant)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return implants, nil
}

func updateLastSeen(db *sql.DB, implantId string) error {
	_, err := db.Exec(
		"UPDATE implants SET last_seen = ? WHERE id = ?",
		time.Now(), implantId,
	)
	if err != nil {
		return err
	}

	return nil
}

func deleteImplant(db *sql.DB, implantId string) error {
	_, err := db.Exec(
		"DELETE FROM implants WHERE id = ?",
		implantId,
	)
	if err != nil {
		return err
	}

	return nil
}
