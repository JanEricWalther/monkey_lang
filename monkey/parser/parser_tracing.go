package parser

import (
	"fmt"
	"strings"
)

var traceLevel uint

const traceIndentPlaceholder string = "\t"
const debug = false

func Identlevel() string {
	return strings.Repeat(traceIndentPlaceholder, int(traceLevel-1))
}

func tracePrint(formatString string) {
	if !debug {
		return
	}
	fmt.Printf("%s%s\n", Identlevel(), formatString)
}

func trace(msg string) string {
	traceLevel += 1
	tracePrint("BEGIN " + msg)
	return msg
}

func untrace(msg string) {
	tracePrint("END " + msg)
	traceLevel -= 1
}
