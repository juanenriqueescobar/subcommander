package commander

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

type Ctx struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

type Reader interface {
	Run(context.Context)
}

type Commander struct {
	signals chan os.Signal
	wg      *sync.WaitGroup
	logger  *logrus.Entry
	ctx     context.Context
	cancel  context.CancelFunc
	counter int
}

func (c *Commander) register(reader Reader) {
	c.counter++
	c.wg.Add(1)
	go func(rr Reader, cc *Commander) {
		rr.Run(c.ctx)
		cc.wg.Done()
	}(reader, c)
}

func (c *Commander) Wait() {
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c.signals
		c.logger.WithField("signal", sig).Info("stop signal received")
		c.cancel()
	}()

	c.logger.WithField("readers", c.counter).Info("start ok")
	c.wg.Wait()
	c.logger.Info("shutdown ok")
}

func NewCommander(p []Reader, ctx Ctx, logger *logrus.Entry) *Commander {
	c := &Commander{
		signals: make(chan os.Signal, 1),
		wg:      &sync.WaitGroup{},
		logger:  logger,
		ctx:     ctx.Ctx,
		cancel:  ctx.Cancel,
	}

	for _, r := range p {
		c.register(r)
	}

	return c
}
