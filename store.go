package main

import (
	"database/sql"
	"errors"
	"fmt"
)

//Store represents a storage layer for families and people, backed by a PostgreSQL database.
type Store struct {
	db *sql.DB
}

//NewStore initializes and returns a new Store instance with the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

//AddFamily creates a new family with the given name and notes, assigns it a unique ID, and stores it in the Store. It returns the newly created Family.
func (s *Store) AddFamily(name, notes string) (Family, error) {
	var id int
	err := s.db.QueryRow(
		`INSERT INTO families (name, notes) VALUES ($1, $2) RETURNING id`,
		name, notes,
	).Scan(&id)
	if err != nil {
		return Family{}, fmt.Errorf("inserting family: %w", err)
	}
	return Family{ID: id, Name: name, Notes: notes}, nil
}

//findByID searches for a person by their ID within the Database. It returns a pointer to the Person and an error if not found.
func (s *Store) findByID(id int) (*Person, error) {
	query := `
		SELECT id,
       COALESCE(family_id, 0),
       COALESCE(birth_family_id, 0),
       first_name,
       COALESCE(last_name, ''),
       COALESCE(gender, ''),
       birth_date,
       death_date,
       COALESCE(address, ''),
       COALESCE(mother_id, 0),
       COALESCE(father_id, 0),
       adopted,
       COALESCE(notes, '')
		FROM people WHERE id = $1`

	var p Person
	err := s.db.QueryRow(query, id).Scan(
		&p.ID, &p.FamilyID, &p.BirthFamilyID, &p.FirstName, &p.LastName, &p.Gender,
		&p.BirthDate, &p.DeathDate, &p.Address, &p.MotherID, &p.FatherID, &p.Adopted, &p.Notes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("person with ID %d not found", id)
		}
		return nil, fmt.Errorf("querying person: %w", err)
	}
	return &p, nil
}

//AddPerson adds a new person to the Store, determining their FamilyID based on the patrilineal rule or explicit assignment if no parents are known. It returns a pointer to the newly added Person and an error if any issues occur.
func (s *Store) AddPerson(p Person) (*Person, error) {
	// Determine FamilyID based on the patrilineal rule
	// If the person has a father, inherit the FamilyID from the father (patrilineal rule).
	// If no father but a mother exists, inherit the FamilyID from the mother.
	// If no parents are known, the FamilyID must be explicitly set by the caller.
	if p.FatherID != 0 {
		father, err := s.findByID(p.FatherID)
		if err != nil {
			return nil, fmt.Errorf("looking up father : %w", err)
		}
		p.FamilyID = father.FamilyID
	} else if p.MotherID != 0 {
		mother, err := s.findByID(p.MotherID)
		if err != nil {
			return nil, fmt.Errorf("looking up mother : %w", err)
		}
		p.FamilyID = mother.FamilyID
	} else {
		// No parents — a root. FamilyID must be set explicitly by the caller.
		if p.FamilyID == 0 {
			return nil, fmt.Errorf("a person with no parents must have a FamilyID set")
		}
	}

	// Insert into the database and get back the generated ID
	query := `
		INSERT INTO people (family_id, birth_family_id, first_name, last_name, gender, birth_date, death_date, address, mother_id, father_id, adopted, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	var newID int
	err := s.db.QueryRow(
		query,
		p.FamilyID,
		nullableInt(p.BirthFamilyID),
		p.FirstName,
		p.LastName,
		p.Gender,
		nullableTime(p.BirthDate),
		nullableTime(p.DeathDate),
		p.Address,
		nullableInt(p.MotherID),
		nullableInt(p.FatherID),
		p.Adopted,
		p.Notes,
	).Scan(&newID)

	if err != nil {
		return nil, fmt.Errorf("inserting person: %w", err)
	}

	p.ID = newID
	return &p, nil
}

//Parents retrieves the parents of a given person (by personID) from the Store. It returns a slice of Person structs representing the parents and an error if the person is not found.
func (s *Store) Parents(personID int) ([]Person, error) {
	person, err := s.findByID(personID)
	if err != nil {
		return nil, err
	}

	var parents []Person
	if person.MotherID != 0 {
		if mother, err := s.findByID(person.MotherID); err == nil {
			parents = append(parents, *mother)
		}
	}
	if person.FatherID != 0 {
		if father, err := s.findByID(person.FatherID); err == nil {
			parents = append(parents, *father)
		}
	}
	return parents, nil
}



//Children retrieves all children of a given parent (by parentID) from the database. It returns a slice of Person structs representing the children and an error if the query fails.
func (s *Store) Children(parentID int) ([]Person, error) {
	query := `
		SELECT id,
       COALESCE(family_id, 0),
       COALESCE(birth_family_id, 0),
       first_name,
       COALESCE(last_name, ''),
       COALESCE(gender, ''),
       birth_date,
       death_date,
       COALESCE(address, ''),
       COALESCE(mother_id, 0),
       COALESCE(father_id, 0),
       adopted,
       COALESCE(notes, '')
		FROM people WHERE mother_id = $1 OR father_id = $1`

	rows, err := s.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("querying children: %w", err)
	}
	defer rows.Close()

	var children []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(
			&p.ID, &p.FamilyID, &p.BirthFamilyID, &p.FirstName, &p.LastName, &p.Gender,
			&p.BirthDate, &p.DeathDate,
			&p.Address,
			&p.MotherID, &p.FatherID, &p.Adopted, &p.Notes,
		); err != nil {
			return nil, fmt.Errorf("scanning child: %w", err)
		}
		children = append(children, p)
	}
	return children, rows.Err()
}

//Siblings retrieves all siblings of a given person (by personID) from the Store. It returns a slice of Person structs representing the siblings and an error if the person is not found.
func (s *Store) Siblings(personID int) ([]Person, error) {
	person, err := s.findByID(personID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id,
       COALESCE(family_id, 0),
       COALESCE(birth_family_id, 0),
       first_name,
       COALESCE(last_name, ''),
       COALESCE(gender, ''),
       birth_date,
       death_date,
       COALESCE(address, ''),
       COALESCE(mother_id, 0),
       COALESCE(father_id, 0),
       adopted,
       COALESCE(notes, '')
		FROM people WHERE id <> $1
		  AND ( ($2 <> 0 AND mother_id = $2) OR ($3 <> 0 AND father_id = $3) )`

	rows, err := s.db.Query(query, personID, person.MotherID, person.FatherID)
	if err != nil {
		return nil, fmt.Errorf("querying siblings: %w", err)
	}
	defer rows.Close()

	var siblings []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(
			&p.ID, &p.FamilyID, &p.BirthFamilyID, &p.FirstName, &p.LastName, &p.Gender,
			&p.BirthDate, &p.DeathDate, &p.Address,
			&p.MotherID, &p.FatherID, &p.Adopted, &p.Notes,
		); err != nil {
			return nil, fmt.Errorf("scanning sibling: %w", err)
		}
		siblings = append(siblings, p)
	}
	return siblings, rows.Err()
}
