package main

import (
	"fmt"
)

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
	BirthDate string // "YYYY-MM-DD" format, empty = unknown
	DeathDate string // "YYYY-MM-DD" format, empty = still alive
	Address string
	MotherID int	// 0 = unknown
	FatherID int	//0 = unknown
	Adopted bool
	Notes string
}

//Store holds all families and people, along with the next IDs to be assigned
type Store struct {
	families     []Family
	people       []Person
	nextFamilyID int
	nextPersonID int
}

//NewStore initializes and returns a new Store instance with starting IDs set to 1
func NewStore() *Store {
	return &Store{
		nextFamilyID: 1,
		nextPersonID: 1,
	}
}

//AddFamily creates a new family with the given name and notes, assigns it a unique ID, and stores it in the Store. It returns the newly created Family.
func (s *Store) AddFamily(name, notes string) Family {
	f := Family{
		ID:    s.nextFamilyID,
		Name:  name,
		Notes: notes,
	}
	s.families = append(s.families, f)
	s.nextFamilyID++
	return f
}

//findByID searches for a person by their ID within the Store. It returns a pointer to the Person and true if found, or nil and false if not found.
func (s *Store) findByID(id int) (*Person, bool) {
	for i := range s.people {
		if s.people[i].ID == id {
			return &s.people[i], true
		}
	}
	return nil, false
}


//AddPerson adds a new person to the Store, determining their FamilyID based on the patrilineal rule or explicit assignment if no parents are known. It returns a pointer to the newly added Person and an error if any issues occur.
func (s *Store) AddPerson(p Person) (*Person, error) {
	// Determine FamilyID based on the patrilineal rule
	// If the person has a father, inherit the FamilyID from the father (patrilineal rule).
	// If no father but a mother exists, inherit the FamilyID from the mother.
	// If no parents are known, the FamilyID must be explicitly set by the caller.
	if p.FatherID != 0 {
		father, ok := s.findByID(p.FatherID)
		if !ok {
			return nil, fmt.Errorf("father with ID %d not found", p.FatherID)
		}
		p.FamilyID = father.FamilyID
	} else if p.MotherID != 0 {
		mother, ok := s.findByID(p.MotherID)
		if !ok {
			return nil, fmt.Errorf("mother with ID %d not found", p.MotherID)
		}
		p.FamilyID = mother.FamilyID
	} else {
		// No parents — a root. FamilyID must be set explicitly by the caller.
		if p.FamilyID == 0 {
			return nil, fmt.Errorf("a person with no parents must have a FamilyID set")
		}
	}

	// Assign ID and store
	p.ID = s.nextPersonID
	s.people = append(s.people, p)
	s.nextPersonID++

	// Retrieve the stored person to return a pointer to it
	stored, ok := s.findByID(p.ID)
	if !ok {
		return nil, fmt.Errorf("internal error: just-added person %d not found", p.ID)
	}
	return stored, nil
}


func (s *Store) Parents(personID int) []Person {
	// 1. find the person
	var parents []Person
	person, ok := s.findByID(personID)
	if !ok {
		return parents
	}
	// 2. if MotherID != 0, look up the mother, append
	if person.MotherID != 0 {
		if mother, ok := s.findByID(person.MotherID); ok {
			parents = append(parents, *mother)
		}
	}
	// 3. if FatherID != 0, look up the father, append
	if person.FatherID != 0 {
		if father, ok := s.findByID(person.FatherID); ok {
			parents = append(parents, *father)
		}
	}
	return parents
}



//Children retrieves all children of a given parent (by parentID) from the Store. It returns a slice of Person structs representing the children.
func (s *Store) Children(parentID int) []Person {
	var children []Person
	for _, p := range s.people {
		if p.MotherID == parentID || p.FatherID == parentID {
			children = append(children, p)
		}
	}
	return children
}

//Siblings retrieves all siblings of a given person (by personID) from the Store. It returns a slice of Person structs representing the siblings and an error if the person is not found.
func (s *Store) Siblings(personID int) ([]Person, error) {
	// First, find the person themselves to know their parents
	person, ok := s.findByID(personID)
	if !ok {
		return nil, fmt.Errorf("person with ID %d not found", personID)
	}

	var siblings []Person
	for _, p := range s.people {
		if p.ID == personID {
			continue // skip the person themselves
		}
		// share a mother (and that mother is actually known, not 0)
		sharesMother := person.MotherID != 0 && p.MotherID == person.MotherID
		// share a father (and that father is actually known, not 0)
		sharesFather := person.FatherID != 0 && p.FatherID == person.FatherID

		if sharesMother || sharesFather {
			siblings = append(siblings, p)
		}
	}
	return siblings, nil
}

func main() {
	store := NewStore()
	family := store.AddFamily("The Adeyemi Family", "Lagos origin")

	// Grandfather (root)
	gf, _ := store.AddPerson(Person{FirstName: "Oludare", Gender: "male", FamilyID: family.ID})

	// Grandmother (root — married in, but we'll keep simple: explicit family)
	gm, _ := store.AddPerson(Person{FirstName: "Folake", Gender: "female", FamilyID: family.ID})

	// Three children of gf + gm (siblings to each other)
	kunle, _ := store.AddPerson(Person{FirstName: "Kunle", Gender: "male", FatherID: gf.ID, MotherID: gm.ID})
	bisi, _ := store.AddPerson(Person{FirstName: "Bisi", Gender: "female", FatherID: gf.ID, MotherID: gm.ID})
	dayo, _ := store.AddPerson(Person{FirstName: "Dayo", Gender: "male", FatherID: gf.ID, MotherID: gm.ID})

	// Test: children of the grandfather
	fmt.Println("Children of Oludare:")
	for _, c := range store.Children(gf.ID) {
		fmt.Printf("  - %s\n", c.FirstName)
	}

	// Test: siblings of Kunle
	fmt.Println("Siblings of Kunle:")
	sibs, err := store.Siblings(kunle.ID)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, s := range sibs {
		fmt.Printf("  - %s\n", s.FirstName)
	}

	_ = bisi // silence unused warnings if needed
	_ = dayo
	
}

