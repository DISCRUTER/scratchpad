package models

import (
	"database/sql"
	"errors"
	"time"
)

type Pads struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type PadsModel struct {
	DB *sql.DB
}

// Insert new Pad in DB and return ID
func (m * PadsModel) Insert(title string, content string, expires int) (int, error) {
	// Raw SQL insert statement
	stmt := `INSERT INTO pads (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	// Query execution with data
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	// ID fetch for the inserted row
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	return int(id), nil
}

// Return a specific Pad via the id
func (m *PadsModel) Get(id int) (Pads, error) {
	stmt := `SELECT id, title, content, created, expires FROM pads
	WHERE expires > UTC_TIMESTAMP() AND id = ?`
	// QueryRow retuen a pointer to the sql row
	row := m.DB.QueryRow(stmt, id)
	// Fetch the value from row
	var p Pads
	// Scan the row to get all the required values
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.Created, &p.Expires)
	if err != nil {
		// Error if no records found
		if errors.Is(err, sql.ErrNoRows) {
			return Pads{}, ErrNoRecord
		} else {
			return Pads{}, err
		}
	}
	return p ,nil
}

// This will the recent 10 Pads
func (m *PadsModel) Latest() ([]Pads, error) {
	// SQL statement
	stmt := `SELECT id, title, content, created, expires FROM pads
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`
	// Query database
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Creating pads array
	var pads []Pads
	for rows.Next() {
		// Creating a pads struct
		var p Pads
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Created, &p.Expires); err != nil {
			return nil, err
		}
		// Append to the array
		pads = append(pads, p)
	}
	// Check if any error is stored
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pads, nil
}
