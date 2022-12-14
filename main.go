package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
)

// ConvertCmd is the CLI command
type ConvertCmd struct {
	Sentence   string            `arg:"" required:"" help:"The sentence to convert (letter replacements are case-insensitive)"`
	Pattern    string            `short:"p" enum:"letter,word,yellow,white" default:"letter" help:"Alternating colour pattern to use for the alphabet (letter, word, yellow, white)"`
	Override   map[string]string `short:"o" help:"Override or add emojis for specific characters (e.g. 4=four)"`
	SpaceEmoji string            `name:"space" default:"star2" help:"Emoji to space words with"`
	HeadEmoji  string            `name:"head" help:"Emoji to start with"`
	TailEmoji  string            `name:"tail" help:"Emoji to end with"`
}

const (
	letterPattern = "letter"
	wordPattern   = "word"
	yellowPattern = "yellow"
	whitePattern  = "white"
)

var (
	// ErrNoMatches is returned when the sentence is only whitespace
	ErrNoMatches = errors.New("plz give at least one word")
)

var cli struct {
	Convert ConvertCmd `cmd:"" default:"withargs" help:"Convert a sentence to use emoji letters & spacing"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Description("Annoy your coworkers with emoji messages in Slack"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

// Run runs the command and outputs to console
func (c ConvertCmd) Run() error {
	out, err := c.Convert()
	if err != nil {
		return fmt.Errorf("converting sentence: %w", err)
	}
	fmt.Println(out)
	return nil
}

// Convert takes the command arg & flag options to turn a sentence into a string of Slack emojis
func (c ConvertCmd) Convert() (string, error) {
	// negated whitespace match so that words are split
	re := regexp.MustCompile(`\S+`)
	words := re.FindAllString(c.Sentence, -1)
	length := len(words)
	if length < 1 {
		return "", ErrNoMatches
	}

	if c.SpaceEmoji != "" {
		c.SpaceEmoji = trimEmoji(c.SpaceEmoji)
	}

	if c.HeadEmoji != "" {
		c.HeadEmoji = trimEmoji(c.HeadEmoji)
	}

	if c.TailEmoji != "" {
		c.TailEmoji = trimEmoji(c.TailEmoji)
	}

	// characters not supported by the default alphabet are mapped to specific emojis here before applying overrides
	dict := c.getDictionary()
	letters := 0
	last := words[length-1]

	var b strings.Builder
	writeEmoji(&b, c.HeadEmoji)

	for i, word := range words {
		for _, char := range word {
			// convert uppercase letters to lowercase
			if char >= 65 && char <= 90 {
				char += 32
			}

			// lookup characters in the dictionary, falling back to the alphabet
			emoji, ok := dict[string(char)]
			if !ok {
				var colour string
				switch c.Pattern {
				case letterPattern:
					colour = whichColour(letters)
				case wordPattern:
					colour = whichColour(i)
				default:
					colour = c.Pattern
				}
				emoji = alphabetEmoji(string(char), colour)
			}
			if emoji == "" {
				continue
			}

			writeEmoji(&b, emoji)
			letters += 1
		}

		if word != last {
			writeEmoji(&b, c.SpaceEmoji)
		}
	}

	writeEmoji(&b, c.TailEmoji)

	return b.String(), nil
}

func (c ConvertCmd) getDictionary() map[string]string {
	dict := map[string]string{}
	for k, v := range c.Override {
		dict[strings.ToLower(k)] = trimEmoji(v)
	}
	return dict
}

func trimEmoji(emoji string) string {
	return strings.TrimPrefix(strings.TrimSuffix(emoji, ":"), ":")
}

func writeEmoji(b *strings.Builder, emoji string) {
	if emoji == "" {
		return
	}
	b.WriteString(fmt.Sprintf(":%s:", emoji))
}

func alphabetEmoji(char, colour string) string {
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

func whichColour(i int) string {
	if i%2 == 0 {
		return whitePattern
	}
	return yellowPattern
}
