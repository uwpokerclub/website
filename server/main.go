package main

import (
	"api/cmd"
	"embed"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Println("Failed to set goose dialect.")
		os.Exit(1)
	}

	cmd.Execute()
}
