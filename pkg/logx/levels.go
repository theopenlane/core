package logx

import (
	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
)

var (
	echoLevels = map[log.Lvl]zerolog.Level{
		log.DEBUG: zerolog.DebugLevel,
		log.INFO:  zerolog.InfoLevel,
		log.WARN:  zerolog.WarnLevel,
		log.ERROR: zerolog.ErrorLevel,
		log.OFF:   zerolog.NoLevel,
	}

	zeroLevels = map[zerolog.Level]log.Lvl{
		zerolog.TraceLevel: log.DEBUG,
		zerolog.DebugLevel: log.DEBUG,
		zerolog.InfoLevel:  log.INFO,
		zerolog.WarnLevel:  log.WARN,
		zerolog.ErrorLevel: log.ERROR,
		zerolog.FatalLevel: log.ERROR,
		zerolog.PanicLevel: log.ERROR,
		zerolog.NoLevel:    log.OFF,
		zerolog.Disabled:   log.OFF,
	}
)

// MatchEchoLevel returns a zerolog level and echo level for a given echo level
func MatchEchoLevel(level log.Lvl) (zerolog.Level, log.Lvl) {
	zlvl, found := echoLevels[level]

	if found {
		return zlvl, level
	}

	return zerolog.NoLevel, log.OFF
}

// MatchZeroLevel returns an echo level and zerolog level for a given zerolog level
func MatchZeroLevel(level zerolog.Level) (log.Lvl, zerolog.Level) {
	elvl, found := zeroLevels[level]

	if found {
		return elvl, level
	}

	return log.OFF, zerolog.NoLevel
}
