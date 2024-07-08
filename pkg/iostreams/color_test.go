package iostreams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorFromRGB(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		text  string
		wants string
		cs    *ColorScheme
	}{
		{
			name:  "truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "\033[38;2;252;3;3mred\033[0m",
			cs:    NewColorScheme(true, true, true),
		},
		{
			name:  "no truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(true, true, false),
		},
		{
			name:  "no color",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(false, false, false),
		},
		{
			name:  "invalid hex",
			hex:   "fc0",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(false, false, false),
		},
	}

	for _, tt := range tests {
		fn := tt.cs.ColorFromRGB(tt.hex)
		assert.Equal(t, tt.wants, fn(tt.text))
	}
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name  string
		hex   string
		text  string
		wants string
		cs    *ColorScheme
	}{
		{
			name:  "truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "\033[38;2;252;3;3mred\033[0m",
			cs:    NewColorScheme(true, true, true),
		},
		{
			name:  "no truecolor",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(true, true, false),
		},
		{
			name:  "no color",
			hex:   "fc0303",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(false, false, false),
		},
		{
			name:  "invalid hex",
			hex:   "fc0",
			text:  "red",
			wants: "red",
			cs:    NewColorScheme(false, false, false),
		},
	}

	for _, tt := range tests {
		output := tt.cs.HexToRGB(tt.hex, tt.text)
		assert.Equal(t, tt.wants, output)
	}
}

func TestHexToRGBInvalidHex(t *testing.T) {
	cs := NewColorScheme(true, true, true)
	output := cs.HexToRGB("XYZ", "text")
	assert.Equal(t, "text", output)
}

func TestCScheme(t *testing.T) {
	cs := NewColorScheme(true, true, true)
	assert.True(t, cs.Enabled())

	csDisabled := NewColorScheme(false, false, false)
	assert.False(t, csDisabled.Enabled())

	testCases := []struct {
		name     string
		input    string
		expected string
		fn       func(string) string
	}{
		{"Bold", "text", "\x1b[0;1;39mtext\x1b[0m", cs.Bold},
		{"Red", "text", "\x1b[0;31mtext\x1b[0m", cs.Red},
		{"Yellow", "text", "\x1b[0;33mtext\x1b[0m", cs.Yellow},
		{"Green", "text", "\x1b[0;32mtext\x1b[0m", cs.Green},
		{"GreenBold", "text", "\x1b[0;1;32mtext\x1b[0m", cs.GreenBold},
		{"Gray", "text", "\x1b[38;5;242mtext\x1b[m", cs.Gray},
		{"LightGrayUnderline", "text", "\x1b[0;2;4;37mtext\x1b[0m", cs.LightGrayUnderline},
		{"Magenta", "text", "\x1b[0;35mtext\x1b[0m", cs.Magenta},
		{"Cyan", "text", "\x1b[0;36mtext\x1b[0m", cs.Cyan},
		{"CyanBold", "text", "\x1b[0;1;36mtext\x1b[0m", cs.CyanBold},
		{"Blue", "text", "\x1b[0;34mtext\x1b[0m", cs.Blue},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.fn(tc.input))
			assert.Equal(t, tc.input, csDisabled.ColorFromString(tc.name)(tc.input))
		})
	}

	// Test format functions
	assert.Equal(t, "\x1b[0;1;39mHello, World!\x1b[0m", cs.Boldf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;31mHello, World!\x1b[0m", cs.Redf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;33mHello, World!\x1b[0m", cs.Yellowf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;32mHello, World!\x1b[0m", cs.Greenf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[38;5;242mHello, World!\x1b[m", cs.Grayf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;35mHello, World!\x1b[0m", cs.Magentaf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;36mHello, World!\x1b[0m", cs.Cyanf("Hello, %s!", "World"))
	assert.Equal(t, "\x1b[0;34mHello, World!\x1b[0m", cs.Bluef("Hello, %s!", "World"))

	// Test icons
	assert.Equal(t, "\x1b[0;32m✓\x1b[0m", cs.SuccessIcon())
	assert.Equal(t, "\x1b[0;33m!\x1b[0m", cs.WarningIcon())
	assert.Equal(t, "\x1b[0;31mX\x1b[0m", cs.FailureIcon())

	// Test ColorFromString
	assert.Equal(t, "\x1b[0;1;39mtext\x1b[0m", cs.ColorFromString("bold")("text"))
	assert.Equal(t, "text", cs.ColorFromString("invalid")("text"))

	// Test SuccessIconWithColor and FailureIconWithColor
	assert.Equal(t, "\x1b[0;34m✓\x1b[0m", cs.SuccessIconWithColor(cs.Blue))
	assert.Equal(t, "\x1b[0;34mX\x1b[0m", cs.FailureIconWithColor(cs.Blue))
}
