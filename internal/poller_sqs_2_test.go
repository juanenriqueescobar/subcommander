package internal

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// func Test_poller(t *testing.T) {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region:      aws.String("us-east-1"),
// 		Credentials: credentials.NewStaticCredentials("id", "secret", "token"),
// 		Endpoint:    aws.String("http://localhost:9324"),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	client := sqs.New(sess)
// 	go func() {
// 		queueURL, _ := client.GetQueueUrl(&sqs.GetQueueUrlInput{
// 			QueueName: aws.String("producto-api-order-operations"),
// 		})
// 		time.Sleep(500 * time.Millisecond)
// 		client.SendMessage(&sqs.SendMessageInput{
// 			MessageAttributes: map[string]*sqs.MessageAttributeValue{
// 				"operation": {
// 					StringValue: aws.String("change-state"),
// 					DataType:    aws.String("String"),
// 				},
// 			},
// 			MessageBody: aws.String(`{"msg":"body-message"}`),
// 			QueueUrl:    queueURL.QueueUrl,
// 		})
// 	}()

// 	type args struct {
// 		c ConfigSqs
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			name: "1",
// 			args: args{
// 				c: ConfigSqs{
// 					QueueName:     "producto-api-order-operations",
// 					AttributeName: "operation",
// 					Commands: []ConfigSqsCommands{
// 						{
// 							AttributeValue: "change-state",
// 							Command:        "php",
// 							Args: []string{
// 								"testing/worker_1.php",
// 							},
// 						},
// 						{
// 							AttributeValue: "create-order-mu",
// 							Command:        "php",
// 							Args: []string{
// 								"testing/worker_2.php",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			poller(tt.args.c, client)
// 		})
// 	}

// 	t.Error(1)
// }

// func TestSQSPoller_run(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		mocker func(Stopper) *SQSMock
// 	}{len(c.Sqs)
// 		{
// 			"1",
// 			func(s Stopper) *SQSMock {
// 				counter := 0
// 				mocksqs := &SQSMock{}
// 				mocksqs.On("ReceiveMessage", &sqs.ReceiveMessageInput{
// 					MaxNumberOfMessages:   aws.Int64(1),
// 					MessageAttributeNames: []*string{aws.String("All")},
// 					QueueUrl:              aws.String("https://abcdefghijk"),
// 					WaitTimeSeconds:       aws.Int64(0),
// 				}).Return(
// 					&sqs.ReceiveMessageOutput{
// 						Messages: []*sqs.Message{},
// 					},
// 					nil,
// 				).Times(10).Run(func(args mock.Arguments) {
// 					counter++
// 					if counter == 10 {
// 						close(s)
// 					}
// 				}).After(10 * time.Millisecond)
// 				return mocksqs
// 			},
// 		},
// 		{
// 			"2",
// 			func(s Stopper) *SQSMock {
// 				counter := 0
// 				mocksqs := &SQSMock{}
// 				mocksqs.On("ReceiveMessage", &sqs.ReceiveMessageInput{
// 					MaxNumberOfMessages:   aws.Int64(1),
// 					MessageAttributeNames: []*string{aws.String("All")},
// 					QueueUrl:              aws.String("https://abcdefghijk"),
// 					WaitTimeSeconds:       aws.Int64(0),
// 				}).Return(
// 					&sqs.ReceiveMessageOutput{
// 						Messages: []*sqs.Message{},
// 					},
// 					nil,
// 				).Times(100).Run(func(args mock.Arguments) {
// 					counter++
// 					if counter == 100 {
// 						close(s)
// 					}
// 				}).After(2 * time.Millisecond)
// 				return mocksqs
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			stopper := make(Stopper)
// 			mock := tt.mocker(stopper)

// 			p := &SQSPoller{
// 				client:   mock,
// 				queueURL: aws.String("https://abcdefghijk"),
// 				waitTime: 0,
// 				logger:   logrus.New().WithField("", ""),
// 				filters:  []*SQSPollerFilter{},
// 			}
// 			p.Run(stopper)

// 			mock.AssertExpectations(t)
// 		})
// 	}
// }

func Test_poller_unit(t *testing.T) {
	type args struct {
		c config.Sqs
	}
	tests := []struct {
		name   string
		args   args
		mocker func() *SQSMock
		want   func(*SQSMock, *logrus.Entry) *PollerSQS
	}{
		{
			name: "1",
			args: args{
				c: config.Sqs{
					QueueName:           "queue_1",
					WaitTimeSeconds:     10,
					MaxNumberOfMessages: 10,
					AttributeName:       "operation",
					Commands: []config.SqsCommands{
						{
							AttributeValue: "task_1",
							Command:        "php",
							Args: []string{
								"testing/worker_1.php",
							},
						},
						{
							AttributeValue: "task_2",
							Command:        "php",
							Args: []string{
								"testing/worker_2.php",
							},
						},
					},
				},
			},
			mocker: func() *SQSMock {
				mock := &SQSMock{}

				mock.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://abcdefghijk"),
					},
					nil,
				).Once()

				mock.On("ReceiveMessageWithContext", context.TODO(), &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(10),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(10),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"operation": {
										DataType:    aws.String("String"),
										StringValue: aws.String("task_1"),
									},
								},
								Body:          aws.String(`{"order_id":1}`),
								MessageId:     aws.String("1"),
								ReceiptHandle: aws.String("1"),
							},
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"operation": {
										DataType:    aws.String("String"),
										StringValue: aws.String("task_1"),
									},
								},
								Body:          aws.String(`{"order_id":2}`),
								MessageId:     aws.String("2"),
								ReceiptHandle: aws.String("2"),
							},
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"operation": {
										DataType:    aws.String("String"),
										StringValue: aws.String("task_2"),
									},
								},
								Body:          aws.String(`{"order_id":3}`),
								MessageId:     aws.String("3"),
								ReceiptHandle: aws.String("3"),
							},
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"operation": {
										DataType:    aws.String("String"),
										StringValue: aws.String("task_2"),
									},
								},
								Body:          aws.String(`{"order_id":4}`),
								MessageId:     aws.String("4"),
								ReceiptHandle: aws.String("4"),
							},
						},
					},
					nil,
				).Once()

				mock.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://abcdefghijk"),
					ReceiptHandle: aws.String("1"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				mock.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://abcdefghijk"),
					ReceiptHandle: aws.String("2"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				mock.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://abcdefghijk"),
					ReceiptHandle: aws.String("3"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				mock.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://abcdefghijk"),
					ReceiptHandle: aws.String("4"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				return mock
			},
			want: func(s *SQSMock, l *logrus.Entry) *PollerSQS {
				l = l.WithField("_queue", "queue_1")
				return &PollerSQS{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(10),
					queueURL:            aws.String("https://abcdefghijk"),
					waitTimeSeconds:     aws.Int64(10),
					sleepOnError:        10 * time.Second,
					filters: []*PollerSQSFilter{
						{
							e: NewExec("php", []string{"testing/worker_1.php"}, l),
							f: "operation",
							v: "task_1",
						},
						{
							e: NewExec("php", []string{"testing/worker_2.php"}, l),
							f: "operation",
							v: "task_2",
						},
					},
				}
			},
		},
		{
			name: "2",
			args: args{
				c: config.Sqs{
					QueueName:     "queue_2",
					AttributeName: "operation_2",
					Commands: []config.SqsCommands{
						{
							AttributeValue: "task_1",
							Command:        "php",
							Args: []string{
								"testing/worker_1.php",
							},
						},
						{
							AttributeValue: "task_2",
							Command:        "php",
							Args: []string{
								"testing/worker_2.php",
							},
						},
					},
				},
			},
			mocker: func() *SQSMock {
				mock := &SQSMock{}

				mock.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_2"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://abcdefghijk"),
					},
					nil,
				).Once()

				mock.On("ReceiveMessageWithContext", context.TODO(), &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"operation_2": {
										DataType:    aws.String("String"),
										StringValue: aws.String("task_1"),
									},
								},
								Body:          aws.String(`{"order_id":21}`),
								MessageId:     aws.String("21"),
								ReceiptHandle: aws.String("21"),
							},
						},
					},
					nil,
				).Once()

				mock.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://abcdefghijk"),
					ReceiptHandle: aws.String("21"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				return mock
			},
			want: func(s *SQSMock, l *logrus.Entry) *PollerSQS {
				l = l.WithField("_queue", "queue_2")
				return &PollerSQS{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(1),
					waitTimeSeconds:     aws.Int64(0),
					sleepOnError:        10 * time.Second,
					queueURL:            aws.String("https://abcdefghijk"),
					filters: []*PollerSQSFilter{
						{
							e: NewExec("php", []string{"testing/worker_1.php"}, l),
							f: "operation_2",
							v: "task_1",
						},
						{
							e: NewExec("php", []string{"testing/worker_2.php"}, l),
							f: "operation_2",
							v: "task_2",
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := tt.mocker()
			logger := logrus.New().WithField("", "")
			p, err := NewPollerSQS(tt.args.c, mock, logger)
			if err != nil {
				t.Error(err)
			}
			w := tt.want(mock, logger)
			if !assert.EqualValues(t, p, w) {
				t.Errorf("a: %#v", p)
				t.Errorf("b: %#v", w)
			}

			p.poll(context.TODO())

			mock.AssertExpectations(t)
		})
	}
}
