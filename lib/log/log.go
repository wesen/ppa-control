package logger

import "github.com/rs/zerolog/log"

func InitializeLogger(withCaller bool) {
	if withCaller {
		log.Logger = log.With().Caller().Logger()
	}
}
