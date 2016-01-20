package discover

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var (
	creds = &credentials.EnvProvider{}
)

type Client struct {
	Region  string
	Cluster string

	creds *credentials.Credentials

	ec2 *ec2.EC2
	ecs *ecs.ECS

	tasks map[string]service
	sync.RWMutex
}

type Service struct {
	Endpoint string
	Tasks    []task
}

type Task struct {
	Name          string
	TaskARN       string
	IPAddress     string
	ContainerPort string
	HostPort      string
}

func NewClient(region, cluster string) *Client {
	return &Client{
		Region:  region,
		Cluster: cluster,
		creds:   &credentials.EnvProvider{},

		ec2: newEC2(region),
		ecs: newECS(region),

		tasks: make(map[string]Service),
	}
}

// Tasks returns the tasks available
func (c *Client) Tasks() map[string]Service {
	c.RLock()
	defer c.RUnlock()

	return c.tasks
}

func (c *Client) DiscoverECSTasks() error {
	var (
		// A set of map[ContainerArn]IPAddress
		containerInstances map[string]string
	)

	ContainerInstanceARNs, err := c.ContainerInstances()
	if err != nil {
		return err
	}

	containerInstances, err = c.createContainerInstances(ContainerInstanceARNs)
	if err != nil {
		return err
	}

	TaskARNs, err := c.TaskARNs()
	if err != nil {
		return err
	}

	Tasks, err := c.ECSTasks(TaskARNs)
	if err != nil {
		return err
	}

	TaskDefinitions, err := c.TaskDefinitions(Tasks)
	if err != nil {
		return err
	}

	for _, task := range Tasks {
		for _, container := range task.Containers {
			var (
				TaskName string
				HostPort int64
				Hostname string
			)

			// Don't discover ECS Containers that don't have a port exposed
			if len(container.NetworkBindings) == 0 {
				continue
			}

			for _, network := range container.NetworkBindings {
				// Create the name which is {task-id}-{container-port}
				Taskname = fmt.Sprintf("%s-%d",
					aws.StringValue(container.Name),
					aws.Int64Value(network.ContainerPort))
				HostPort = aws.Int64Value(network.HostPort)
			}

			// Check if a container has defined a Docker Label hostname to override the visibility
			if definition, ok := TaskDefinitions[aws.StringValue(task.TaskDefinitionArn)]; ok {
				for _, containerDefinition := range definition.ContainerDefinitions {
					if aws.StringValue(container.Name) != aws.StringValue(containerDefinition.Name) {
						continue
					}

					if host, ok := containerDefinition.DockerLabels["hostname"]; ok {
						hostname = aws.StringValue(host)
					}
					break
				}
			}

			if _, ok := c.tasks[TaskName]; !ok {
				c.tasks[TaskName] = Service{Tasks: []Task{}}
			}

			tasks := c.tasks[TaskName].Tasks
			tasks = append(tasks, Task{
				Name:      aws.StringValue(container.Name),
				ARN:       strings.Split(aws.StringValue(task.TaskArn), "/")[1],
				IPAddress: containerInstances[aws.StringValue(task.ContainerInstanceArn)],
				HostPort:  HostPort,
			})

			if hostname == "" {
				hostname = aws.StringValue(container.Name)
			}

			c.tasks[TaskName] = Service{
				Endpoint: hostname,
				Tasks:    tasks,
			}
		}
	}

	return nil
}

// createContainerInstances creates a map[ContainerInstanceArn]ipaddress
func (c *Client) createContainerInstances(ContainerInstanceARNs []*string) (map[string]string, error) {
	containerInstances := make(map[string]string)
	for _, arn := range ContainerInstanceARNs {
		instanceID, err := c.ContainerInstance(arn)
		if err != nil {
			return containerInstances, err
		}

		ipaddress, err := c.LookupInstanceIPAddress(instanceID)
		if err != nil {
			return containerInstances, err
		}

		containerInstances[aws.StringValue(arn)] = aws.StringValue(ipaddress)
	}

	return containerInstances, nil
}

// ECSTasks takes a list of TaskARNs and returns all the associated ECS Tasks
func (c *Client) ECSTasks(TaskARNs []*string) ([]*ecs.Task, error) {
	var tasks []*ecs.Task
	for _, arn := range TaskARNs {
		task, err := c.ECSTask(arn)
		if err != nil {
			return tasks, err
		}

		tasks = append(tasks, task.Tasks...)
	}

	return tasks, nil
}

// TaskDefinitions takes a list of Tasks and pulls the Task Definition for each
func (c *Client) TaskDefinitions(Tasks []*ecs.Task) (map[string]*ecs.TaskDefinition, error) {
	TaskDefinitions := make(map[string]*ecs.TaskDefinition)
	for _, task := range Tasks {
		TaskDefinition, err := c.TaskDefinition(task.TaskDefinitionArn)
		if err != nil {
			return TaskDefinitions, err
		}

		TaskDefinitions[aws.StringValue(task.TaskDefinitionArn)] = TaskDefinition
	}

	return TaskDefinitions, nil
}
