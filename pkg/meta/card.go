package meta

import "time"

type Card struct {
	Key           string
	Attempt       int
	Quality       int
	Efactor       float32
	DaysInterval  int
	NextRehearsal time.Time
	Hash          string
}
