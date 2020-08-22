//+build wireinject

package di

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/wire"
	"github.com/juanenriqueescobar/subcommander/internal"
	"github.com/juanenriqueescobar/subcommander/internal/commander"
	"github.com/juanenriqueescobar/subcommander/internal/config"
)

var stdset = wire.NewSet(
	logger2,
	readers,
	ctx,
	config.NewConfigFromArgs,
	internal.PollerSQSBuilder,

	sqs.New,
	session.NewSession,
	awsConfig,

	commander.NewCommander,

	wire.Bind(new(internal.SQS), new(*sqs.SQS)),
	wire.Bind(new(client.ConfigProvider), new(*session.Session)),

	ec2metadata.New,
)
