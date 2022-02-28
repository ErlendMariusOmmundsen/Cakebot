package main

import (
	"log"
	"math/rand"
	"time"
)

func Remove(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

func GetIndexInSlice(s []string, str string) int {
	for i := 0; i < len(s); i++ {
		if str == s[i] {
			return i
		}
	}
	return -1
}

func IsEnoughTimePassed(lastDate time.Time, gap time.Duration) bool {
	now := time.Now()
	diff := now.Sub(lastDate)
	return diff > gap
}

func Contains(s []string, str string) bool {
	for _, a := range s {
		if a == str {
			return true
		}
	}
	return false
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
	log.Printf("Updated candidate pool: %v\n", allCandidates)
	return chosenCandidate, allCandidates
}

func resetCandidates(currentPool []string, candidates []string) []string {
	currentPool = nil
	for i := range candidates {
		currentPool = append(currentPool, candidates[i])
	}
	return currentPool
}
