package internal

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/sirupsen/logrus"
)

func TestPollerSQS_PollerSQSBuilder(t *testing.T) {
	type args struct {
		c      *config.Config
		client *SQSMock
	}
	tests := []struct {
		name    string
		args    args
		want    func(*SQSMock, *logrus.Entry) []*PollerSQS
		wantLen int
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				c: &config.Config{
					Sqs: []config.Sqs{
						{
							QueueName:           "queue_1",
							WaitTimeSeconds:     10,
							MaxNumberOfMessages: 10,
							WaitBetweenRequest:  10,
							AttributeName:       "name",
							ThreadsNumber:       10,
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
				client: &SQSMock{},
			},
			want: func(s *SQSMock, l *logrus.Entry) []*PollerSQS {

				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_1"),
					},
					nil,
				).Times(10)

				l = l.WithField("sc_task.queue", "queue_1")
				x := make([]*PollerSQS, 10)
				for i := 0; i < 10; i++ {
					x[i] = &PollerSQS{
						client:              s,
						logger:              l,
						maxNumberOfMessages: aws.Int64(10),
						queueURL:            aws.String("https://queue_1"),
						waitTimeSeconds:     aws.Int64(10),
						sleepOnError:        time.Duration(10 * time.Second),
						waitBetweenRequest:  time.Duration(10 * time.Second),
						filters: []*PollerSQSFilter{
							{
								e: NewExec("cmd_1", []string{"a1", "b1"}, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_1"})),
								f: "name",
								v: "value_1",
							},
							{
								e: NewExec("cmd_2", []string{"a2", "b2"}, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_2"})),
								f: "name",
								v: "value_2",
							},
						},
					}
				}
				return x
			},
			wantLen: 10,
			wantErr: false,
		},
		{
			name: "2",
			args: args{
				c: &config.Config{
					Sqs: []config.Sqs{
						{
							QueueName:           "queue_1",
							WaitTimeSeconds:     10,
							MaxNumberOfMessages: 10,
							WaitBetweenRequest:  10,
							AttributeName:       "name",
							ThreadsNumber:       0,
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
				client: &SQSMock{},
			},
			want: func(s *SQSMock, l *logrus.Entry) []*PollerSQS {

				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_1"),
					},
					nil,
				).Once()

				l = l.WithField("sc_task.queue", "queue_1")
				x := make([]*PollerSQS, 1)
				for i := 0; i < 1; i++ {
					x[i] = &PollerSQS{
						client:              s,
						logger:              l,
						maxNumberOfMessages: aws.Int64(10),
						queueURL:            aws.String("https://queue_1"),
						waitTimeSeconds:     aws.Int64(10),
						sleepOnError:        time.Duration(10 * time.Second),
						waitBetweenRequest:  time.Duration(10 * time.Second),
						filters: []*PollerSQSFilter{
							{
								e: NewExec("cmd_1", []string{"a1", "b1"}, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_1"})),
								f: "name",
								v: "value_1",
							},
							{
								e: NewExec("cmd_2", []string{"a2", "b2"}, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_2"})),
								f: "name",
								v: "value_2",
							},
						},
					}
				}
				return x
			},
			wantLen: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New().WithField("", "")

			want := tt.want(tt.args.client, logger)
			got, err := PollerSQSBuilder(tt.args.c, tt.args.client, logger)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Error("ups", len(got))
			}

			if !reflect.DeepEqual(want, got) {
				t.Error("ups, pollers are not equals")
				fmt.Printf("%#v\n", want)
				fmt.Printf("%#v\n", got)
			}

			tt.args.client.AssertExpectations(t)
		})
	}
}
