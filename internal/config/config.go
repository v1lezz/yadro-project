package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type JsonDBConfig struct {
	DBFile string `yaml:"db_file"`
}

type PostgresDBConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
}

func (cfg PostgresDBConfig) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

type AppConfig struct {
	SourceURL string `yaml:"source_url"`
	Parallel  int    `yaml:"parallel"`
}

type IndexConfig struct {
	IndexFile string `yaml:"index_file"`
}

type ServerConfig struct {
	Port             int `yaml:"port"`
	ConcurrencyLimit int `yaml:"concurrency_limit"`
	RateLimit        int `yaml:"rate_limit"`
}

type AuthConfig struct {
	TokenMaxTime time.Duration `yaml:"token_max_time"`
}

type Config struct {
	DbCFG    PostgresDBConfig `yaml:"database"`
	AppCFG   AppConfig        `yaml:"app"`
	IndexCFG IndexConfig      `yaml:"index"`
	SrvCFG   ServerConfig     `yaml:"server"`
	AuthCFG  AuthConfig       `yaml:"auth"`
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
	c.IndexCFG.SetDefault()
	c.SrvCFG.SetDefault()
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
	if c.ConcurrencyLimit == 0 {
		c.ConcurrencyLimit = 10
	}
	if c.RateLimit == 0 {
		c.RateLimit = 0
	}
}

func (c *AuthConfig) SetDefault() {
	if c.TokenMaxTime == 0 {
		c.TokenMaxTime = time.Second * 10
	}
}
