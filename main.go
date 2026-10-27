package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	_ "github.com/jackc/pgx/v5/stdlib"
)


func main() {
	// 1. Connect to the database
	db, err := sql.Open("pgx", "postgres://Mac@localhost:5432/family_tree")
	if err != nil {
		log.Fatal("open:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("ping:", err)
	}

	store := NewStore(db)

	// 2. Add a person WITH a birth date
	bd := time.Date(1950, 3, 15, 0, 0, 0, 0, time.UTC)
	withDate, err := store.AddPerson(Person{
		FirstName: "Dated",
		Gender:    "male",
		FamilyID:  1,
		BirthDate: &bd,
	})
	if err != nil {
		log.Fatal("add dated person:", err)
	}

	// 3. Add a person WITHOUT a birth date (BirthDate stays nil)
	noDate, err := store.AddPerson(Person{
		FirstName: "Undated",
		Gender:    "female",
		FamilyID:  1,
	})
	if err != nil {
		log.Fatal("add undated person:", err)
	}

	// 4. Read both back from the database
	d1, err := store.findByID(withDate.ID)
	if err != nil {
		log.Fatal("find dated person:", err)
	}
	d2, err := store.findByID(noDate.ID)
	if err != nil {
		log.Fatal("find undated person:", err)
	}

	// 5. Check what came back
	if d1.BirthDate != nil {
		fmt.Printf("%s born: %s\n", d1.FirstName, d1.BirthDate.Format("2006-01-02"))
	} else {
		fmt.Printf("%s: birth date came back nil (expected a date!)\n", d1.FirstName)
	}

	if d2.BirthDate == nil {
		fmt.Printf("%s has no recorded birth date\n", d2.FirstName)
	} else {
		fmt.Printf("%s: unexpected birth date %s\n", d2.FirstName, d2.BirthDate.Format("2006-01-02"))
	}

	kids, err := store.Children(1)
	if err != nil {
		log.Fatal("children:", err)
	}
	fmt.Println("Children of person 1:", len(kids))

	sibs, err := store.Siblings(3)
	if err != nil {
		log.Fatal("siblings:", err)
	}
	fmt.Println("Siblings of person 3:", len(sibs))
}
