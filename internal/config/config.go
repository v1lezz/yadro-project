package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	DBFile string `yaml:"db_file"`
}

type AppConfig struct {
	SourceURL string `yaml:"source_url"`
}

type Config struct {
	DBCfg  DBConfig  `yaml:"database"`
	AppCFG AppConfig `yaml:"app"`
}

func NewConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) SetDefault() {
	c.AppCFG.SetDefault()
	c.DBCfg.SetDefault()
}

func (c *DBConfig) SetDefault() {
	if c.DBFile == "" {
		c.DBFile = "database.json"
	}
}

func (c *AppConfig) SetDefault() {
	if c.SourceURL == "" {
		c.SourceURL = "xkcd.com"
	}
}
