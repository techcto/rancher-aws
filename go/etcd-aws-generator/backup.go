package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/coreos/go-etcd/etcd"
	"github.com/crewjam/ec2cluster"
)

// backupService invokes backupOnce() periodically if the current node is the cluster leader.
func backupService(s *ec2cluster.Cluster, backupBucket, backupKey, dataDir string, interval time.Duration) error {
	instance, err := s.Instance()
	if err != nil {
		return err
	}

	ticker := time.Tick(interval)
	for {
		<-ticker

		resp, err := http.Get(fmt.Sprintf("http://%s:2379/v2/stats/self", *instance.PrivateIpAddress))
		if err != nil {
			return fmt.Errorf("%s: http://%s:2379/v2/stats/self: %s", *instance.InstanceId,
				*instance.PrivateIpAddress, err)
		}

		nodeState := etcdState{}
		if err := json.NewDecoder(resp.Body).Decode(&nodeState); err != nil {
			return fmt.Errorf("%s: http://%s:2379/v2/stats/self: %s", *instance.InstanceId,
				*instance.PrivateIpAddress, err)
		}

		// if the cluster has a leader other than the current node, then don't do the
		// backup.
		if nodeState.LeaderInfo.Leader != "" && nodeState.ID != nodeState.LeaderInfo.Leader {
			log.Printf("backup: %s: http://%s:2379/v2/stats/self: not the leader", *instance.InstanceId,
				*instance.PrivateIpAddress)
			<-ticker
			continue
		}

		if err := backupOnce(s, backupBucket, backupKey, dataDir); err != nil {
			return err
		}
	}
	panic("not reached")
}

// getInstanceTag returns the first occurrence of the specified tag on the instance
// or an empty string if the tag is not found.
func getInstanceTag(instance *ec2.Instance, tagName string) string {
	rv := ""
	for _, tag := range instance.Tags {
		if *tag.Key == tagName {
			rv = *tag.Value
		}
	}
	return rv
}

// dumpEtcdNode writes a JSON representation of the nodes and and below `key`
// to `w`. Returns the number of nodes traversed.
func dumpEtcdNode(key string, etcdClient *etcd.Client, w io.Writer) (int, error) {
	response, err := etcdClient.Get(key, false, false)
	if err != nil {
		return 0, err
	}

	childNodes := response.Node.Nodes
	response.Node.Nodes = nil
	if err := json.NewEncoder(w).Encode(response.Node); err != nil {
		return 0, err
	}
	count := 1

	// enumerate all the child nodes. If a child node is
	// a directory, then it must be backed up recursively.
	// Otherwise it can just be written
	for _, childNode := range childNodes {
		if childNode.Dir {
			c, err := dumpEtcdNode(childNode.Key, etcdClient, w)
			if err != nil {
				return 0, err
			}
			count += c
		} else {
			if err := json.NewEncoder(w).Encode(childNode); err != nil {
				return 0, err
			}
			count += 1
		}
	}

	return count, nil
}

// backupOnce dumps all the nodes in the etcd cluster to the specified S3 bucket. On
// success it emits a CloudWatch metric for the number of keys backed up. The absence
// of data on this metric indicates the backup has failed.
func backupOnce(s *ec2cluster.Cluster, backupBucket, backupKey, dataDir string) error {
	instance, err := s.Instance()
	if err != nil {
		return err
	}
	etcdClient := etcd.NewClient([]string{fmt.Sprintf("http://%s:2379", *instance.PrivateIpAddress)})
	if success := etcdClient.SyncCluster(); !success {
		return fmt.Errorf("backupOnce: cannot sync machines")
	}

	var valueCount int
	bodyReader, bodyWriter := io.Pipe()
	go func() {
		gzWriter := gzip.NewWriter(bodyWriter)

		var err error
		valueCount, err = dumpEtcdNode("/", etcdClient, gzWriter)
		if err != nil {
			gzWriter.Close()
			bodyWriter.CloseWithError(err)
			return
		}

		gzWriter.Close()
		bodyWriter.Close()
	}()

	s3mgr := s3manager.NewUploader(s.AwsSession)
	_, err = s3mgr.Upload(&s3manager.UploadInput{
		Key:                  aws.String(backupKey),
		Bucket:               aws.String(backupBucket),
		Body:                 bodyReader,
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAes256),
		ACL:                  aws.String(s3.ObjectCannedACLPrivate),
	})
	if err != nil {
		return fmt.Errorf("upload s3://%s%s: %s", backupBucket, backupKey, err)
	}

	log.Printf("backup written to s3://%s%s (%d values)", backupBucket, backupKey,
		valueCount)

	cloudwatch.New(s.AwsSession).PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("Local/etcd"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("BackupKeyCount"),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("AutoScalingGroupName"),
						Value: aws.String(getInstanceTag(instance, "aws:autoscaling:groupName")),
					},
				},
				Unit:  aws.String(cloudwatch.StandardUnitCount),
				Value: aws.Float64(float64(valueCount)),
			},
		},
	})

	return nil
}

// loadEtcdNode reads `r` containing JSON objects representing etcd nodes and
// loads them into server.
func loadEtcdNode(etcdClient *etcd.Client, r io.Reader) error {
	jsonReader := json.NewDecoder(r)
	for {
		node := etcd.Node{}
		if err := jsonReader.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if node.Key == "" && node.Dir {
			continue // skip root
		}

		// compute a new TTL
		ttl := 0
		if node.Expiration != nil {
			ttl := node.Expiration.Sub(time.Now()).Seconds()
			if ttl < 0 {
				// expired, skip
				continue
			}
		}

		if node.Dir {
			_, err := etcdClient.SetDir(node.Key, uint64(ttl))
			if err != nil {
				return fmt.Errorf("%s: %s", node.Key, err)
			}
		} else {
			_, err := etcdClient.Set(node.Key, node.Value, uint64(ttl))
			if err != nil {
				return fmt.Errorf("%s: %s", node.Key, err)
			}
		}
	}
	return nil
}

// restoreBackup reads the backup from S3 and applies it to the current cluster.
func restoreBackup(s *ec2cluster.Cluster, backupBucket, backupKey, dataDir string) error {
	instance, err := s.Instance()
	if err != nil {
		return err
	}
	etcdClient := etcd.NewClient([]string{fmt.Sprintf("http://%s:2379", *instance.PrivateIpAddress)})
	if success := etcdClient.SyncCluster(); !success {
		return fmt.Errorf("restore: cannot sync machines")
	}

	s3svc := s3.New(s.AwsSession)
	resp, err := s3svc.GetObject(&s3.GetObjectInput{
		Key:    &backupKey,
		Bucket: &backupBucket,
	})
	if err != nil {
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == http.StatusNotFound {
				log.Printf("restore: s3://%s%s does not exist", backupBucket, backupKey)
				return nil
			}
		}
		return fmt.Errorf("cannot fetch backup file: %s", err)
	}

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	if err := loadEtcdNode(etcdClient, gzipReader); err != nil {
		return err
	}
	log.Printf("restore: complete")
	return nil
}
