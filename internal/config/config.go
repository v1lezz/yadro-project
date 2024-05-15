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

type ServerConfig struct {
	Port int `yaml:"port"`
}

type Config struct {
	DbCFG    DBConfig     `yaml:"database"`
	AppCFG   AppConfig    `yaml:"app"`
	IndexCFG IndexConfig  `yaml:"index"`
	SrvCFG   ServerConfig `yaml:"server"`
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
	c.DbCFG.SetDefault()
	c.IndexCFG.SetDefault()
	c.SrvCFG.SetDefault()
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

func (c *IndexConfig) SetDefault() {
	if c.IndexFile == "" {
		c.IndexFile = "index.json"
	}
}

func (c *ServerConfig) SetDefault() {
	if c.Port == 0 {
		c.Port = 9000
	}
}
