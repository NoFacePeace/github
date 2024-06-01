package log

import "log"

func Init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Config struct {
	Mode string
}
