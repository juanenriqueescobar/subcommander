package di

import (
	"context"

	"github.com/juanenriqueescobar/subcommander/internal"
	"github.com/juanenriqueescobar/subcommander/internal/commander"
)

func readers(a []*internal.PollerSQS) []commander.Reader {
	r := make([]commander.Reader, len(a))

	for i, rr := range a {
		r[i] = rr
	}

	return r
}

func ctx() commander.Ctx {
	ctx, cancel := context.WithCancel(context.Background())
	return commander.Ctx{
		Ctx:    ctx,
		Cancel: cancel,
	}
}
