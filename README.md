# Rancher HA on AWS
This project is under active development and is continually being improved. If you find any bugs or issues, please email help@solodev.com.

## Overview

## Prerequisites
The following is a list of prerequisites need to launch a stack. Please note that each of the following must be configured within the region you intend to launch the stack. If the following items are already created, you can skip to deploying.

* [VPC](https://s3.amazonaws.com/techcto-datacenter/aws/corp-vpc.yaml)
* [EC2 Key Pair](https://console.aws.amazon.com/ec2/v2/home?#KeyPairs:sort=keyName)
* [SSL/TLS Certificates](https://console.aws.amazon.com/acm/home?#/privatewizard/)

## Steps to Run
To launch the entire stack and deploy on AWS, click on one of the ***Launch Stack*** links below or download the Master template and launch it locally.

You can launch this CloudFormation stack, using your account, in the following AWS Regions:

| AWS Region Code | Name | Launch |
| --- | --- | --- 
| us-east-1 |US East (N. Virginia)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| us-east-2 |US East (Ohio)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| us-west-1 |US West (N. California)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| us-west-2 |US West (Oregon)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| eu-west-1 |EU (Ireland)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| eu-west-2 |EU (London)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| eu-central-1 |EU (Frankfurt)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |
| ca-central-1 |Canada (Central)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=ca-central-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/techcto-datacenter/aws/rancher-ha-cluster.yaml) |

This will open the Create a New Stack configuration with the template preconfigured:

![Create a New Stack Configuration](images/stack-template.jpg)

Click "Next" to proceed to specifying the necessary stack parameters.

## Parameters

## FAQs
