package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCmd_Convert(t *testing.T) {
	tests := map[string]struct {
		cmd         ConvertCmd
		expected    string
		expectedErr string
	}{
		"writes letters with alternating colours with whitespace between words": {
			cmd: ConvertCmd{
				Sentence: []string{"test !?#@"},
				EmojiSet: alphabetSet,
				Pattern:  letterPattern,
			},
			expected: ":alphabet-white-t::alphabet-yellow-e::alphabet-white-s::alphabet-yellow-t:    :alphabet-white-exclamation::alphabet-yellow-question::alphabet-white-hash::alphabet-yellow-at:",
		},
		"writes words with alternating colours with emojis between them": {
			cmd: ConvertCmd{
				Sentence:   []string{"te st"},
				EmojiSet:   alphabetSet,
				Pattern:    wordPattern,
				SpaceEmoji: "catjam",
			},
			expected: ":alphabet-white-t::alphabet-white-e::catjam::alphabet-yellow-s::alphabet-yellow-t:",
		},
		"writes all yellow colour": {
			cmd: ConvertCmd{
				Sentence: []string{"test"},
				EmojiSet: alphabetSet,
				Pattern:  yellowPattern,
			},
			expected: ":alphabet-yellow-t::alphabet-yellow-e::alphabet-yellow-s::alphabet-yellow-t:",
		},
		"writes all white colour": {
			cmd: ConvertCmd{
				Sentence: []string{"test"},
				EmojiSet: alphabetSet,
				Pattern:  whitePattern,
			},
			expected: ":alphabet-white-t::alphabet-white-e::alphabet-white-s::alphabet-white-t:",
		},
		"writes scrabble letters with a default space emoji": {
			cmd: ConvertCmd{
				Sentence: []string{"te s%t"},
				EmojiSet: scrabbleSet,
			},
			expected: ":scrabble-t::scrabble-e::scrabble-blank::scrabble-s::scrabble-t:",
		},
		"writes alphabet with reaction constraints": {
			cmd: ConvertCmd{
				Sentence: []string{"happy", "birthday"},
				EmojiSet: reactionSet,
			},
			expected: ":alphabet-white-h:\n:alphabet-yellow-a:\n:alphabet-white-p:\n:alphabet-yellow-p:\n:alphabet-white-y:\n:alphabet-yellow-b:\n:alphabet-white-i:\n:alphabet-yellow-r:\n:alphabet-white-t:\n:alphabet-yellow-h:\n:alphabet-white-d:\n:alphabet-white-a:\n:alphabet-yellow-y:\n",
		},
		"adds head and tail emojis": {
			cmd: ConvertCmd{
				Sentence:  []string{"a"},
				EmojiSet:  alphabetSet,
				Pattern:   letterPattern,
				HeadEmoji: "catjam",
				TailEmoji: "catjammer",
			},
			expected: ":catjam::alphabet-white-a::catjammer:",
		},
		"overrides characters with different emojis": {
			cmd: ConvertCmd{
				Sentence: []string{"t$"},
				EmojiSet: alphabetSet,
				Pattern:  letterPattern,
				Override: map[string]string{
					"t": "catjam",
					"$": "money",
				},
			},
			expected: ":catjam::money:",
		},
		"lower-cases letters": {
			cmd: ConvertCmd{
				Sentence: []string{"T"},
				EmojiSet: alphabetSet,
				Pattern:  letterPattern,
			},
			expected: ":alphabet-white-t:",
		},
		"supports multiple positional arguments, ignoring unsupported symbols": {
			cmd: ConvertCmd{
				Sentence: []string{"te ", "±", "s$t"},
				EmojiSet: alphabetSet,
				Pattern:  wordPattern,
			},
			expected: ":alphabet-white-t::alphabet-white-e:    :alphabet-yellow-s::alphabet-yellow-t:",
		},
		"errors when no words are provided": {
			expectedErr: ErrNoMatches.Error(),
		},
		"errors when an unsupported set is given": {
			cmd: ConvertCmd{
				Sentence: []string{"T"},
				EmojiSet: "foo",
			},
			expectedErr: ErrNotSupported.Error(),
		},
		"errors when a character is used past the limit for reactions": {
			cmd: ConvertCmd{
				Sentence: []string{"tttest"},
				EmojiSet: reactionSet,
			},
			expectedErr: `the character "t" at position 3 cannot be used more than 2 times`,
		},
		"errors when sentence length is past the limit for reactions": {
			cmd: ConvertCmd{
				Sentence: []string{"aabbccddeeffgghhiijjkkll"},
				EmojiSet: reactionSet,
			},
			expectedErr: "slack won't let you react with more than 23 emojis :(",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.cmd.Convert()
			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func BenchmarkConvertCmd_Convert(b *testing.B) {
	benches := map[string]ConvertCmd{
		"alphabet - word": {
			Sentence: []string{"The quick brown fox jumps over the lazy dog"},
			EmojiSet: alphabetSet,
			Pattern:  wordPattern,
		},
		"alphabet - letter": {
			Sentence: []string{"The quick brown fox jumps over the lazy dog"},
			EmojiSet: alphabetSet,
			Pattern:  letterPattern,
		},
		"alphabet - colour": {
			Sentence: []string{"The quick brown fox jumps over the lazy dog"},
			EmojiSet: alphabetSet,
			Pattern:  yellowPattern,
		},
		"alphabet - extreme": {
			Sentence: []string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc scelerisque lobortis urna, eget convallis turpis eleifend a.",
				"Aliquam mollis pharetra quam. Integer ac velit at velit posuere euismod in non mauris. Nulla sem leo, bibendum vel facilisis a, convallis ut enim.",
				"Sed egestas eget metus in dignissim. Mauris sollicitudin mauris nec velit congue, pretium luctus justo cursus. Aliquam convallis felis vel commodo rhoncus.",
				"Nam laoreet ornare molestie.",
			},
			EmojiSet: alphabetSet,
			Pattern:  letterPattern,
			Override: map[string]string{
				".": "full_moon",
				",": "sweat_drops",
			},
			SpaceEmoji: "latin_cross",
			HeadEmoji:  "catjam",
			TailEmoji:  "catjammer",
		},
		"scrabble": {
			Sentence: []string{"The quick brown fox jumps over the lazy dog"},
			EmojiSet: scrabbleSet,
		},
		"reaction": {
			Sentence: []string{"Happy birthday best friend"},
			EmojiSet: reactionSet,
		},
	}

	for name, bb := range benches {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bb.Convert()
			}
		})
	}
}
