package cmd

import (
	"fmt"

	"ep.k16/newsfeed/config"
	"ep.k16/newsfeed/pkg/logger"
)

// InitLogger load log config and init logger
func InitLogger() error {
	// init logger
	logCfg, err := config.LoadLogConfig()
	if err != nil {
		logger.Error("failed to load logger config", logger.E(err))
		return fmt.Errorf("load log config: %s", err)
	}

	// set logger
	cfg := logger.Config{
		Type:     logCfg.Type,
		Level:    logCfg.Level,
		Output:   logCfg.Output,
		Filename: logCfg.Filename,
		UseJSON:  logCfg.UseJSON,
	}
	logger.SetLogger(cfg)

	logger.Info("init logger successfully", logger.F("cfg", cfg))
	return nil
}
