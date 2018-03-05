package aws

import (
	"fmt"
	"strconv"
	"strings"

	cfn "github.com/crewjam/go-cloudformation"
)

// EtcdAwsService is a systemd unit that wraps and runs etcd2
var EtcdAwsService = `[Unit]
Description=configures and runs etcd in AWS.

[Install]
WantedBy=multi-user.target

[Service]
Restart=always
EnvironmentFile=/etc/etcd_aws.env
ExecStart=/usr/bin/docker run --name etcd-aws \
  -p 2379:2379 -p 2380:2380 \
  -v /var/lib/etcd2:/var/lib/etcd2 \
  -e ETCD_BACKUP_BUCKET -e ETCD_BACKUP_KEY \
  --rm crewjam/etcd-aws
ExecStop=-/usr/bin/docker rm -f etcd-aws
`

// CfnSignal is a systemd unit that emits a
// cloudformation ready signal when etcd-aws.service
// is running.
var CfnSignal = `[Unit]
Description=Cloudformation Signal Ready
After=etcd-aws.service
Requires=etcd-aws.service

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
EnvironmentFile=/etc/environment
ExecStart=/bin/bash -c '\
set -ex; \
eval $(docker run crewjam/ec2cluster); \
docker run --rm crewjam/awscli cfn-signal \
  --resource MasterAutoscale --stack $TAG_AWS_CLOUDFORMATION_STACK_ID \
  --region $REGION || true; \
'
`

func indent(s, indent string) string {
	return indent + strings.Replace(s, "\n", "\n"+indent, -1)
}

// MakeMaster creates a cluster of etcd nodes.
func MakeMaster(parameters *Parameters, t *cfn.Template) error {
	t.Resources["MasterLaunchConfiguration"] = &cfn.Resource{
		DependsOn: []string{"VpcInternetGatewayAttachment", "BackupBucket"},
		Properties: cfn.AutoScalingLaunchConfiguration{
			ImageId:                  AMI,
			InstanceType:             cfn.Ref("InstanceType").String(),
			IamInstanceProfile:       cfn.Ref("MasterInstanceProfile").String(),
			KeyName:                  cfn.Ref("KeyPair").String(),
			SecurityGroups:           cfn.StringList(cfn.Ref("MasterSecurityGroup")),
			AssociatePublicIpAddress: cfn.Bool(true),
			UserData: cfn.Base64(cfn.Join("", cfn.String(""+
				"#cloud-config\n"+
				"\n"+
				"write_files:\n"+
				"  - path: \"/etc/etcd_aws.env\"\n"+
				"    permissions: \"0444\"\n"+
				"    owner: \"root\"\n"+
				"    content: \"ETCD_BACKUP_BUCKET="), cfn.Ref("BackupBucket").String(), cfn.String("\"\n"+
				"coreos:\n"+
				"  units:\n"+
				"    - name: etcd-aws.service\n"+
				"      command: start\n"+
				"      enable: true\n"+
				"      content: |\n"+
				indent(EtcdAwsService, "        ")+"\n"+
				"    - name: cfn-signal.service\n"+
				"      command: start\n"+
				"      enable: true\n"+
				"      content: |\n"+
				indent(CfnSignal, "        ")+"\n"+
				""))),
		},
	}

	scale := 3
	t.Resources["MasterAutoscale"] = &cfn.Resource{
		UpdatePolicy: &cfn.UpdatePolicy{
			AutoScalingRollingUpdate: &cfn.UpdatePolicyAutoScalingRollingUpdate{
				MinInstancesInService: cfn.Integer(int64(scale)),
				PauseTime:             cfn.String("PT5M"),
				WaitOnResourceSignals: cfn.Bool(true),
			},
		},
		CreationPolicy: &cfn.CreationPolicy{
			ResourceSignal: &cfn.CreationPolicyResourceSignal{
				Count:   cfn.Integer(int64(scale)),
				Timeout: cfn.String("PT5M"),
			},
		},
		Properties: cfn.AutoScalingAutoScalingGroup{
			AvailabilityZones: cfn.StringList(
				cfn.FindInMap("AvailablityZones", cfn.Ref("AWS::Region"), cfn.String("0")),
				cfn.FindInMap("AvailablityZones", cfn.Ref("AWS::Region"), cfn.String("1")),
				cfn.FindInMap("AvailablityZones", cfn.Ref("AWS::Region"), cfn.String("2"))),
			Cooldown:                cfn.String("300"),
			DesiredCapacity:         cfn.String(strconv.Itoa(scale)),
			MaxSize:                 cfn.String(strconv.Itoa(2 * scale)),
			MinSize:                 cfn.String(strconv.Itoa(scale)),
			HealthCheckType:         cfn.String("EC2"), // XXX should be ELB
			HealthCheckGracePeriod:  cfn.Integer(300),
			VPCZoneIdentifier:       cfn.StringList(parameters.VpcSubnets...),
			LaunchConfigurationName: cfn.Ref("MasterLaunchConfiguration").String(),
			LoadBalancerNames:       cfn.StringList(cfn.Ref("MasterLoadBalancer")),
			Tags: &cfn.AutoScalingTagsList{
				cfn.AutoScalingTags{
					Key:               cfn.String("Name"),
					Value:             cfn.String(fmt.Sprintf("%s-master", parameters.DnsName)),
					PropagateAtLaunch: cfn.Bool(true),
				},
				cfn.AutoScalingTags{
					Key:               cfn.String("application"),
					Value:             cfn.Ref("Application").String(),
					PropagateAtLaunch: cfn.Bool(true),
				},
				cfn.AutoScalingTags{
					Key:               cfn.String("cluster"),
					Value:             cfn.String(parameters.DnsName),
					PropagateAtLaunch: cfn.Bool(true),
				},
			},
		},
	}

	t.AddResource("MasterSecurityGroup", cfn.EC2SecurityGroup{
		GroupDescription: cfn.String(fmt.Sprintf("master-%s", parameters.DnsName)),
		VpcId:            cfn.Ref("Vpc").String(),
		SecurityGroupIngress: &cfn.EC2SecurityGroupRuleList{
			cfn.EC2SecurityGroupRule{
				IpProtocol: cfn.String("tcp"),
				FromPort:   cfn.Integer(22),
				ToPort:     cfn.Integer(22),
				CidrIp:     cfn.String("0.0.0.0/0"),
			},
		},
		SecurityGroupEgress: &cfn.EC2SecurityGroupRuleList{
			cfn.EC2SecurityGroupRule{
				IpProtocol: cfn.String("-1"),
				CidrIp:     cfn.String("0.0.0.0/0"),
			},
		},
	})

	// allow tcp/2379 and tcp/2380 between cluster nodes
	for _, tcpPort := range []int64{2379, 2380} {
		t.AddResource(fmt.Sprintf("MasterSecurityEtcdClusterTcp%d", tcpPort), cfn.EC2SecurityGroupIngress{
			GroupId:               cfn.GetAtt("MasterSecurityGroup", "GroupId"),
			IpProtocol:            cfn.String("tcp"),
			FromPort:              cfn.Integer(tcpPort),
			ToPort:                cfn.Integer(tcpPort),
			SourceSecurityGroupId: cfn.GetAtt("MasterSecurityGroup", "GroupId"),
		})
	}

	// The IAM role for master nodes allows introspection and reading from
	// the lifecycle notification queue.
	t.AddResource("MasterInstanceRole", cfn.IAMRole{
		AssumeRolePolicyDocument: PolicyDocument{
			Version: "2012-10-17",
			Statement: []Policy{
				Policy{
					Effect:    "Allow",
					Principal: &Principal{Service: cfn.StringList(cfn.String("ec2.amazonaws.com"))},
					Action:    cfn.StringList(cfn.String("sts:AssumeRole")),
				},
			},
		},
		Path: cfn.String("/"),
		Policies: &cfn.IAMPoliciesList{
			cfn.IAMPolicies{
				PolicyName: cfn.String("MasterInstancePolicy"),
				PolicyDocument: PolicyDocument{
					Version: "2012-10-17",
					Statement: []Policy{
						Policy{
							Sid:    "AllowIntrospection",
							Effect: "Allow",
							Action: cfn.StringList(
								cfn.String("ec2:DescribeInstances"),
								cfn.String("autoscaling:DescribeAutoScalingGroups"),
								cfn.String("autoscaling:DescribeLifecycleHooks"),
								cfn.String("autoscaling:CompleteLifecycleAction"),
							),
							Resource: cfn.StringList(cfn.String("*")),
						},
						Policy{
							Sid:    "AllowSQS",
							Effect: "Allow",
							Action: cfn.StringList(
								cfn.String("sqs:GetQueueUrl"),
								cfn.String("sqs:ReceiveMessage"),
								cfn.String("sqs:DeleteMessage"),
							),
							Resource: cfn.StringList(cfn.GetAtt("MasterAutoscaleLifecycleHookQueue", "Arn")),
						},
						Policy{
							Sid:    "AllowBackupBucket",
							Effect: "Allow",
							Action: cfn.StringList(
								/*
									cfn.String("s3:PutObject"),
									cfn.String("s3:GetObject"),
									cfn.String("s3:CreateMultipartUpload"),
									cfn.String("s3:UploadPart"),
									cfn.String("s3:AbortMultipartUpload"),
									cfn.String("s3:CompleteMultipartUpload"),
								*/
								cfn.String("s3:*"),
							),
							Resource: cfn.StringList(cfn.Join("", cfn.String("arn:aws:s3:::"), cfn.Ref("BackupBucket"), cfn.String("/*"))),
						},
						Policy{
							Sid:    "AllowPutMetric",
							Effect: "Allow",
							Action: cfn.StringList(
								cfn.String("cloudwatch:PutMetricData"),
							),
							// TODO(ross): is there any way to be more granular here?
							Resource: cfn.StringList(cfn.String("*")),
						},
					},
				},
			},
		},
	})

	t.AddResource("MasterInstanceProfile", cfn.IAMInstanceProfile{
		Path:  cfn.String("/"),
		Roles: cfn.StringList(cfn.Ref("MasterInstanceRole")),
	})

	t.AddResource("BackupBucket", &cfn.S3Bucket{})

	t.AddResource("MasterAutoscaleLifecycleHookQueue", cfn.SQSQueue{})
	t.AddResource("MasterAutoscaleLifecycleHookTerminating", cfn.AutoScalingLifecycleHook{
		AutoScalingGroupName:  cfn.Ref("MasterAutoscale").String(),
		NotificationTargetARN: cfn.GetAtt("MasterAutoscaleLifecycleHookQueue", "Arn"),
		RoleARN:               cfn.GetAtt("MasterAutoscaleLifecycleHookRole", "Arn"),
		LifecycleTransition:   cfn.String("autoscaling:EC2_INSTANCE_TERMINATING"),
		HeartbeatTimeout:      cfn.Integer(30),
	})
	t.AddResource("MasterAutoscaleLifecycleHookRole", cfn.IAMRole{
		AssumeRolePolicyDocument: PolicyDocument{
			Version: "2012-10-17",
			Statement: []Policy{
				Policy{
					Effect:    "Allow",
					Principal: &Principal{Service: cfn.StringList(cfn.String("autoscaling.amazonaws.com"))},
					Action:    cfn.StringList(cfn.String("sts:AssumeRole")),
				},
			},
		},
		Path: cfn.String("/"),
		Policies: &cfn.IAMPoliciesList{
			cfn.IAMPolicies{
				PolicyName: cfn.String("MasterInstancePolicy"),
				PolicyDocument: PolicyDocument{
					Version: "2012-10-17",
					Statement: []Policy{
						Policy{
							Sid:    "AllowSQSPublish",
							Effect: "Allow",
							Action: cfn.StringList(
								cfn.String("sqs:SendMessage"),
								cfn.String("sqs:GetQueueUrl"),
							),
							Resource: cfn.StringList(cfn.GetAtt("MasterAutoscaleLifecycleHookQueue", "Arn")),
						},
						Policy{
							Sid:    "AllowSNSPublish",
							Effect: "Allow",
							Action: cfn.StringList(
								cfn.String("sns:Publish"),
							),
							Resource: cfn.StringList(cfn.String("*")),
						},
					},
				},
			},
		},
	})

	return nil
}
