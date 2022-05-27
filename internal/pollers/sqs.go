package pollers

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/juanenriqueescobar/subcommander/internal/executor"
	"github.com/sirupsen/logrus"
)

type Sqs interface {
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	ReceiveMessageWithContext(ctx aws.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error)
}

type Executor interface {
	Run(string, executor.StdinData) (bool, error)
}

type SqsFilter struct {
	e Executor
	f string
	v string
}

func (s *SqsFilter) filter(attrs map[string]*sqs.MessageAttributeValue) bool {
	attr, ok := attrs[s.f]
	return ok && *attr.StringValue == s.v
}

type SqsPoller struct {
	client              Sqs
	queueURL            *string
	waitTimeSeconds     *int64
	maxNumberOfMessages *int64
	waitBetweenRequest  time.Duration
	logger              *logrus.Entry
	filters             []*SqsFilter
	sleepOnError        time.Duration
}

// poll
// si retorna true entonces espera antes de hacer la siguiente consulta
func (p *SqsPoller) poll(ctx context.Context) (bool, error) {

	messages, err := p.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		MaxNumberOfMessages:   p.maxNumberOfMessages,
		QueueUrl:              p.queueURL,
		WaitTimeSeconds:       p.waitTimeSeconds,
		MessageAttributeNames: []*string{aws.String("All")},
	})

	if err != nil {
		return true, err
	}

	if len(messages.Messages) == 0 {
		return true, nil
	}

	wg := sync.WaitGroup{}

	for _, m := range messages.Messages {
		for _, f := range p.filters {
			if f.filter(m.MessageAttributes) {
				wg.Add(1)
				go func(mm *sqs.Message, ff *SqsFilter) {
					defer wg.Done()
					data := executor.StdinData{
						Metadata: map[string]interface{}{},
						Body:     *mm.Body,
					}
					delete, err := ff.e.Run(*mm.MessageId, data)
					if err != nil {
						// TODO do something!!
					} else {
						if delete {
							_, err := p.client.DeleteMessage(&sqs.DeleteMessageInput{
								QueueUrl:      p.queueURL,
								ReceiptHandle: mm.ReceiptHandle,
							})
							if err != nil {
								p.logger.WithError(err).Error("message cant be deleted")
							}
						}
					}
				}(m, f)
			}
		}
	}
	wg.Wait()

	return len(messages.Messages) == 0, nil
}

func (p *SqsPoller) Run(ctx context.Context) {
	p.logger.Info("start")
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("stop")
			return
		default:
			wait, err := p.poll(ctx)
			var sleep time.Duration
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
					p.logger.Debug("request canceled by context")
				} else {
					p.logger.WithError(err).Error("error getting messages")
					sleep = p.sleepOnError
				}
			} else {
				sleep = p.waitBetweenRequest
			}

			if wait {
				// sleep some time or until context were canceled
				ctx2, cancel := context.WithTimeout(ctx, sleep)
				defer cancel()
				<-ctx2.Done()
			}
		}
	}
}

type SqsConfig struct {
	AttributeName       string
	MaxNumberOfMessages int64
	QueueURL            *string
	VisibilityTimeout   int
	WaitTimeSeconds     int64
	WaitBetweenRequest  int64
}

type SqsPollerConstructor struct {
	Client Sqs
}

func (b *SqsPollerConstructor) New(cnf SqsConfig, commands []config.SqsCommands, logger *logrus.Entry) *SqsPoller {
	filters := make([]*SqsFilter, len(commands))
	for i, c := range commands {

		vt := time.Duration(cnf.VisibilityTimeout) * time.Second
		if c.Timeout == 0 || c.Timeout > vt {
			c.Timeout = vt
		}

		filters[i] = &SqsFilter{
			e: executor.NewExec(c.Command, c.Args, c.Timeout, logger.WithFields(logrus.Fields{
				"sc_task.attr_name":  cnf.AttributeName,
				"sc_task.attr_value": c.AttributeValue,
			})),
			f: cnf.AttributeName,
			v: c.AttributeValue,
		}
	}

	p := &SqsPoller{
		client:              b.Client,
		filters:             filters,
		logger:              logger,
		queueURL:            cnf.QueueURL,
		waitTimeSeconds:     &cnf.WaitTimeSeconds,
		maxNumberOfMessages: Max(cnf.MaxNumberOfMessages, 1),
		sleepOnError:        10 * time.Second,
		waitBetweenRequest:  time.Duration(cnf.WaitBetweenRequest) * time.Second,
	}
	return p
}
