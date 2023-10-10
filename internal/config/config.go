package config

import "gopkg.in/yaml.v3"

type Detail struct {
	Selector string `yaml:"selector" json:"selector"`
}

type Title struct {
	Selector string `yaml:"selector" json:"selector"`
}

type Next struct {
	Selector string `yaml:"selector" json:"selector"`
}

type Article struct {
	Selector string `yaml:"selector" json:"selector"`
	Title    Title  `yaml:"title" json:"title"`
	Detail   Detail `yaml:"detail" json:"detail"`
	Next     Next   `yaml:"next" json:"next"`
}

type SourceConfig struct {
	ArticleCard Article `yaml:"article" json:"article"`
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
