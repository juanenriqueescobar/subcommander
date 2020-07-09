package internal

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/sirupsen/logrus"
)

type SQS interface {
	GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	// ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	ReceiveMessageWithContext(ctx aws.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
}

type ExecutorI interface {
	run(string, string) bool
}

type PollerSQSFilter struct {
	e ExecutorI
	f string
	v string
}

func (s *PollerSQSFilter) filter(attrs map[string]*sqs.MessageAttributeValue) bool {
	attr, ok := attrs[s.f]
	return ok && *attr.StringValue == s.v
}

type PollerSQS struct {
	client              SQS
	queueURL            *string
	waitTimeSeconds     *int64
	maxNumberOfMessages *int64
	waitBetweenRequest  time.Duration
	logger              *logrus.Entry
	filters             []*PollerSQSFilter
	sleepOnError        time.Duration
}

func (p *PollerSQS) poll(ctx context.Context) error {

	messages, err := p.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		MaxNumberOfMessages:   p.maxNumberOfMessages,
		QueueUrl:              p.queueURL,
		WaitTimeSeconds:       p.waitTimeSeconds,
		MessageAttributeNames: []*string{aws.String("All")},
	})

	if err != nil {
		return err
	}

	if len(messages.Messages) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}

	for _, m := range messages.Messages {
		for _, f := range p.filters {
			if f.filter(m.MessageAttributes) {
				wg.Add(1)
				go func(mm *sqs.Message, ff *PollerSQSFilter) {
					delete := ff.e.run(*mm.MessageId, *mm.Body)
					if delete {
						_, err := p.client.DeleteMessage(&sqs.DeleteMessageInput{
							QueueUrl:      p.queueURL,
							ReceiptHandle: mm.ReceiptHandle,
						})
						if err != nil {
							p.logger.WithError(err).Error("message cant be deleted")
						}
					}
					wg.Done()
				}(m, f)
			}
		}
	}
	wg.Wait()

	return nil
}

func (p *PollerSQS) Run(ctx context.Context) {
	p.logger.Info("start")
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("stop")
			return
		default:
			err := p.poll(ctx)
			var sleep time.Duration
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
					p.logger.Debug("request canceled by context")
					// sleep = 0
				} else {
					p.logger.WithError(err).Error("error getting messages")
					sleep = p.sleepOnError
				}
			} else {
				sleep = p.waitBetweenRequest
			}

			if sleep > 0 {
				// sleep some time or until context were canceled
				ctx2, cancel := context.WithTimeout(ctx, sleep)
				defer cancel()
				<-ctx2.Done()
			}
		}
	}
}

func max(a, b int64) *int64 {
	if a > b {
		return &a
	}
	return &b
}

func NewPollerSQS(config config.Sqs, client SQS, parentLogger *logrus.Entry) (*PollerSQS, error) {

	logger := parentLogger.WithField("_queue", config.QueueName)

	queueURL, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(config.QueueName),
	})
	if err != nil {
		return nil, err
	}

	filters := make([]*PollerSQSFilter, len(config.Commands))
	for i, c := range config.Commands {
		filters[i] = &PollerSQSFilter{
			e: NewExec(c.Command, c.Args, logger),
			f: config.AttributeName,
			v: c.AttributeValue,
		}
	}

	p := &PollerSQS{
		client:              client,
		filters:             filters,
		logger:              logger,
		queueURL:            queueURL.QueueUrl,
		waitTimeSeconds:     aws.Int64(config.WaitTimeSeconds),
		maxNumberOfMessages: max(config.MaxNumberOfMessages, 1),
		sleepOnError:        10 * time.Second,
		waitBetweenRequest:  time.Duration(config.WaitBetweenRequest) * time.Second,
	}
	return p, nil
}

func PollerSQSBuilder(c *config.Config, s SQS, l *logrus.Entry) ([]*PollerSQS, error) {
	pollers := make([]*PollerSQS, len(c.Sqs))
	for i, r := range c.Sqs {
		p, err := NewPollerSQS(r, s, l)
		if err != nil {
			return nil, err
		}
		pollers[i] = p
	}
	return pollers, nil
}
