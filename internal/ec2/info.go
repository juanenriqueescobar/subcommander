package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/juanenriqueescobar/subcommander/internal/config"
	"github.com/sirupsen/logrus"
)

type Info struct {
	id string
}

func (i Info) ID() string {
	return i.id
}

func GetInfo(c *ec2metadata.EC2Metadata, config *config.Config, logger *logrus.Logger) Info {
	if config.InstanceID.AWS {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		i, err := c.GetInstanceIdentityDocumentWithContext(ctx)
		if err != nil {
			logger.WithError(err).Trace("ec2info not load")
			return Info{}
		}
		return Info{
			id: i.InstanceID,
		}
	}
	logger.Debug("ec2info is disabled")
	return Info{}
}
