package cfg

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Laps        int    `json:"laps" env-default:"2"`
	LapLen      int    `json:"lapLen" env-default:"3500"`
	PenaltyLen  int    `json:"penaltyLen" env-default:"150"`
	FiringLines int    `json:"firingLines" env-default:"2"`
	Start       string `json:"start" env-default:"10:00:00.000"`
	StartDelta  string `json:"startDelta" env-default:"00:01:30"`
}

func MustLoad() *Config {
	const cfg_path = "../internal/cfg/config.json"
	if _, err := os.Stat(cfg_path); err != nil {
		log.Fatalf("cfg not found in %s", cfg_path)
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(cfg_path, cfg); err != nil {
		log.Fatalf("failed to read config %s", cfg_path)
	}

	return cfg
}
