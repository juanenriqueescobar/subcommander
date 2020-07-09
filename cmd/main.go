package main

import (
	"os"

	"github.com/juanenriqueescobar/subcommander/internal/di"
)

func main() {
	l := di.Logger()
	c, err := di.Commander(l, os.Args[1:])
	if err != nil {
		l.WithError(err).Fatal("unable to start")
	}
	c.Wait()
}
