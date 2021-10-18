package main

import (
	"log"
	"math/rand"
	"time"
)

func Remove(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

func IsEnoughTimePassed(lastDate time.Time, gap time.Duration) bool {
	now := time.Now()
	diff := now.Sub(lastDate)
	return diff > gap
}

func ItIsMondayMyDudes() bool {
	return time.Now().Weekday().String() == "Monday"
}

func GetStringsOfSlice(coll []string) string {
	result := coll[0]
	for i := 1; i < len(coll); i++ {
		result = result + ", " + coll[i]
	}
	return result
}

func PopCandidate(allCandidates []string) (string, []string) {
	now := time.Now()
	rand.Seed(now.UnixNano())
	i := rand.Intn(len(allCandidates))
	chosenCandidate := allCandidates[i]
	allCandidates = Remove(allCandidates, i)
	log.Printf("Chosen candidate is %s\n", chosenCandidate)
	log.Printf("Remaining eligible candidates are %v\n", allCandidates)
	return chosenCandidate, allCandidates
}
