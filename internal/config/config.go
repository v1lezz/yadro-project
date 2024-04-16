package config

import (
	"flag"
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

type Config struct {
	DBCfg  DBConfig  `yaml:"database"`
	AppCFG AppConfig `yaml:"app"`
}

func NewConfig() (Config, error) {
	file, err := os.Open(ParsePathConfigFile())
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ParsePathConfigFile() string {
	var c string
	flag.StringVar(&c, "c", "", "parse file path config")
	if c == "" {
		c = "config.yaml"
	}
	return c
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
