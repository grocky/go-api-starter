package log

import "fmt"

// reference: [https://chrisyeh96.github.io/2020/03/28/terminal-colors.html]
const pre = "\033["
const post = "m"

type code int

const (
	reset code = iota
)

const (
	black code = iota + 30
	red
	green
	yello
	blue
	magenta
	cyan
	lightGray
)

const (
	darkGray code = iota + 90
	lightRed
	lightGreen
	lightYellow
	lightBlue
	lightMagenta
	lightCyan
	white
)

func (c code) String() string {
	return fmt.Sprintf("%s%d%s", pre, c, post)
}

func colorize(colorCode code, v string) string {
	return fmt.Sprintf("%s%s%s", colorCode, v, reset)
}
