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
	ID int	
	FamilyID int	// current/primary family (father's, by rule)
	BirthFamilyID int	// family born into; differs only for someone who married in
	FirstName string
	LastName string
	Gender string
	BirthDate *time.Time // "YYYY-MM-DD" format, empty = unknown
	DeathDate *time.Time // "YYYY-MM-DD" format, empty = still alive
	Address string
	MotherID int	// 0 = unknown
	FatherID int	//0 = unknown
	Adopted bool
	Notes string
}