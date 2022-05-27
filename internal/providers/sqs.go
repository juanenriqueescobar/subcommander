package providers

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/juanenriqueescobar/subcommander/internal/pollers"
	"github.com/sirupsen/logrus"
)

type Sqs interface {
	GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error)
	GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
}

type SqsConstructor interface {
	New(pc pollers.SqsConfig, commands []config.SqsCommands, logger *logrus.Entry) *pollers.SqsPoller
}

func SqsPollerProvider(cnf *config.Config, client Sqs, con SqsConstructor, logger *logrus.Entry) ([]*pollers.SqsPoller, error) {
	ps := make([]*pollers.SqsPoller, 0)
	for _, c := range cnf.Sqs {
		queueURL, err := client.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(c.QueueName),
		})
		if err != nil {
			return nil, err
		}
		attributes, err := client.GetQueueAttributes(&sqs.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String("VisibilityTimeout")},
			QueueUrl:       queueURL.QueueUrl,
		})
		if err != nil {
			return nil, err
		}

		vt1, ok := attributes.Attributes["VisibilityTimeout"]
		if !ok || len(*vt1) == 0 {
			vt1 = aws.String("30")
		}
		// ignoramos el error pues aws siempre va a devolver un numero
		vt2, _ := strconv.Atoi(*vt1)

		pc := pollers.SqsConfig{
			AttributeName:       c.AttributeName,
			MaxNumberOfMessages: c.MaxNumberOfMessages,
			QueueURL:            queueURL.QueueUrl,
			VisibilityTimeout:   vt2,
			WaitTimeSeconds:     c.WaitTimeSeconds,
			WaitBetweenRequest:  c.WaitBetweenRequest,
		}
		logger2 := logger.WithFields(logrus.Fields{
			"sc_task.queue":      c.QueueName, // TODO deprecated
			"Subcommander.queue": c.QueueName,
		})

		x := *pollers.Max(1, c.ThreadsNumber)
		for i := int64(0); i < x; i++ {
			ps = append(ps, con.New(pc, c.Commands, logger2))
		}
	}
	return ps, nil
}
