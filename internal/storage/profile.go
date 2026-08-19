package storage

import (
	"database/sql"
	"errors"
	"time"

	"life/internal/models"
)

func (d *Database) GetProfile() (*models.Profile, error) {
	var profile models.Profile
	var birthDate string
	err := d.DB.QueryRow(`SELECT id, name, birth_date FROM profiles WHERE id = 1`).Scan(
		&profile.ID,
		&profile.Name,
		&birthDate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	profile.BirthDate, err = time.ParseInLocation("2006-01-02", birthDate, time.Local)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (d *Database) SaveProfile(profile models.Profile) error {
	_, err := d.DB.Exec(`INSERT INTO profiles (id, name, birth_date)
		VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, birth_date=excluded.birth_date`,
		profile.Name,
		profile.BirthDate.Format("2006-01-02"),
	)
	return err
}
