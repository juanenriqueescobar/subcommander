package commander

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

type readerMock struct {
	mock.Mock
}

func (t *readerMock) Run(ctx context.Context) {
	t.Called(ctx)
	<-ctx.Done()
}

func convert(a []*readerMock, ctx context.Context) []Reader {
	r := make([]Reader, len(a))
	for i, rr := range a {
		rr.On("Run", ctx).Once()
		r[i] = rr
	}
	return r
}

func TestCommander_Wait(t *testing.T) {
	type fields struct {
		sleep   time.Duration
		readers []*readerMock
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "1",
			fields: fields{
				sleep: 10 * time.Millisecond,
				readers: []*readerMock{
					&readerMock{},
				},
			},
		},
		{
			name: "2",
			fields: fields{
				sleep: 20 * time.Millisecond,
				readers: []*readerMock{
					&readerMock{},
					&readerMock{},
				},
			},
		},
		{
			name: "3",
			fields: fields{
				sleep: 30 * time.Millisecond,
				readers: []*readerMock{
					&readerMock{},
					&readerMock{},
					&readerMock{},
					&readerMock{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, hook := test.NewNullLogger()
			ctx, cancel := context.WithCancel(context.Background())
			r := convert(tt.fields.readers, ctx)
			c := NewCommander(r, Ctx{Ctx: ctx, Cancel: cancel}, logrus.NewEntry(logger))
			go func() {
				time.Sleep(tt.fields.sleep)
				c.signals <- syscall.SIGINT
			}()
			c.Wait()

			time.Sleep(2 * time.Millisecond)

			if c.counter != len(tt.fields.readers) {
				t.Error("fail reader counter")
			}
			if len(hook.Entries) != 3 {
				t.Error("fail logger entries count", len(hook.Entries))
			}
			for _, m := range tt.fields.readers {
				m.AssertExpectations(t)
			}
		})
	}
}
