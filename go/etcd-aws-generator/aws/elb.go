package aws

import (
	"fmt"

	cfn "github.com/crewjam/go-cloudformation"
)

// MakeMasterLoadBalancer creates a load balancer the connects to each
// master node.
func MakeMasterLoadBalancer(parameters *Parameters, t *cfn.Template) error {
	t.AddResource("MasterLoadBalancer", cfn.ElasticLoadBalancingLoadBalancer{
		Scheme:  cfn.String("internal"),
		Subnets: cfn.StringList(parameters.VpcSubnets...),
		Listeners: &cfn.ElasticLoadBalancingListenerList{
			cfn.ElasticLoadBalancingListener{
				LoadBalancerPort: cfn.String("2379"),
				InstancePort:     cfn.String("2379"),
				Protocol:         cfn.String("HTTP"),
			},
		},
		HealthCheck: &cfn.ElasticLoadBalancingHealthCheck{
			Target:             cfn.String("HTTP:2379/health"),
			HealthyThreshold:   cfn.String("2"),
			UnhealthyThreshold: cfn.String("10"),
			Interval:           cfn.String("10"),
			Timeout:            cfn.String("5"),
		},
		SecurityGroups: cfn.StringList(
			cfn.Ref("MasterLoadBalancerSecurityGroup")),
	})

	// allow tcp/2379 from the load balancer to the nodes
	for _, tcpPort := range []int64{2379} {
		t.AddResource(fmt.Sprintf("MasterSecurityIngressFromLoadBalancerTcp%d", tcpPort), cfn.EC2SecurityGroupIngress{
			GroupId:               cfn.GetAtt("MasterSecurityGroup", "GroupId"),
			IpProtocol:            cfn.String("tcp"),
			FromPort:              cfn.Integer(tcpPort),
			ToPort:                cfn.Integer(tcpPort),
			SourceSecurityGroupId: cfn.GetAtt("MasterLoadBalancerSecurityGroup", "GroupId"),
		})
	}

	// allow tcp/2379 in to the load balancer
	t.AddResource("MasterLoadBalancerSecurityGroup", cfn.EC2SecurityGroup{
		GroupDescription: cfn.String(fmt.Sprintf("LoadBalancer-%s", parameters.DnsName)),
		VpcId:            cfn.Ref("Vpc").String(),
		SecurityGroupIngress: &cfn.EC2SecurityGroupRuleList{
			cfn.EC2SecurityGroupRule{
				IpProtocol: cfn.String("tcp"),
				FromPort:   cfn.Integer(2379),
				ToPort:     cfn.Integer(2379),
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

	return nil
}
