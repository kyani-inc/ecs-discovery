package discover

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func newECS(region string) *ecs.ECS {
	session := session.New()
	return ecs.New(session, &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
}

// ContainerInstances fetches ALL ContainerARNs for a given cluster.
// The API limits to 100 items per call, this function will keep looping
// until everything is pulled.
func (c *Client) ContainerInstances() ([]*string, error) {
	var (
		ContainerInstanceARNs []*string
		NextToken             *string
	)
	for {
		instances, err := c.containerInstances(NextToken)
		if err != nil {
			return ContainerInstanceARNs, err
		}

		ContainerInstanceARNs = append(ContainerInstanceARNs, instances.ContainerInstanceArns...)
		if NextToken = instances.NextToken; NextToken == nil {
			break
		}
	}

	return ContainerInstanceARNs, nil
}

func (c *Client) containerInstances(NextToken *string) (*ecs.ListContainerInstancesOutput, error) {
	params := &ecs.ListContainerInstancesInput{
		Cluster:   aws.String(c.Cluster),
		NextToken: NextToken,
	}
	resp, err := c.ecs.ListContainerInstances(params)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// ContainerInstance takes a Container Instance ARN and returns the EC2
// Instance ID attached to the ARN
func (c *Client) ContainerInstance(ContainerInstanceARN *string) (*string, error) {
	params := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(c.Cluster),
		ContainerInstances: []*string{ContainerInstanceARN},
	}
	resp, err := c.ecs.DescribeContainerInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.ContainerInstances) == 0 {
		return nil, fmt.Errorf("no container instances were found")
	}

	return resp.ContainerInstances[0].Ec2InstanceId, nil
}

// TaskARNs fetches ALL TaskARNs for a given cluster.
// The API limits to 100 items per call, this function will keep looping
// until everything is pulled.
func (c *Client) TaskARNs() ([]*string, error) {
	var (
		TaskARNs  []*string
		NextToken *string
	)
	for {
		tasks, err := c.taskARNs(NextToken)
		if err != nil {
			return TaskARNs, err
		}

		TaskARNs = append(TaskARNs, tasks.TaskArns...)
		if NextToken = tasks.NextToken; NextToken == nil {
			break
		}
	}

	return TaskARNs, nil
}

func (c *Client) taskARNs(NextToken *string) (*ecs.ListTasksOutput, error) {
	params := &ecs.ListTasksInput{
		Cluster:   aws.String(c.Cluster),
		NextToken: NextToken,
	}
	resp, err := c.ecs.ListTasks(params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ECSTask takes a TaskARN and returns a Task
func (c *Client) ECSTask(TaskARN *string) (*ecs.Task, error) {
	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(c.Cluster),
		Tasks:   []*string{TaskARN},
	}
	resp, err := c.ecs.DescribeTasks(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found for given task arn")
	}

	return resp.Tasks[0], nil
}

// TaskDefinition takes a TaskDefinitionARN and returns an associated Task Definition
func (c *Client) TaskDefinition(TaskDefinitionARN *string) (*ecs.TaskDefinition, error) {
	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: TaskDefinitionARN,
	}
	resp, err := c.ecs.DescribeTaskDefinition(params)
	if err != nil {
		return nil, err
	}

	return resp.TaskDefinition, nil
}
