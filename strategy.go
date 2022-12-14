package main

import (
	"fmt"
)

type emojiStrategy interface {
	Get(char string) string
	LetterCallback()
	WordCallback()
}

type alphabetStrategy struct {
	pattern string
	// keep track of committed letters & words for alternating patterns
	writtenLetters int
	writtenWords   int
}

type scrabbleStrategy struct{}

const (
	alphabetSet   = "alphabet"
	scrabbleSet   = "scrabble"
	letterPattern = "letter"
	wordPattern   = "word"
	yellowPattern = "yellow"
	whitePattern  = "white"
)

func newAlphabetStrategy(pattern string) *alphabetStrategy {
	return &alphabetStrategy{
		pattern: pattern,
	}
}

func newScrabbleStrategy(cmd *ConvertCmd) scrabbleStrategy {
	// default to a blank tile
	if cmd.SpaceEmoji == "" {
		cmd.SpaceEmoji = "scrabble-blank"
	}
	return scrabbleStrategy{}
}

func (a *alphabetStrategy) Get(char string) string {
	var colour string
	switch a.pattern {
	case letterPattern:
		colour = whichColour(a.writtenLetters)
	case wordPattern:
		colour = whichColour(a.writtenWords)
	default:
		colour = a.pattern
	}

	var suffix string
	switch {
	case char == "!":
		suffix = "exclamation"
	case char == "#":
		suffix = "hash"
	case char == "?":
		suffix = "question"
	case char == "@":
		suffix = "at"
	case char >= "a" && char <= "z":
		suffix = char
	default:
		return ""
	}
	return fmt.Sprintf("alphabet-%s-%s", colour, suffix)
}

func (a *alphabetStrategy) LetterCallback() {
	a.writtenLetters += 1
}

func (a *alphabetStrategy) WordCallback() {
	a.writtenWords += 1
}

func (s scrabbleStrategy) Get(char string) string {
	if char < "a" || char > "z" {
		return ""
	}
	return fmt.Sprintf("scrabble-%s", char)
}

func (s scrabbleStrategy) LetterCallback() {}

func (s scrabbleStrategy) WordCallback() {}

func whichColour(i int) string {
	if i%2 == 0 {
		return whitePattern
	}
	return yellowPattern
}
