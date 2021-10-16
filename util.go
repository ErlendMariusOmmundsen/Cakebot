package main

import (
	"log"
	"math/rand"
	"time"
)

func GetAllCandidates() []string {
	return []string{"Hans", "Therese", "Sigurd", "Trym", "Sivert", "Asbjørn"}
}

func IsEnoughTimePassed(lastDate time.Time, gap time.Duration) bool {
	now := time.Now()
	diff := now.Sub(lastDate)
	return diff > gap
}

func ItIsMondayMyDudes() bool {
	return time.Now().Weekday().String() == "Monday"
}

func GetCandidate() string {
	now := time.Now()
	rand.Seed(now.UnixNano())
	allCandidates := GetAllCandidates()
	i := rand.Intn(len(allCandidates))
	chosenCandidate := allCandidates[i]
	log.Printf("Chosen candidate is %s\n", chosenCandidate)
	return chosenCandidate
}
