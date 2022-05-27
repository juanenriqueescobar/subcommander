package di

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/wire"
	"github.com/juanenriqueescobar/subcommander/internal/commander"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/juanenriqueescobar/subcommander/internal/pollers"
	"github.com/juanenriqueescobar/subcommander/internal/providers"
)

var stdset = wire.NewSet(
	logger2,
	readers,
	ctx,
	config.NewConfigFromArgs,

	sqs.New,
	session.NewSession,
	awsConfig,

	commander.NewCommander,

	wire.Bind(new(pollers.Sqs), new(*sqs.SQS)),
	wire.Struct(new(pollers.SqsPollerConstructor), "*"),

	wire.Bind(new(client.ConfigProvider), new(*session.Session)),

	providers.SqsPollerProvider,
	wire.Bind(new(providers.SqsConstructor), new(*pollers.SqsPollerConstructor)),
	wire.Bind(new(providers.Sqs), new(*sqs.SQS)),

	ec2metadata.New,
)
