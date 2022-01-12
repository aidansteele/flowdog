package lambda_acceptor

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
)

type Enricher struct {
	// TODO: obviously this should update live in response to eventbridge events
	cache map[string]*ec2.Instance
}

func NewEnricher(api ec2iface.EC2API) (*Enricher, error) {
	m := map[string]*ec2.Instance{}

	err := api.DescribeInstancesPages(&ec2.DescribeInstancesInput{}, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, r := range page.Reservations {
			for _, i := range r.Instances {
				instance := i
				for _, iface := range i.NetworkInterfaces {
					m[*iface.PrivateIpAddress] = instance
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Enricher{cache: m}, nil
}

func (e *Enricher) InstanceByIp(ip string) *ec2.Instance {
	return e.cache[ip]
}
