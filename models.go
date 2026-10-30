package main

import "time"

//Family is a distinct family unit
type Family struct{
	ID int
	Name string
	Notes string
}

//Person is an individual within a family tree
type Person struct{
	ID int `json:"id"`
	FamilyID int `json:"family_id"` // current/primary family (father's, by rule)
	BirthFamilyID int `json:"birth_family_id"` // family born into; differs only for someone who married in
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Gender string `json:"gender"`
	BirthDate *time.Time `json:"birth_date,omitempty"` // "YYYY-MM-DD" format, empty = unknown
	DeathDate *time.Time `json:"death_date,omitempty"` // "YYYY-MM-DD" format, empty = still alive
	Address string `json:"address"`
	MotherID int `json:"mother_id"` // 0 = unknown
	FatherID int `json:"father_id"` //0 = unknown
	Adopted bool `json:"adopted"`
	Notes string `json:"notes"`
}