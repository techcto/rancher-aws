# Rancher HA on AWS
This project is under active development and is continually being improved. If you find any bugs or issues, please email help@solodev.com.

## Overview

## Prerequisites
The following is a list of prerequisites need to launch a stack. Please note that each of the following must be configured within the region you intend to launch the stack. If the following items are already created, you can skip to deploying.

* [VPC](https://s3.amazonaws.com/rancher-aws/aws/corp-vpc.yaml)
* [EC2 Key Pair](https://console.aws.amazon.com/ec2/v2/home?#KeyPairs:sort=keyName)
* [SSL/TLS Certificates](https://console.aws.amazon.com/acm/home?#/privatewizard/)

## Steps to Run
To launch the entire stack and deploy on AWS, click on one of the ***Launch Stack*** links below or download the Master template and launch it locally.

You can launch this CloudFormation stack, using your account, in the following AWS Regions:

| AWS Region Code | Name | Launch |
| --- | --- | --- 
| us-east-1 |US East (N. Virginia)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| us-east-2 |US East (Ohio)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| us-west-1 |US West (N. California)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| us-west-2 |US West (Oregon)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| eu-west-1 |EU (Ireland)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| eu-west-2 |EU (London)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-2#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| eu-central-1 |EU (Frankfurt)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |
| ca-central-1 |Canada (Central)| [![cloudformation-launch-stack](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=ca-central-1#/stacks/new?stackName=rancher&templateURL=https://s3.amazonaws.com/rancher-aws/aws/rancher-ha-cluster.yaml) |

This will open the Create a New Stack configuration with the template preconfigured:

![Create a New Stack Configuration](images/stack-template.jpg)

Click "Next" to proceed to specifying the necessary stack parameters.

## Parameters

![Create a New Stack Configuration](images/stack-parameters.jpg)

| Parameter | Description |
| --- | --- |
| CertificateArn | (Prerequisite) The SSL certificate for AWS ALB HTTPS listener. [Create your certificate](https://console.aws.amazon.com/acm/home?#/privatewizard/) beforehand if you have not done so. |
| FQDN | The fully qualified URL for using the app. DNS of FQDN must be pointed to the CNAME of ALB. |
| InstanceType | The EC2 instance type you would be to use. |
| InstanceUser | A random user that is generated automatically so that you don't have to specify your main ec2-user and master key. A key is created just for use with the stack. |
| KeyName | (Prerequisite) The name of an existing EC@ KeyPair to enable SSH access to the instances. [Create a KeyPair](https://console.aws.amazon.com/ec2/v2/home?#KeyPairs:sort=keyName) beforehand if you have not done so.|
| Subnets | Choose which subnets this ECS cluster should be deployed to.|
| VPC | (Prerequisite) Choose which VPC the Application Load Balancer should be deployed to. Create the VPC beforehand using [this configuration](https://s3.amazonaws.com/rancher-aws/aws/corp-vpc.yaml) if you have not done so. |

## FAQs
