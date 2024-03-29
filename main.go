package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

//go:embed migration.template
var migrationTemplate string

//go:embed config.yaml
var defaultConfig []byte

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run pg-migration Add column to table")
		os.Exit(0)
	}

	data, err := os.ReadFile("./migration.yaml")

	if err != nil {
		data = defaultConfig
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Invalid config: %s\n", err)
		os.Exit(1)
	}

	validate := validator.New()
	err = validate.Struct(&config)
	if err != nil {
		fmt.Printf("Invalid config: %s\n", err)
		os.Exit(1)
	}

	t, _ := template.New("migration").Parse(migrationTemplate)

	args := os.Args[1:]
	now := time.Now()
	time := strings.ReplaceAll(now.Format("20060102150405.000"), ".", "")
	comments := strings.ToLower(strings.Join(args, "_"))
	name := fmt.Sprintf("%s_%s", time, comments)
	fileName := fmt.Sprintf("%s.go", name)
	fmt.Printf("Creating new migration file: %s/%s\n", config.Path, fileName)
	err = os.MkdirAll(config.Path, os.ModePerm)
	if err != nil {
		fmt.Printf("Creating folder '%s' failed: %s\n", config.Path, err.Error())
		os.Exit(1)
	}

	f, err := os.Create(filepath.Join(config.Path, fileName))
	if err != nil {
		fmt.Printf("Creating file '%s' failed: %s\n", fileName, err.Error())
		os.Exit(1)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("Close file failed: %s\n", err)
		}
	}(f)

	err = t.Execute(f, map[string]string{
		"Package": config.Package,
		"Version": time,
	})
	if err != nil {
		fmt.Printf("Execute template failed: %s\n", err)
	}
}
