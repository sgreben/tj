package color

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type rgb struct{ r, g, b uint8 }

// Scale is a color scale, a function mapping [0,1] to rgb colors.
type Scale func(float64) (r, g, b uint8)

func index(r, g, b uint8) int {
	ri := (int(r) * 5) / 0xFF
	gi := (int(g) * 5) / 0xFF
	bi := (int(b) * 5) / 0xFF
	return 36*ri + 6*gi + bi + 16
}

func clamp(c float64) float64 {
	if c < 0 {
		c = 0
	}
	if c > 1 {
		c = 1
	}
	return c
}

var notHexChars = regexp.MustCompile("[^0-9a-fA-F]")
var spaces = regexp.MustCompile("\\s+")

func parse3(s string, c *rgb) {
	r, _ := strconv.ParseUint(s[0:1], 16, 8)
	c.r = uint8((r << 4) | r)
	g, _ := strconv.ParseUint(s[1:2], 16, 8)
	c.g = uint8((g << 4) | g)
	b, _ := strconv.ParseUint(s[2:3], 16, 8)
	c.b = uint8((b << 4) | b)
}

func parse6(s string, c *rgb) {
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	c.r = uint8(r)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	c.g = uint8(g)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	c.b = uint8(b)
}

// ParseScale parses a sequence of hex colors as a Scale
func ParseScale(scale string) Scale {
	hexOnly := notHexChars.ReplaceAllString(scale, " ")
	singleSpaced := spaces.ReplaceAllString(hexOnly, " ")
	trimmed := strings.TrimSpace(singleSpaced)
	lowercase := strings.ToLower(trimmed)
	parts := strings.Split(lowercase, " ")

	colors := make([]rgb, len(parts))
	for i, s := range parts {
		switch len(s) {
		case 3:
			parse3(s, &colors[i])
		case 6:
			parse6(s, &colors[i])
		}
	}
	return func(c float64) (r, g, b uint8) {
		return interpolate(c, colors)
	}
}

func interpolate2(c float64, r1, g1, b1, r2, g2, b2 uint8) (r, g, b uint8) {
	c = clamp(c)
	r = uint8(float64(r1)*(1-c) + float64(r2)*c)
	g = uint8(float64(g1)*(1-c) + float64(g2)*c)
	b = uint8(float64(b1)*(1-c) + float64(b2)*c)
	return
}

func interpolate(c float64, points []rgb) (r, g, b uint8) {
	c = clamp(c)
	x := float64(len(points)-1) * c
	i := int(x)
	left := points[i]
	j := int(x + 1)
	if j >= len(points) {
		j = i
	}
	right := points[j]
	c = x - float64(i)
	return interpolate2(c, left.r, left.g, left.b, right.r, right.g, right.b)
}

// Foreground returns the closest matching terminal foreground color escape sequence
func Foreground(r, g, b uint8) string {
	return fmt.Sprintf("\033[38;5;%dm", index(r, g, b))
}

// Background returns the closest matching terminal background color escape sequence
func Background(r, g, b uint8) string {
	return fmt.Sprintf("\033[48;5%dm", index(r, g, b))
}

// Reset is the color reset terminal escape sequence
const Reset = "\033[0;00m"
