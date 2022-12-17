package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"golang.design/x/clipboard"
)

// ConvertCmd is the CLI command
type ConvertCmd struct {
	Sentence       []string          `arg:"" required:"" passthrough:"" help:"The sentence to convert. Letter replacements are case-insensitive, with each word separated by 4 spaces by default"`
	EmojiSet       string            `short:"e" enum:"alphabet,scrabble,reaction" default:"alphabet" help:"Specify the emoji-set to use. Defaults to Slack's alphabet which supports patterns"`
	Pattern        string            `short:"p" enum:"letter,word,yellow,white" default:"letter" help:"Alternating colour pattern to use for the alphabet emoji-set (letter, word, yellow, white)"`
	Override       map[string]string `short:"o" help:"Override or add emojis for specific characters (e.g. 4=four)"`
	SpaceEmoji     string            `name:"space" help:"Emoji to separate words with instead of whitespace"`
	HeadEmoji      string            `name:"head" help:"Emoji to start the sentence with"`
	TailEmoji      string            `name:"tail" help:"Emoji to end the sentence with"`
	Copy           bool              `default:"true" negatable:"" help:"Copy to the clipboard by default, negating will output to console instead"`
	DefaultSpacing string            `kong:"-"`
}

var (
	// ErrNoMatches is returned when the sentence is only whitespace
	ErrNoMatches = errors.New("plz give at least one word")
	// ErrNotSupported is returned when a matching strategy doesn't exist for the emoji-set
	ErrNotSupported = errors.New("emoji-set not supported")
)

var cli struct {
	Convert ConvertCmd `cmd:"" default:"withargs" help:"Convert a sentence to use emojis, annoying everyone"`
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

	if !c.Copy {
		fmt.Println(out)
		return nil
	}

	if err := clipboard.Init(); err != nil {
		return fmt.Errorf("initialising clipboard: %w", err)
	}

	clipboard.Write(clipboard.FmtText, []byte(out))
	return nil
}

// Convert takes the command arg & flag options to turn a sentence into a string of Slack emojis
func (c ConvertCmd) Convert() (string, error) {
	// negated whitespace match so that words are split
	re := regexp.MustCompile(`\S+`)
	words := re.FindAllString(strings.Join(c.Sentence, " "), -1)
	length := len(words)
	if length < 1 {
		return "", ErrNoMatches
	}

	c.DefaultSpacing = "    "

	if c.SpaceEmoji != "" {
		c.SpaceEmoji = trimEmoji(c.SpaceEmoji)
	}

	if c.HeadEmoji != "" {
		c.HeadEmoji = trimEmoji(c.HeadEmoji)
	}

	if c.TailEmoji != "" {
		c.TailEmoji = trimEmoji(c.TailEmoji)
	}

	strategy, err := c.buildStrategy()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	writeEmoji(&b, c.HeadEmoji)

	last := words[length-1]
	for _, word := range words {
		committed := false
		for _, char := range word {
			// convert uppercase letters to lowercase
			if char >= 65 && char <= 90 {
				char += 32
			}

			emoji, err := strategy.Get(string(char))
			if err != nil {
				return "", err
			}

			if emoji == "" {
				continue
			}

			writeEmoji(&b, emoji)
			strategy.CharacterCallback(&b)
			committed = true
		}

		if !committed || word == last {
			continue
		}

		if c.SpaceEmoji != "" {
			writeEmoji(&b, c.SpaceEmoji)
		} else {
			b.WriteString(c.DefaultSpacing)
		}
		strategy.WordCallback(&b)
	}

	writeEmoji(&b, c.TailEmoji)

	return b.String(), nil
}

func (c *ConvertCmd) buildStrategy() (emojiStrategy, error) {
	var s emojiStrategy
	switch c.EmojiSet {
	case alphabetSet:
		s = &alphabetStrategy{
			pattern: c.Pattern,
		}
	case scrabbleSet:
		// default to a blank tile
		if c.SpaceEmoji == "" {
			c.SpaceEmoji = "scrabble-blank"
		}
		s = scrabbleStrategy{}
	case reactionSet:
		c.DefaultSpacing = ""
		s = &reactionStrategy{
			used: make(map[string][]string),
			max:  23,
		}
	default:
		return s, ErrNotSupported
	}

	overrides := map[string]string{}
	for k, v := range c.Override {
		overrides[strings.ToLower(k)] = trimEmoji(v)
	}

	return &overrideStrategy{
		emojiStrategy: s,
		overrides:     overrides,
	}, nil
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
