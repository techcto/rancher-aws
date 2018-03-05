package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/crewjam/ec2cluster"
)

// handleLifecycleEvent is invoked whenever we get a lifecycle terminate message. It removes
// terminated instances from the etcd cluster.
func handleLifecycleEvent(m *ec2cluster.LifecycleMessage) (shouldContinue bool, err error) {
	if m.LifecycleTransition != "autoscaling:EC2_INSTANCE_TERMINATING" {
		return true, nil
	}

	// look for the instance in the cluster
	resp, err := http.Get(fmt.Sprintf("%s/v2/members", etcdLocalURL))
	if err != nil {
		return false, err
	}
	members := etcdMembers{}
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return false, err
	}
	memberID := ""
	for _, member := range members.Members {
		if member.Name == m.EC2InstanceID {
			memberID = member.ID
		}
	}

	if memberID == "" {
		log.WithField("InstanceID", m.EC2InstanceID).Warn("received termination event for non-member")
		return true, nil
	}

	log.WithFields(log.Fields{
		"InstanceID": m.EC2InstanceID,
		"MemberID":   memberID}).Info("removing from cluster")
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v2/members/%s", etcdLocalURL, memberID), nil)
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	return false, nil
}

func watchLifecycleEvents(s *ec2cluster.Cluster, localInstance *ec2.Instance) {
	etcdLocalURL = fmt.Sprintf("http://%s:2379", *localInstance.PrivateIpAddress)
	for {
		queueUrl, err := s.LifecycleEventQueueURL()

		// The lifecycle hook might not exist yet if we're being created
		// by cloudformation.
		if err == ec2cluster.ErrLifecycleHookNotFound {
			log.Printf("WARNING: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if err != nil {
			log.Fatalf("ERROR: LifecycleEventQueueUrl: %s", err)
		}
		log.Printf("Found Lifecycle SQS Queue: %s", queueUrl)

		err = s.WatchLifecycleEvents(queueUrl, handleLifecycleEvent)

		if err != nil {
			log.Fatalf("ERROR: WatchLifecycleEvents: %s", err)
		}
		panic("not reached")
	}
}
