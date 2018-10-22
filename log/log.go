package log

import (
  "github.com/fatih/color"
  "fmt"
)

func printMessage(symbol, s string, colorFunc func(a ...interface{}) string) {
  fmt.Printf("[%s] %s\n", colorFunc(symbol), colorFunc(s))
}

func Error(msg string) {
    printMessage("*", msg, color.New(color.FgRed).SprintFunc())
}

func Info(msg string) {
    printMessage("+", msg, color.New(color.FgBlue).SprintFunc())
}

func Warn(msg string) {
    printMessage("!", msg, color.New(color.FgYellow).SprintFunc())
}

func Debug(msg string) {
    printMessage("?", msg, color.New(color.FgCyan).SprintFunc())
}

