package logger

import (
	"github.com/nelsw/bytelyon/internal/config"
	gormlogger "gorm.io/gorm/logger"
)

func NewGorm() gormlogger.Interface {

	var level gormlogger.LogLevel
	if config.IsReleaseMode() {
		level = gormlogger.Error
	} else if config.IsDebugMode() {
		level = gormlogger.Warn
	} else {
		level = gormlogger.Info
	}

	return gormlogger.Default.LogMode(level)
}
