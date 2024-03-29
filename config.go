package main

type Config struct {
	Path    string `yaml:"path" validate:"required"`
	Package string `yaml:"package" validate:"required"`
}
