package utils

import (
	"log"
)

var current_level = 6

func SetLevel(level int) {
	current_level = level
}

func Log(level int, v ...interface{}) {
	if current_level <= level {
		log.Println(v...)
	}
}

func Logf(level int, format string, v ...interface{}) {
	if current_level <= level {
		log.Printf(format, v...)
	}
}

func IfLog(level int, b bool, v ...interface{}) {
	if b {
		Log(level, v...)
	}
}
func IfLogf(level int, b bool, format string, v ...interface{}) {
	if b {
		Logf(level, format, v...)
	}
}

func Panic(message string) {
	panic(message)
}

func IfPanic(b bool, message string) {
	if b {
		panic(message)
	}
}

func Recover(funcName string) {
	if r := recover(); r != nil {
		log.Printf("%s Recovered: %v", funcName, r)
	}
}
