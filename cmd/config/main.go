package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/denisdubovitskiy/feedparser/internal/config"
	"github.com/denisdubovitskiy/feedparser/internal/database"
)

var (
	configFilename string
	databasePath   string
)

func init() {
	flag.StringVar(&configFilename, "config", "", "config file path")
	flag.StringVar(&databasePath, "database", "", "database file path")
	flag.Parse()
}

func main() {
	configContent, err := os.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("config: unable to open config file %source: %v", configFilename, err)
	}

	conf, err := config.Parse(configContent)
	if err != nil {
		log.Fatalf("config: unable to parse config file %source: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(databasePath), os.ModePerm); err != nil {
		log.Fatalf("config: unable to create a directory for a database: %v", err)
	}

	db, err := database.Open(databasePath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatalln(err)
	}

	service := database.NewService(db)

	for _, source := range conf.Sources {
		confBytes, err := json.Marshal(source.Config)
		if err != nil {
			slog.Error(fmt.Sprintf("config: skipping source %s due to marshalling error", err))
			continue
		}

		if err := service.UpsertSource(context.Background(), source.Name, source.URL, string(confBytes)); err != nil {
			slog.Error(fmt.Sprintf(`config: source "%s" update error: %v`, source.Name, err))
			continue
		}

		slog.Info(fmt.Sprintf(`config: source "%s" updated`, source.Name))
	}
}
