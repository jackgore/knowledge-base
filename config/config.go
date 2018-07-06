package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	DefaultAllowPublic      = false
	DefaultCookieName       = "knowledge_base"
	DefaultPublicCookieName = "kb-public"
	DefaultCookieDuration   = 3600 * 24 * 365
	DefaultDBHost           = "0.0.0.0"
	DefaultDBName           = "kbase"
	DefaultDBUser           = "kbase"
	DefaultDBPassword       = "password"
	DefaultPort             = 3001
)

type DBConfig struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
}

type Config struct {
	AllowPublicQuestions bool     `yaml:"allow-public-questions"`
	CookieName           string   `yaml:"cookie-name"`
	PublicCookieName     string   `yaml:"public-cookie-name"`
	CookieDuration       int64    `yaml:"cookie-duration"`
	Port                 int      `yaml:"port"`
	Database             DBConfig `yaml:"database"`
}

// DefaultConfig builds a Config object using all the default values.
func DefaultConfig() Config {
	return Config{
		AllowPublicQuestions: DefaultAllowPublic,
		CookieDuration:       DefaultCookieDuration,
		CookieName:           DefaultCookieName,
		Database: DBConfig{
			Name:     DefaultDBName,
			User:     DefaultDBUser,
			Password: DefaultDBPassword,
			Host:     DefaultDBHost,
		},
		Port:             DefaultPort,
		PublicCookieName: DefaultPublicCookieName,
	}
}

// New creates a new config from the config file specified in the filename.
func New(filename string) (Config, error) {
	var c Config

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}

	c = DefaultConfig()

	err = yaml.Unmarshal(contents, &c)
	if err != nil {
		return c, err
	}

	return c, nil
}
