package config

import (
	"os"
	"strings"
)

type EnvType string

const (
	EnvKey = "ENV"

	EnvTypeLocal EnvType = "local"
	EnvTypeProd  EnvType = "prod"
)

func getEnvType() EnvType {
	e := os.Getenv(EnvKey)
	switch {
	case e == strings.ToLower(string(EnvTypeLocal)):
		return EnvTypeLocal
	case e == strings.ToLower(string(EnvTypeProd)):
		return EnvTypeProd
	default:
		return EnvTypeLocal
	}
}
