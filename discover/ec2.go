package discover

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func newEC2(region string) *ec2.EC2 {
	session := session.New()
	return ec2.New(session, &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
}

// LookupInstanceIPAddress takes an Instance ID and returns the Private IP Address
func (c *Client) LookupInstanceIPAddress(instanceID *string) (*string, error) {
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{instanceID},
	}
	resp, err := c.ec2.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return nil, fmt.Errorf("no reservations found")
	}

	if len(resp.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no instances found")
	}

	return resp.Reservations[0].Instances[0].PrivateIpAddress, nil
}
