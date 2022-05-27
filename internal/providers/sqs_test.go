package providers

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/juanenriqueescobar/subcommander/internal/pollers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

type SqsMock struct {
	mock.Mock
}

func (m *SqsMock) GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.GetQueueAttributesOutput), args.Error(1)
}

func (m *SqsMock) GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.GetQueueUrlOutput), args.Error(1)
}

type SQSConstructorMock struct {
	mock.Mock
}

func (m *SQSConstructorMock) New(pc pollers.SqsConfig, commands []config.SqsCommands, logger *logrus.Entry) *pollers.SqsPoller {
	return m.Called(pc, commands, logger).Get(0).(*pollers.SqsPoller)
}

func TestSqsPollerProvider(t *testing.T) {
	type args struct {
		c       *config.Config
		client  *SqsMock
		builder *SQSConstructorMock
	}
	tests := []struct {
		name    string
		args    args
		mocker  func(*SqsMock, *SQSConstructorMock, *logrus.Entry)
		want    []*pollers.SqsPoller
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				c: &config.Config{
					Sqs: []config.Sqs{
						{
							QueueName:           "queue_1",
							AttributeName:       "name_1",
							WaitTimeSeconds:     10,
							MaxNumberOfMessages: 10,
							WaitBetweenRequest:  10,
							ThreadsNumber:       3,
							Commands: []config.SqsCommands{
								{
									AttributeValue: "value_1.1",
									Command:        "cmd_1.1",
									Args: []string{
										"a11",
										"b11",
									},
								},
								{
									AttributeValue: "value_1.2",
									Command:        "cmd_1.2",
									Args: []string{
										"a12",
										"b12",
									},
									Timeout: time.Hour,
								},
							},
						},
						{
							QueueName:     "queue_2",
							AttributeName: "name_2",
							Commands: []config.SqsCommands{
								{
									AttributeValue: "value_2",
									Command:        "cmd_2",
								},
							},
						},
						{
							QueueName:     "queue_3",
							AttributeName: "name_3",
							Commands: []config.SqsCommands{
								{
									AttributeValue: "value_3",
									Command:        "cmd_3",
								},
							},
						},
					},
				},
				client:  &SqsMock{},
				builder: &SQSConstructorMock{},
			},
			mocker: func(s *SqsMock, b *SQSConstructorMock, l *logrus.Entry) {

				// 1
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_1"),
					},
					nil,
				).Once()
				s.On("GetQueueAttributes", &sqs.GetQueueAttributesInput{
					AttributeNames: []*string{aws.String("VisibilityTimeout")},
					QueueUrl:       aws.String("https://queue_1"),
				}).Return(
					&sqs.GetQueueAttributesOutput{
						Attributes: map[string]*string{
							"VisibilityTimeout": aws.String("180"),
						},
					},
					nil,
				).Once()
				b.On(
					"New",
					pollers.SqsConfig{
						AttributeName:       "name_1",
						MaxNumberOfMessages: 10,
						QueueURL:            aws.String("https://queue_1"),
						VisibilityTimeout:   180,
						WaitTimeSeconds:     10,
						WaitBetweenRequest:  10,
					},
					[]config.SqsCommands{
						{
							AttributeValue: "value_1.1",
							Command:        "cmd_1.1",
							Args: []string{
								"a11",
								"b11",
							},
						},
						{
							AttributeValue: "value_1.2",
							Command:        "cmd_1.2",
							Args: []string{
								"a12",
								"b12",
							},
							Timeout: time.Hour,
						},
					},
					l.WithFields(logrus.Fields{
						"sc_task.queue":      "queue_1",
						"Subcommander.queue": "queue_1",
					}),
				).Return(&pollers.SqsPoller{}).Times(3)

				// 2
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_2"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_2"),
					},
					nil,
				).Once()
				s.On("GetQueueAttributes", &sqs.GetQueueAttributesInput{
					AttributeNames: []*string{aws.String("VisibilityTimeout")},
					QueueUrl:       aws.String("https://queue_2"),
				}).Return(
					&sqs.GetQueueAttributesOutput{
						Attributes: map[string]*string{
							"VisibilityTimeout": aws.String(""),
						},
					},
					nil,
				).Once()
				b.On(
					"New",
					pollers.SqsConfig{
						AttributeName:     "name_2",
						QueueURL:          aws.String("https://queue_2"),
						VisibilityTimeout: 30,
					},
					[]config.SqsCommands{
						{
							AttributeValue: "value_2",
							Command:        "cmd_2",
						},
					},
					l.WithFields(logrus.Fields{
						"sc_task.queue":      "queue_2",
						"Subcommander.queue": "queue_2",
					}),
				).Return(&pollers.SqsPoller{}).Once()

				// 3
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_3"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_3"),
					},
					nil,
				).Once()
				s.On("GetQueueAttributes", &sqs.GetQueueAttributesInput{
					AttributeNames: []*string{aws.String("VisibilityTimeout")},
					QueueUrl:       aws.String("https://queue_3"),
				}).Return(
					&sqs.GetQueueAttributesOutput{
						Attributes: map[string]*string{},
					},
					nil,
				).Once()
				b.On(
					"New",
					pollers.SqsConfig{
						AttributeName:     "name_3",
						QueueURL:          aws.String("https://queue_3"),
						VisibilityTimeout: 30,
					},
					[]config.SqsCommands{
						{
							AttributeValue: "value_3",
							Command:        "cmd_3",
						},
					},
					l.WithFields(logrus.Fields{
						"sc_task.queue":      "queue_3",
						"Subcommander.queue": "queue_3",
					}),
				).Return(&pollers.SqsPoller{}).Once()
			},
			want: []*pollers.SqsPoller{
				{},
				{},
				{},
				{},
				{},
			},
			wantErr: false,
		},
		{
			name: "err: get queue url",
			args: args{
				c: &config.Config{
					Sqs: []config.Sqs{
						{
							QueueName:           "queue_1",
							WaitTimeSeconds:     10,
							MaxNumberOfMessages: 10,
							WaitBetweenRequest:  10,
							AttributeName:       "name",
							ThreadsNumber:       3,
							Commands: []config.SqsCommands{
								{
									AttributeValue: "value_1",
									Command:        "cmd_1",
									Args: []string{
										"a1",
										"b1",
									},
								},
								{
									AttributeValue: "value_2",
									Command:        "cmd_2",
									Args: []string{
										"a2",
										"b2",
									},
								},
							},
						},
					},
				},
				client:  &SqsMock{},
				builder: &SQSConstructorMock{},
			},
			mocker: func(s *SqsMock, b *SQSConstructorMock, l *logrus.Entry) {
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{},
					errors.New("ups"),
				).Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "err: get queue attributes",
			args: args{
				c: &config.Config{
					Sqs: []config.Sqs{
						{
							QueueName:           "queue_1",
							WaitTimeSeconds:     10,
							MaxNumberOfMessages: 10,
							WaitBetweenRequest:  10,
							AttributeName:       "name",
							ThreadsNumber:       3,
							Commands: []config.SqsCommands{
								{
									AttributeValue: "value_1",
									Command:        "cmd_1",
									Args: []string{
										"a1",
										"b1",
									},
								},
								{
									AttributeValue: "value_2",
									Command:        "cmd_2",
									Args: []string{
										"a2",
										"b2",
									},
								},
							},
						},
					},
				},
				client:  &SqsMock{},
				builder: &SQSConstructorMock{},
			},
			mocker: func(s *SqsMock, b *SQSConstructorMock, l *logrus.Entry) {
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_1"),
					},
					nil,
				).Once()

				s.On("GetQueueAttributes", &sqs.GetQueueAttributesInput{
					AttributeNames: []*string{aws.String("VisibilityTimeout")},
					QueueUrl:       aws.String("https://queue_1"),
				}).Return(
					&sqs.GetQueueAttributesOutput{},
					errors.New("ups"),
				).Once()
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.NewEntry(logrus.New())

			tt.mocker(tt.args.client, tt.args.builder, logger)

			got, err := SqsPollerProvider(tt.args.c, tt.args.client, tt.args.builder, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Error("ups, pollers are not equals")
				fmt.Printf("%#v\n", got)
			}

			tt.args.client.AssertExpectations(t)
			tt.args.builder.AssertExpectations(t)
		})
	}
}
