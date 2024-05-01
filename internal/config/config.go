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
	Parallel  int    `yaml:"parallel"`
}

type IndexConfig struct {
	IndexFile string `yaml:"index_file"`
}

type Config struct {
	DBCfg    DBConfig    `yaml:"database"`
	AppCFG   AppConfig   `yaml:"app"`
	IndexCfg IndexConfig `yaml:"index"`
}

func NewConfig(c string) (Config, error) {
	if c == "" {
		c = "config.yaml"
	}
	file, err := os.Open(c)
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
