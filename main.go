package main

import (
	"os"
	"time"

	"bottles/brew"

	"github.com/charmbracelet/log"
)

func main() {
	var logger = log.NewWithOptions(os.Stderr, log.Options{
		Level:      log.DebugLevel,
		TimeFormat: time.TimeOnly,
	})

	b := brew.New(logger)

	for _, pkg := range b.GetPackages(os.Args[1:]...) {
		logger.Debug(pkg.Name, "version", pkg.Version, "prebottled", pkg.Bottles.ContainsAny("all", "catalina"))
		pkg.Bottle()
	}
}
