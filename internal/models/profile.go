package models

import "time"

type Profile struct {
	ID        int64
	Name      string
	BirthDate time.Time
}
