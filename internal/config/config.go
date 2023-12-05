package config

import "gopkg.in/yaml.v3"

type SourceConfig struct {
	ArticleSelector string `yaml:"article" json:"article"`
	TitleSelector   string `yaml:"title" json:"title"`
	DetailSelector  string `yaml:"detail" json:"detail"`
	// Tags теги, которые будут отрисованы в сообщении.
	Tags []string `yaml:"tags"`
	// Channel переопределяет канал для отсылки.
	Channels []string `yaml:"channels"`
}

type Source struct {
	Name   string       `yaml:"name" json:"name"`
	URL    string       `yaml:"url" json:"url"`
	Config SourceConfig `yaml:"config" json:"config"`
}

type Config struct {
	Sources []Source `yaml:"sources" json:"sources"`
}

func Parse(data []byte) (Config, error) {
	var config Config
	return config, yaml.Unmarshal(data, &config)
}

func ParseSourceConfig(data []byte) (SourceConfig, error) {
	var config SourceConfig
	return config, yaml.Unmarshal(data, &config)
}
