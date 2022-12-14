package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCmd_Convert(t *testing.T) {
	tests := map[string]struct {
		cmd         ConvertCmd
		expected    string
		expectedErr error
	}{
		"writes letters with alternating colours with whitespace between words": {
			cmd: ConvertCmd{
				Sentence: []string{"test !?#@"},
				Pattern:  letterPattern,
			},
			expected: ":alphabet-white-t::alphabet-yellow-e::alphabet-white-s::alphabet-yellow-t:    :alphabet-white-exclamation::alphabet-yellow-question::alphabet-white-hash::alphabet-yellow-at:",
		},
		"writes words with alternating colours with emojis between them": {
			cmd: ConvertCmd{
				Sentence:   []string{"te st"},
				Pattern:    wordPattern,
				SpaceEmoji: "catjam",
			},
			expected: ":alphabet-white-t::alphabet-white-e::catjam::alphabet-yellow-s::alphabet-yellow-t:",
		},
		"writes all yellow colour": {
			cmd: ConvertCmd{
				Sentence: []string{"test"},
				Pattern:  yellowPattern,
			},
			expected: ":alphabet-yellow-t::alphabet-yellow-e::alphabet-yellow-s::alphabet-yellow-t:",
		},
		"writes all white colour": {
			cmd: ConvertCmd{
				Sentence: []string{"test"},
				Pattern:  whitePattern,
			},
			expected: ":alphabet-white-t::alphabet-white-e::alphabet-white-s::alphabet-white-t:",
		},
		"adds head and tail emojis": {
			cmd: ConvertCmd{
				Sentence:  []string{"a"},
				Pattern:   letterPattern,
				HeadEmoji: "catjam",
				TailEmoji: "catjammer",
			},
			expected: ":catjam::alphabet-white-a::catjammer:",
		},
		"overrides characters with different emojis": {
			cmd: ConvertCmd{
				Sentence: []string{"t$"},
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
				Pattern:  letterPattern,
			},
			expected: ":alphabet-white-t:",
		},
		"supports multiple positional arguments, ignoring unsupported symbols": {
			cmd: ConvertCmd{
				Sentence: []string{"te ", "±", "s$t"},
				Pattern:  wordPattern,
			},
			expected: ":alphabet-white-t::alphabet-white-e:    :alphabet-yellow-s::alphabet-yellow-t:",
		},
		"errors when no words are provided": {
			expectedErr: ErrNoMatches,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.cmd.Convert()
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func BenchmarkConvertCmd_Convert(b *testing.B) {
	benches := map[string]ConvertCmd{
		"basic": {
			Sentence:   []string{"The quick brown fox jumps over the lazy dog"},
			Pattern:    wordPattern,
			SpaceEmoji: "dog",
		},
		"extreme": {
			Sentence: []string{"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc scelerisque lobortis urna, eget convallis turpis eleifend a. Aliquam mollis pharetra quam. Integer ac velit at velit posuere euismod in non mauris. Nulla sem leo, bibendum vel facilisis a, convallis ut enim. Sed egestas eget metus in dignissim. Mauris sollicitudin mauris nec velit congue, pretium luctus justo cursus. Aliquam convallis felis vel commodo rhoncus. Nam laoreet ornare molestie."},
			Pattern:  letterPattern,
			Override: map[string]string{
				".": "full_moon",
				",": "sweat_drops",
			},
			SpaceEmoji: "latin_cross",
			HeadEmoji:  "catjam",
			TailEmoji:  "catjammer",
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
