package discovery

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
	"github.com/kyani-inc/ecs-discovery/discover"
)

// oldTasks holds the tasks of the last time it ran to do a diff.
var oldTasks map[string]discover.Service

type consulKV struct {
	kv *consul.KV

	discovery
}

// ConsulKV calls KV() which creates a new KV instances returns
// the Discoverer interface. It takes the consul Config struct.
func ConsulKV(config *consul.Config) (Discoverer, error) {
	var err error

	consulkv := &consulKV{}
	consulkv.kv, err = consulkv.KV(config)
	return consulkv, err
}

// KV creates a new consul/api/KV instance and returns it with any errors.
func (*consulKV) KV(config *consul.Config) (*consul.KV, error) {
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client.KV(), nil
}

func (kv *consulKV) Discover() error {
	if kv.Client == nil {
		return fmt.Errorf(ERR_NO_CLIENT)
	}

	err := kv.Client.DiscoverECSTasks()
	if err != nil {
		return err
	}
	newTasks := kv.Client.Tasks()

	kv.CompareForDeletion(oldTasks, newTasks)
	kv.CompareForAddition(oldTasks, newTasks)

	oldTasks = newTasks
	return nil
}

// CompareForDeletion does a diff against the last run vs the current run, it
// then determines which keys/trees need to be deleted.
func (kv *consulKV) CompareForDeletion(old, canon map[string]discover.Service) {
	for k, v := range old {
		_, ok := canon[k]
		if !ok {
			kv.deleteTree(k)
			continue
		}

		vv := canon[k].Endpoint
		if v.Endpoint != vv {
			kv.deleteKey(fmt.Sprintf("%s/%s", k, v))
		}

		for _, task := range v.Tasks {
			if !kv.inTasks(task, canon[k].Tasks) {
				kv.deleteKey(fmt.Sprintf("%s/instances/%s", k, task.TaskARN))
			}
		}
	}
}

// CompareForAddition does a diff against the last run vs the current run, it
// then determines which keys need to be added.
func (kv *consulKV) CompareForAddition(old, canon map[string]discover.Service) {
	for k, v := range canon {
		if _, ok := old[k]; !ok {
			old[k] = discover.Service{
				Tasks: []discover.Task{},
			}
			kv.addKey(fmt.Sprintf("%s/endpoint", k), []byte(canon[k].Endpoint))
		}

		for _, task := range v.Tasks {
			if !kv.inTasks(task, old[k].Tasks) {
				ip := fmt.Sprintf("%s:%d", task.IPAddress, task.HostPort)
				kv.addKey(fmt.Sprintf("%s/instances/%s", k, task.TaskARN), []byte(ip))
			}
		}
	}
}

func (*consulKV) inTasks(task discover.Task, tasks []discover.Task) bool {
	for _, t := range tasks {
		if t == task {
			return true
		}
	}
	return false
}

// Delete an entire tree of prefixes.
func (kv *consulKV) deleteTree(prefix string) error {
	_, err := kv.kv.DeleteTree(kv.Client.Cluster+"/"+prefix, &consul.WriteOptions{})
	return err
}

// Delete a specific key
func (kv *consulKV) deleteKey(key string) error {
	_, err := kv.kv.Delete(kv.Client.Cluster+"/"+key, &consul.WriteOptions{})
	return err
}

// Create a key / value
func (kv *consulKV) addKey(key string, value []byte) error {
	_, err := kv.kv.Put(&consul.KVPair{
		Key:   key,
		Value: value,
	}, &consul.WriteOptions{})
	return err
}
