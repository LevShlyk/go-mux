package main

import (
	"database/sql"
	"lshlyk/case/internal/shorter"
)

type short struct {
	ID     uint64 `json:"id"`
	Source string `json:"source"`
	Short  string `json:"short"`
}

func (s *short) GenerateShort(id uint64) string {
	shorterInstance := shorter.BuildShorter()
	return shorterInstance.GetShortByID(id)
}

func (s *short) getShort(db *sql.DB) error {
	return db.QueryRow("SELECT source, short FROM shorts WHERE id=$1",
		s.ID).Scan(&s.Source, &s.Short)
}

func (s *short) getShortByShort(db *sql.DB) error {
	return db.QueryRow("SELECT id, source, short FROM shorts WHERE short=$1",
		s.Short).Scan(&s.ID, &s.Source, &s.Short)
}

func (s *short) updateShort(db *sql.DB) error {
	s.Short = s.GenerateShort(s.ID)
	_, err :=
		db.Exec("UPDATE shorts SET source=$1, short=$2 WHERE id=$3",
			s.Source, s.Short, s.ID)

	return err
}

func (s *short) deleteShort(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM shorts WHERE id=$1", s.ID)

	return err
}

func (s *short) createShort(db *sql.DB) error {
	needleId := uint64(1)

	err := db.QueryRow("SELECT currval(pg_get_serial_sequence('shorts','id'))").Scan(&needleId)
	if err != nil {
		needleId = 1
	}

	s.Short = s.GenerateShort(needleId)

	err = db.QueryRow(
		"INSERT INTO shorts(source, short) VALUES($1, $2) RETURNING id",
		s.Source, s.Short).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func getShorts(db *sql.DB, start, count int) ([]short, error) {
	rows, err := db.Query(
		"SELECT id, source, short FROM shorts LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := []short{}

	for rows.Next() {
		var s short
		if err := rows.Scan(&s.ID, &s.Source, &s.Short); err != nil {
			return nil, err
		}
		products = append(products, s)
	}

	return products, nil
}