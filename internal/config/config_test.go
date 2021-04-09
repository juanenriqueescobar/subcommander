package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewConfigFromArgs(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				path: []string{"testdata/config_1.yml"},
			},
			want: &Config{
				Sqs: []Sqs{
					{
						QueueName:           "producto-api-order-operations",
						WaitTimeSeconds:     20,
						MaxNumberOfMessages: 5,
						WaitBetweenRequest:  5,
						AttributeName:       "operation",
						ThreadsNumber:       3,
						Commands: []SqsCommands{
							{
								AttributeValue: "change-state",
								Command:        "php",
								Args: []string{
									"/srv/www/producto-api/artisan",
									"order:change-state",
								},
							},
							{
								AttributeValue: "create-order-mu",
								Command:        "php",
								Args: []string{
									"/srv/www/producto-api/artisan",
									"order:create-order-mu",
								},
							},
						},
					},
					{
						QueueName:           "producto-api-catalog-products",
						WaitTimeSeconds:     0,
						MaxNumberOfMessages: 0,
						AttributeName:       "type",
						ThreadsNumber:       0,
						Commands: []SqsCommands{
							{
								AttributeValue: "product-create",
								Command:        "php",
								Args: []string{
									"/srv/www/producto-api/artisan",
									"product:create",
								},
							},
							{
								AttributeValue: "product-delete",
								Command:        "php",
								Args: []string{
									"/srv/www/producto-api/artisan",
									"product:delete",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no config",
			args: args{
				path: []string{},
			},
			wantErr: true,
		},
		{
			name: "not exists",
			args: args{
				path: []string{"xxx/yyy/zzz/1.yml"},
			},
			wantErr: true,
		},
		{
			name: "invalid",
			args: args{
				path: []string{"testdata/invalid.yml"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromArgs(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func Test_NewConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "not exists",
			args: args{
				path: "xxx/yyy/zzz/1.yml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, got, tt.want)
		})
	}
}
