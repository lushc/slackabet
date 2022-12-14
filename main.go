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
	Sentence   []string          `arg:"" required:"" passthrough:"" help:"The sentence to convert. Letter replacements are case-insensitive, with each word separated by 4 spaces by default"`
	EmojiSet   string            `short:"e" enum:"alphabet,scrabble" default:"alphabet" help:"Specify the emoji-set to use. Defaults to Slack's alphabet which supports patterns"`
	Pattern    string            `short:"p" enum:"letter,word,yellow,white" default:"letter" help:"Alternating colour pattern to use for the alphabet emoji-set (letter, word, yellow, white)"`
	Override   map[string]string `short:"o" help:"Override or add emojis for specific characters (e.g. 4=four)"`
	SpaceEmoji string            `name:"space" help:"Emoji to separate words with instead of whitespace"`
	HeadEmoji  string            `name:"head" help:"Emoji to start the sentence with"`
	TailEmoji  string            `name:"tail" help:"Emoji to end the sentence with"`
	NoCopy     bool              `help:"Print the output to console instead of copying to the clipboard"`
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

	if c.NoCopy {
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

	if c.SpaceEmoji != "" {
		c.SpaceEmoji = trimEmoji(c.SpaceEmoji)
	}

	if c.HeadEmoji != "" {
		c.HeadEmoji = trimEmoji(c.HeadEmoji)
	}

	if c.TailEmoji != "" {
		c.TailEmoji = trimEmoji(c.TailEmoji)
	}

	var strategy emojiStrategy
	switch c.EmojiSet {
	case alphabetSet:
		strategy = newAlphabetStrategy(c.Pattern)
	case scrabbleSet:
		strategy = newScrabbleStrategy(&c)
	default:
		return "", ErrNotSupported
	}

	// characters not supported by the default alphabet are mapped to specific emojis here with applied overrides
	dict := c.getDictionary()

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

			// lookup characters in the dictionary, falling back to the alphabet of the strategy
			emoji, ok := dict[string(char)]
			if !ok {
				emoji = strategy.Get(string(char))
			}
			if emoji == "" {
				continue
			}

			writeEmoji(&b, emoji)
			strategy.LetterCallback()
			committed = true
		}

		if !committed || word == last {
			continue
		}

		if c.SpaceEmoji != "" {
			writeEmoji(&b, c.SpaceEmoji)
		} else {
			b.WriteString("    ")
		}
		strategy.WordCallback()
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
