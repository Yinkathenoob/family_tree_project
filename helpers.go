package main

import (
	"time"
)
//nullableInt converts an integer ID to an interface{} suitable for SQL queries, returning nil if the ID is 0.
func nullableInt(id int) interface{} {
	if id == 0 {
		return nil
	}
	return id
}

//nullableTime converts a *time.Time to an interface{} suitable for SQL queries, returning nil if the time is nil.
func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil // SQL NULL
	}
	return *t // the actual time value
}
