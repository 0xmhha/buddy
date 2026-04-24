package format_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wm-it-22-00661/buddy/internal/format"
)

func TestDuration_Boundaries(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{-1, "0ms"}, // negative clamps to zero
		{0, "0ms"},
		{1, "1ms"},
		{123, "123ms"},
		{999, "999ms"},
		{1000, "1.0s"},
		{1234, "1.2s"},
		{5_000, "5.0s"},
		{59_949, "59.9s"}, // last ms whose tenths-rounding stays under 60.0s
		{59_950, "1.0m"},  // promotes to minutes once tenths-of-second would render 60.0s
		{59_999, "1.0m"},  // boundary fix: previously rendered as "60.0s"
		{60_000, "1.0m"},
		{125_000, "2.1m"},
		{600_000, "10.0m"},
	}
	for _, c := range cases {
		t.Run(strconv.FormatInt(c.in, 10), func(t *testing.T) {
			assert.Equal(t, c.want, format.Duration(c.in))
		})
	}
}

func TestThousands_Boundaries(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{9, "9"},
		{99, "99"},
		{999, "999"},
		{1_000, "1,000"},
		{1_234, "1,234"},
		{12_345, "12,345"},
		{123_456, "123,456"},
		{1_234_567, "1,234,567"},
		{1_000_000, "1,000,000"},
		{-1_234, "-1,234"},
		{-1, "-1"},
	}
	for _, c := range cases {
		t.Run(strconv.FormatInt(c.in, 10), func(t *testing.T) {
			assert.Equal(t, c.want, format.Thousands(c.in))
		})
	}
}
