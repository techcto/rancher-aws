
# etcd-aws

This repository contains tools for building a robust etcd cluster in AWS.

It uses CloudFormation to establish a three node autoscaling group of etcd instances. In case of the failure of a single node, the cluster remains available and the replacement nodes are integrated automatically into the cluster. Each node in the cluster can be replaced by a new node, one at a time, and the cluster remains available. In the event of failure of all nodes simultaneously, the cluster recovers from the backup stored in S3 without intervention.

Please see [this blog post](https://crewjam.com/etcd-aws) for more on how this little utility came to be.

Invoking the `etcd-aws` program will configure and launch etcd based on the 
current autoscaling group:

    etcd-aws

It is also available as a Docker container:

    /usr/bin/docker run --name etcd-aws \
      -p 2379:2379 -p 2380:2380 \
      -v /var/lib/etcd2:/var/lib/etcd2 \
      -e ETCD_BACKUP_BUCKET=my-etcd-backups \
      --rm crewjam/etcd-aws

# CloudFormation

The program `etcd-aws-cfn` generates and deploys a CloudFormation template:

    go install ./...
    etcd-aws-cfn -key-pair my-key

You can also generate the CloudFormation template and deploy it yourself:

    etcd-aws-cfn -key-pair my-key -dry-run > etcd.template

The template consists of:

- A VPC containing three subnets across three availability zones.
- An autoscaling group of CoreOS instances running etcd with an initial size of 3.
- An internal load balancer that routes etcd client requests to the autoscaling group.
- A lifecycle hook that monitors the autoscaling group and sends termination events to an SQS queue.
- An S3 bucket that stores the backup.
- CloudWatch alarms that monitor the health of the cluster and that the backup is happening.

# Cluster Discovery

The program `etcd-aws` discovers other cluster members by looking for EC2 instances that are part of the same autoscaling group. It invokes `etcd` with appropriate configuration settings based on the result of cluster discovery.

When adding nodes to an existing cluster, `etcd-aws` automatically registers the node before it is launched.

The program monitors an AWS AutoScaling Lifecycle Hook to detect when nodes are terminated and removes them from the cluster. This is important because the terminated nodes no longer count against the `etcd` quorum calculation.

# Backup

Periodically, `etcd-aws` writes a file to S3 containing the value of all the keys in the `etcd` database.

When creating the first node of a cluster, `etcd-aws` checks for an existing backup and automatically restores it. In this way, an `etcd-aws` cluster can recover from failure of all nodes in the cluster.

# Load Balancer

The CloudFormation template creates a load balancer which can be used by etcd clients to discover cluster members. Etcd clients tend to be cluster aware -- they discover the cluster members on initial contact. You can configure an etcd client to connect to the load balancer, which will provide the initial node list, and then the client will connect directly to the current nodes in the cluster. This avoids the need for clients to maintain and update a list of etcd nodes.
