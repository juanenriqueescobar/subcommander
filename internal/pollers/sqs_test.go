package pollers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/juanenriqueescobar/subcommander/internal/executor"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SqsMock struct {
	mock.Mock
}

func (m *SqsMock) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *SqsMock) ReceiveMessageWithContext(ctx aws.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *SqsMock) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

type ExecutorMock struct {
	mock.Mock
}

func (m *ExecutorMock) Run(id string, in executor.StdinData) (bool, error) {
	args := m.Called(id, in)
	return args.Bool(0), args.Error(1)
}

func TestSqsPoller_Run(t *testing.T) {
	type fields struct {
		waitBetweenRequest time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		mocker func(context.Context, context.CancelFunc) *SqsMock
	}{
		{
			"1",
			fields{
				waitBetweenRequest: time.Microsecond,
			},
			func(ctx context.Context, cancel context.CancelFunc) *SqsMock {
				counter := 0
				mocksqs := &SqsMock{}
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
			func(ctx context.Context, cancel context.CancelFunc) *SqsMock {
				counter := 0
				mocksqs := &SqsMock{}
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
			func(ctx context.Context, cancel context.CancelFunc) *SqsMock {
				mocksqs := &SqsMock{}
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
			func(ctx context.Context, cancel context.CancelFunc) *SqsMock {
				mocksqs := &SqsMock{}
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
			func(ctx context.Context, cancel context.CancelFunc) *SqsMock {
				mocksqs := &SqsMock{}
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

			p := &SqsPoller{
				client:              mock,
				queueURL:            aws.String("https://abcdefghijk"),
				waitTimeSeconds:     aws.Int64(0),
				logger:              logrus.New().WithField("", ""),
				filters:             []*SqsFilter{},
				maxNumberOfMessages: aws.Int64(1),
				sleepOnError:        time.Hour,
				waitBetweenRequest:  tt.fields.waitBetweenRequest,
			}
			p.Run(ctx)

			mock.AssertExpectations(t)
		})
	}
}

func TestSqsPoller_poll(t *testing.T) {
	type fields struct {
		client              *SqsMock
		queueURL            *string
		waitTime            *int64
		maxNumberOfMessages *int64
		logger              *logrus.Entry
		filters             []*SqsFilter
		sleepOnError        time.Duration
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		wantWait bool
		mocker   func(*SqsMock, []*SqsFilter) context.Context
	}{
		{
			name: "one msg, no filter for message, does not delete it",
			fields: fields{
				client:              &SqsMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters:             []*SqsFilter{},
				sleepOnError:        1 * time.Microsecond,
			},
			wantErr:  false,
			wantWait: false,
			mocker: func(m *SqsMock, f []*SqsFilter) context.Context {
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
				client:              &SqsMock{},
				queueURL:            aws.String("https://2"),
				waitTime:            aws.Int64(20),
				maxNumberOfMessages: aws.Int64(10),
				logger:              logrus.New().WithField("", ""),
				filters: []*SqsFilter{
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
			wantErr:  false,
			wantWait: false,
			mocker: func(m *SqsMock, f []*SqsFilter) context.Context {
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

				f[0].e.(*ExecutorMock).On("Run", "i2", executor.StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":2}`}).Return(false, nil).Once()

				return ctx
			},
		},
		{
			name: "one msg, with filter, delete it",
			fields: fields{
				client:              &SqsMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters: []*SqsFilter{
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
			wantErr:  false,
			wantWait: false,
			mocker: func(m *SqsMock, f []*SqsFilter) context.Context {
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

				f[1].e.(*ExecutorMock).On("Run", "i1", executor.StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":1}`}).Return(true, nil).Once()

				return ctx
			},
		},
		{
			name: "multiple msg, with filter, delete it",
			fields: fields{
				client:              &SqsMock{},
				queueURL:            aws.String("https://1"),
				waitTime:            aws.Int64(1),
				maxNumberOfMessages: aws.Int64(2),
				logger:              logrus.New().WithField("", ""),
				filters: []*SqsFilter{
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
			wantErr:  false,
			wantWait: false,
			mocker: func(m *SqsMock, f []*SqsFilter) context.Context {
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

				f[1].e.(*ExecutorMock).On("Run", "i1", executor.StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":1}`}).Return(true, nil).Once()
				f[1].e.(*ExecutorMock).On("Run", "i2", executor.StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":2}`}).Return(true, nil).Once()
				f[0].e.(*ExecutorMock).On("Run", "i3", executor.StdinData{Metadata: map[string]interface{}{}, Body: `{"order_id":3}`}).Return(true, nil).Once()

				return ctx
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SqsPoller{
				client:              tt.fields.client,
				queueURL:            tt.fields.queueURL,
				waitTimeSeconds:     tt.fields.waitTime,
				maxNumberOfMessages: tt.fields.maxNumberOfMessages,
				logger:              tt.fields.logger,
				filters:             tt.fields.filters,
				sleepOnError:        tt.fields.sleepOnError,
			}
			ctx := tt.mocker(tt.fields.client, tt.fields.filters)
			wait, err := p.poll(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("PollerSQS.poll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if wait != tt.wantWait {
				t.Errorf("PollerSQS.poll() wait = %v, wantWait %v", wait, tt.wantWait)
			}

			tt.fields.client.AssertExpectations(t)
			for _, f := range tt.fields.filters {
				f.e.(*ExecutorMock).AssertExpectations(t)
			}

		})
	}
}

func TestSqsPollerConstructor_New(t *testing.T) {
	type args struct {
		pc       SqsConfig
		commands []config.SqsCommands
	}
	tests := []struct {
		name string
		args args
		want func(*SqsMock, *logrus.Entry) *SqsPoller
	}{
		{
			name: "1",
			args: args{
				pc: SqsConfig{
					AttributeName:       "name",
					MaxNumberOfMessages: 10,
					QueueURL:            aws.String("https://queue_1"),
					VisibilityTimeout:   60,
					WaitTimeSeconds:     10,
					WaitBetweenRequest:  10,
				},
				commands: []config.SqsCommands{
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
			want: func(s *SqsMock, l *logrus.Entry) *SqsPoller {
				return &SqsPoller{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(10),
					queueURL:            aws.String("https://queue_1"),
					waitTimeSeconds:     aws.Int64(10),
					sleepOnError:        time.Duration(10 * time.Second),
					waitBetweenRequest:  time.Duration(10 * time.Second),
					filters: []*SqsFilter{
						{
							e: executor.NewExec("cmd_1", []string{"a1", "b1"}, 60*time.Second, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_1"})),
							f: "name",
							v: "value_1",
						},
						{
							e: executor.NewExec("cmd_2", []string{"a2", "b2"}, 60*time.Second, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_2"})),
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
				pc: SqsConfig{
					AttributeName:       "name",
					MaxNumberOfMessages: 10,
					QueueURL:            aws.String("https://queue_1"),
					VisibilityTimeout:   60,
					WaitTimeSeconds:     10,
					WaitBetweenRequest:  10,
				},
				commands: []config.SqsCommands{
					{
						AttributeValue: "value_1",
						Command:        "cmd_1",
						Args: []string{
							"a1",
							"b1",
						},
						Timeout: 10000 * time.Second,
					},
					{
						AttributeValue: "value_2",
						Command:        "cmd_2",
						Args: []string{
							"a2",
							"b2",
						},
						Timeout: 20000 * time.Second,
					},
				},
			},
			want: func(s *SqsMock, l *logrus.Entry) *SqsPoller {
				return &SqsPoller{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(10),
					queueURL:            aws.String("https://queue_1"),
					waitTimeSeconds:     aws.Int64(10),
					sleepOnError:        time.Duration(10 * time.Second),
					waitBetweenRequest:  time.Duration(10 * time.Second),
					filters: []*SqsFilter{
						{
							e: executor.NewExec("cmd_1", []string{"a1", "b1"}, 60*time.Second, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_1"})),
							f: "name",
							v: "value_1",
						},
						{
							e: executor.NewExec("cmd_2", []string{"a2", "b2"}, 60*time.Second, l.WithFields(logrus.Fields{"sc_task.attr_name": "name", "sc_task.attr_value": "value_2"})),
							f: "name",
							v: "value_2",
						},
					},
				}
			},
		},
		{
			name: "3",
			args: args{
				pc: SqsConfig{
					AttributeName: "operation",
					QueueURL:      aws.String("https://queue_2"),
				},
				commands: []config.SqsCommands{
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
			want: func(s *SqsMock, l *logrus.Entry) *SqsPoller {
				return &SqsPoller{
					client:              s,
					logger:              l,
					maxNumberOfMessages: aws.Int64(1),
					waitTimeSeconds:     aws.Int64(0),
					sleepOnError:        time.Duration(10 * time.Second),
					queueURL:            aws.String("https://queue_2"),
					filters: []*SqsFilter{
						{
							e: executor.NewExec("php", []string{"testing/worker_1.php"}, 0, l.WithFields(logrus.Fields{"sc_task.attr_name": "operation", "sc_task.attr_value": "task_1"})),
							f: "operation",
							v: "task_1",
						},
						{
							e: executor.NewExec("php", []string{"testing/worker_2.php"}, 0, l.WithFields(logrus.Fields{"sc_task.attr_name": "operation", "sc_task.attr_value": "task_2"})),
							f: "operation",
							v: "task_2",
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New().WithField("", "")
			c := &SqsMock{}
			b := &SqsPollerConstructor{
				Client: c,
			}
			want := tt.want(c, logger)
			got := b.New(tt.args.pc, tt.args.commands, logger)

			if !assert.EqualValues(t, want, got) {
				t.Errorf("a: %#v", got)
				t.Errorf("b: %#v", want)
			}

			c.AssertExpectations(t)
		})
	}
}
