package wsbase

import "log"

type LogHandler func(logType int, val string)

func PrintDefault(logType int, val string) {
	log.Println(val)
}
