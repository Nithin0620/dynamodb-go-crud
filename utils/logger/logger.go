package logger

import "log"

func PANIC(message string, err error) {
	if err != nil {
		log.Panic(message, err)
	}
}

func ERROR(message string, err error) {
	if err != nil {
		log.Println(message, err)
	}
}

func INFO(message string, data interface{}) {
	log.Print(message, data)
}
