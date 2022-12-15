package main

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type emojiStrategy interface {
	Get(char string) (string, error)
	CharacterCallback(b *strings.Builder)
	WordCallback(b *strings.Builder)
}

type alphabetStrategy struct {
	pattern string
	// keep track of committed characters & words for alternating patterns
	writtenCharacters int
	writtenWords      int
}

type scrabbleStrategy struct{}

type reactionStrategy struct {
	// create an alternating pattern where possible up until the limit of available colours
	written int
	used    map[string][]string
	max     int
}

const (
	alphabetSet     = "alphabet"
	scrabbleSet     = "scrabble"
	reactionSet     = "reaction"
	letterPattern   = "letter"
	wordPattern     = "word"
	yellowPattern   = "yellow"
	whitePattern    = "white"
	alphabetColours = 2
)

func newAlphabetStrategy(pattern string) *alphabetStrategy {
	return &alphabetStrategy{
		pattern: pattern,
	}
}

// Get returns an alphabet emoji of the colour that best matches the configured pattern
func (a *alphabetStrategy) Get(char string) (string, error) {
	var colour string
	switch a.pattern {
	case letterPattern:
		colour = whichColour(a.writtenCharacters)
	case wordPattern:
		colour = whichColour(a.writtenWords)
	default:
		colour = a.pattern
	}

	if suffix := alphabetSuffix(char); suffix != "" {
		return fmt.Sprintf("alphabet-%s-%s", colour, suffix), nil
	}

	return "", nil
}

// CharacterCallback keeps track of how many characters have been written for the "letter" alternating pattern
func (a *alphabetStrategy) CharacterCallback(b *strings.Builder) {
	a.writtenCharacters += 1
}

// WordCallback keeps track of how many words have been written for the "word" alternating pattern
func (a *alphabetStrategy) WordCallback(b *strings.Builder) {
	a.writtenWords += 1
}

func newScrabbleStrategy(cmd *ConvertCmd) scrabbleStrategy {
	// default to a blank tile
	if cmd.SpaceEmoji == "" {
		cmd.SpaceEmoji = "scrabble-blank"
	}
	return scrabbleStrategy{}
}

// Get returns a scrabble alphabet emoji
func (s scrabbleStrategy) Get(char string) (string, error) {
	if char < "a" || char > "z" {
		return "", nil
	}
	return fmt.Sprintf("scrabble-%s", char), nil
}

// CharacterCallback is a no-op
func (s scrabbleStrategy) CharacterCallback(b *strings.Builder) {}

// WordCallback is a no-op
func (s scrabbleStrategy) WordCallback(b *strings.Builder) {}

func newReactionStrategy(cmd *ConvertCmd) *reactionStrategy {
	cmd.DefaultSpacing = ""
	return &reactionStrategy{
		used: make(map[string][]string),
		max:  23,
	}
}

// Get returns an alphabet emoji in a colour that has not yet been used. It will error when there are more of the same
// characters in-use than available colours, or when the hard-limit of Slack reactions has been reached
func (r *reactionStrategy) Get(char string) (string, error) {
	if r.written >= r.max {
		return "", fmt.Errorf("slack won't let you react with more than %d emojis :(", r.max)
	}

	suffix := alphabetSuffix(char)
	if suffix == "" {
		return "", nil
	}

	if _, ok := r.used[suffix]; !ok {
		r.used[suffix] = make([]string, alphabetColours)
	}

	// find a colour for the alphabet emoji that's not yet in-use
	for i := 0; i < alphabetColours; i++ {
		colour := whichColour(r.written + i)
		if slices.Index(r.used[suffix], colour) != -1 {
			continue
		}
		r.used[suffix] = append(r.used[suffix], colour)
		return fmt.Sprintf("alphabet-%s-%s", colour, suffix), nil
	}

	return "", fmt.Errorf(`the character "%s" at position %d cannot be used more than %d times`, char, r.written+1, alphabetColours)
}

// CharacterCallback keeps track of emoji characters written, adding a new line after each one to ensure rapid-fire cut-and-paste into the reaction search
func (r *reactionStrategy) CharacterCallback(b *strings.Builder) {
	r.written += 1
	b.WriteString("\n")
}

// WordCallback is a no-op
func (r *reactionStrategy) WordCallback(b *strings.Builder) {}

func whichColour(i int) string {
	if i%alphabetColours == 0 {
		return whitePattern
	}
	return yellowPattern
}

func alphabetSuffix(char string) string {
	switch {
	case char == "!":
		return "exclamation"
	case char == "#":
		return "hash"
	case char == "?":
		return "question"
	case char == "@":
		return "at"
	case char >= "a" && char <= "z":
		return char
	default:
		return ""
	}
}
