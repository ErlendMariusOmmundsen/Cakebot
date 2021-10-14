package main

import (
	"log"
	"math/rand"
	"time"
)

func GetAllCandidates() []string  {
	return []string{"Hans", "Therese", "Sigurd", "Trym", "Sivert, Asbjørn"}
}

func GetCandidate(lastDate time.Time, gap time.Duration) string {
	now := time.Now()
	diff := now.Sub(lastDate)
	if diff > gap {
		rand.Seed(now.UnixNano())
		i := rand.Intn(6)
		allCandidates := GetAllCandidates()
		cand := allCandidates[i]
		log.Printf("Chosen candidate is %s\n", cand)
		return cand
	}
	return "Trym"
}
