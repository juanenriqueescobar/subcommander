package internal

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SQSMock struct {
	mock.Mock
}

func (m *SQSMock) GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.GetQueueUrlOutput), args.Error(1)
}

func (m *SQSMock) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *SQSMock) ReceiveMessageWithContext(ctx aws.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *SQSMock) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

type ExecutorMock struct {
	mock.Mock
}

func (m *ExecutorMock) run(id string, in StdinData) (bool, error) {
	args := m.Called(id, in)
	return args.Bool(0), args.Error(1)
}

func TestPollerSQS_Run(t *testing.T) {
	type fields struct {
		waitBetweenRequest time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		mocker func(context.Context, context.CancelFunc) *SQSMock
	}{
		{
			"1",
			fields{
				waitBetweenRequest: time.Microsecond,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SQSMock {
				counter := 0
				mocksqs := &SQSMock{}
				mocksqs.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{},
					},
					nil,
				).Times(10).Run(func(args mock.Arguments) {
					counter++
					if counter == 10 {
						cancel()
					}
				})
				return mocksqs
			},
		},
		{
			"2",
			fields{
				waitBetweenRequest: time.Microsecond,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SQSMock {
				counter := 0
				mocksqs := &SQSMock{}
				mocksqs.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{},
					},
					nil,
				).Times(100).Run(func(args mock.Arguments) {
					counter++
					if counter == 100 {
						cancel()
					}
				})
				return mocksqs
			},
		},
		{
			"context canceled",
			fields{
				waitBetweenRequest: time.Hour,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SQSMock {
				mocksqs := &SQSMock{}
				mocksqs.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{},
					awserr.New(request.CanceledErrorCode, "", nil),
				).Once().Run(func(args mock.Arguments) {
					cancel()
				})
				return mocksqs
			},
		},
		{
			"request failed",
			fields{
				waitBetweenRequest: time.Microsecond,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SQSMock {
				mocksqs := &SQSMock{}
				mocksqs.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{},
					errors.New("crash"),
				).Once()
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
				return mocksqs
			},
		},
		{
			"context canceled while waiting between requests",
			fields{
				waitBetweenRequest: time.Hour,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SQSMock {
				mocksqs := &SQSMock{}
				mocksqs.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(1),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://abcdefghijk"),
					WaitTimeSeconds:       aws.Int64(0),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{},
					nil,
				)
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
				return mocksqs
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			mock := tt.mocker(ctx, cancel)

			p := &PollerSQS{
				client:              mock,
				queueURL:            aws.String("https://abcdefghijk"),
				waitTimeSeconds:     aws.Int64(0),
				logger:              logrus.New().WithField("", ""),
				filters:             []*PollerSQSFilter{},
				maxNumberOfMessages: aws.Int64(1),
				sleepOnError:        time.Hour,
				waitBetweenRequest:  tt.fields.waitBetweenRequest,
			}
			p.Run(ctx)

			mock.AssertExpectations(t)
		})
	}
}

func TestPollerSQS_poll(t *testing.T) {
	type fields struct {
		client              *SQSMock
		queueURL            *string
		waitTime            *int64
		maxNumberOfMessages *int64
		logger              *logrus.Entry
		filters             []*PollerSQSFilter
		sleepOnError        time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		mocker  func(*SQSMock, []*PollerSQSFilter) context.Context
	}{
		{
			name: "one msg, no filter for message, does not delete it",
			fields: fields{
				client:              &SQSMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters:             []*PollerSQSFilter{},
				sleepOnError:        1 * time.Microsecond,
			},
			wantErr: false,
			mocker: func(m *SQSMock, f []*PollerSQSFilter) context.Context {
				ctx := context.Background()

				m.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(2),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://1"),
					WaitTimeSeconds:       aws.Int64(1),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_a"),
									},
								},
								Body:          aws.String(`{"order_id":1}`),
								MessageId:     aws.String("1"),
								ReceiptHandle: aws.String("1"),
							},
						},
					},
					nil,
				).Once()

				return ctx
			},
		},
		{
			name: "one msg, with filter, error, does not delete it",
			fields: fields{
				client:              &SQSMock{},
				queueURL:            aws.String("https://2"),
				waitTime:            aws.Int64(20),
				maxNumberOfMessages: aws.Int64(10),
				logger:              logrus.New().WithField("", ""),
				filters: []*PollerSQSFilter{
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_b",
					},
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_a",
					},
				},
				sleepOnError: 1 * time.Microsecond,
			},
			wantErr: false,
			mocker: func(m *SQSMock, f []*PollerSQSFilter) context.Context {
				ctx := context.Background()

				m.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(10),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://2"),
					WaitTimeSeconds:       aws.Int64(20),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_b"),
									},
								},
								Body:          aws.String(`{"order_id":2}`),
								MessageId:     aws.String("i2"),
								ReceiptHandle: aws.String("r2"),
							},
						},
					},
					nil,
				).Once()

				f[0].e.(*ExecutorMock).On("run", "i2", StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":2}`}).Return(false, nil).Once()

				return ctx
			},
		},
		{
			name: "one msg, with filter, delete it",
			fields: fields{
				client:              &SQSMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters: []*PollerSQSFilter{
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_b",
					},
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_a",
					},
				},
				sleepOnError: 1 * time.Microsecond,
			},
			wantErr: false,
			mocker: func(m *SQSMock, f []*PollerSQSFilter) context.Context {
				ctx := context.Background()

				m.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(2),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://1"),
					WaitTimeSeconds:       aws.Int64(1),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_a"),
									},
								},
								Body:          aws.String(`{"order_id":1}`),
								MessageId:     aws.String("i1"),
								ReceiptHandle: aws.String("r1"),
							},
						},
					},
					nil,
				).Once()

				m.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://1"),
					ReceiptHandle: aws.String("r1"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				f[1].e.(*ExecutorMock).On("run", "i1", StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":1}`}).Return(true, nil).Once()

				return ctx
			},
		},
		{
			name: "multiple msg, with filter, delete it",
			fields: fields{
				client:              &SQSMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters: []*PollerSQSFilter{
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_b",
					},
					{
						e: &ExecutorMock{},
						f: "type",
						v: "type_a",
					},
				},
				sleepOnError: 1 * time.Microsecond,
			},
			wantErr: false,
			mocker: func(m *SQSMock, f []*PollerSQSFilter) context.Context {
				ctx := context.Background()

				m.On("ReceiveMessageWithContext", ctx, &sqs.ReceiveMessageInput{
					MaxNumberOfMessages:   aws.Int64(2),
					MessageAttributeNames: []*string{aws.String("All")},
					QueueUrl:              aws.String("https://1"),
					WaitTimeSeconds:       aws.Int64(1),
				}, []request.Option(nil)).Return(
					&sqs.ReceiveMessageOutput{
						Messages: []*sqs.Message{
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_a"),
									},
								},
								Body:          aws.String(`{"order_id":1}`),
								MessageId:     aws.String("i1"),
								ReceiptHandle: aws.String("r1"),
							},
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_a"),
									},
								},
								Body:          aws.String(`{"order_id":2}`),
								MessageId:     aws.String("i2"),
								ReceiptHandle: aws.String("r2"),
							},
							{
								MessageAttributes: map[string]*sqs.MessageAttributeValue{
									"type": {
										DataType:    aws.String("String"),
										StringValue: aws.String("type_b"),
									},
								},
								Body:          aws.String(`{"order_id":3}`),
								MessageId:     aws.String("i3"),
								ReceiptHandle: aws.String("r3"),
							},
						},
					},
					nil,
				).Once()

				m.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://1"),
					ReceiptHandle: aws.String("r1"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				m.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://1"),
					ReceiptHandle: aws.String("r2"),
				}).Return(&sqs.DeleteMessageOutput{}, nil).Once()

				m.On("DeleteMessage", &sqs.DeleteMessageInput{
					QueueUrl:      aws.String("https://1"),
					ReceiptHandle: aws.String("r3"),
				}).Return(&sqs.DeleteMessageOutput{}, errors.New("crash")).Once()

				f[1].e.(*ExecutorMock).On("run", "i1", StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":1}`}).Return(true, nil).Once()
				f[1].e.(*ExecutorMock).On("run", "i2", StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":2}`}).Return(true, nil).Once()
				f[0].e.(*ExecutorMock).On("run", "i3", StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":3}`}).Return(true, nil).Once()

				return ctx
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PollerSQS{
				client:              tt.fields.client,
				queueURL:            tt.fields.queueURL,
				waitTimeSeconds:     tt.fields.waitTime,
				maxNumberOfMessages: tt.fields.maxNumberOfMessages,
				logger:              tt.fields.logger,
				filters:             tt.fields.filters,
				sleepOnError:        tt.fields.sleepOnError,
			}
			ctx := tt.mocker(tt.fields.client, tt.fields.filters)
			if err := p.poll(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PollerSQS.poll() error = %v, wantErr %v", err, tt.wantErr)
			}

			tt.fields.client.AssertExpectations(t)
			for _, f := range tt.fields.filters {
				f.e.(*ExecutorMock).AssertExpectations(t)
			}

		})
	}
}

func TestPollerSQS_NewPollerSQS(t *testing.T) {
	type args struct {
		c      config.Sqs
		client *SQSMock
	}
	tests := []struct {
		name    string
		args    args
		want    func(*SQSMock, *logrus.Entry) *PollerSQS
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				c: config.Sqs{
					QueueName:           "queue_1",
					WaitTimeSeconds:     10,
					MaxNumberOfMessages: 10,
					WaitBetweenRequest:  10,
					AttributeName:       "name",
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
				client: &SQSMock{},
			},
			want: func(s *SQSMock, l *logrus.Entry) *PollerSQS {

				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_1"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_1"),
					},
					nil,
				).Once()

				l = l.WithField("_queue", "queue_1")
				return &PollerSQS{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(10),
					queueURL:            aws.String("https://queue_1"),
					waitTimeSeconds:     aws.Int64(10),
					sleepOnError:        time.Duration(10 * time.Second),
					waitBetweenRequest:  time.Duration(10 * time.Second),
					filters: []*PollerSQSFilter{
						{
							e: NewExec("cmd_1", []string{"a1", "b1"}, l),
							f: "name",
							v: "value_1",
						},
						{
							e: NewExec("cmd_2", []string{"a2", "b2"}, l),
							f: "name",
							v: "value_2",
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
					AttributeName: "operation",
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
				client: &SQSMock{},
			},
			want: func(s *SQSMock, l *logrus.Entry) *PollerSQS {
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_2"),
				}).Return(
					&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String("https://queue_2"),
					},
					nil,
				).Once()

				l = l.WithField("_queue", "queue_2")
				return &PollerSQS{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(1),
					waitTimeSeconds:     aws.Int64(0),
					sleepOnError:        time.Duration(10 * time.Second),
					queueURL:            aws.String("https://queue_2"),
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
			name: "3",
			args: args{
				c: config.Sqs{
					QueueName:     "queue_2",
					AttributeName: "operation",
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
				client: &SQSMock{},
			},
			wantErr: true,
			want: func(s *SQSMock, l *logrus.Entry) *PollerSQS {
				s.On("GetQueueUrl", &sqs.GetQueueUrlInput{
					QueueName: aws.String("queue_2"),
				}).Return(
					&sqs.GetQueueUrlOutput{},
					errors.New("crash"),
				).Once()

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New().WithField("", "")
			w := tt.want(tt.args.client, logger)
			p, err := NewPollerSQS(tt.args.c, tt.args.client, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !assert.EqualValues(t, p, w) {
				t.Errorf("a: %#v", p)
				t.Errorf("b: %#v", w)
			}

			tt.args.client.AssertExpectations(t)
		})
	}
}

func TestPollerSQSBuilder(t *testing.T) {
	type args struct {
		c *config.Config
		s SQS
		l *logrus.Entry
	}
	tests := []struct {
		name    string
		args    args
		want    []*PollerSQS
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PollerSQSBuilder(tt.args.c, tt.args.s, tt.args.l)
			if (err != nil) != tt.wantErr {
				t.Errorf("PollerSQSBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PollerSQSBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}
