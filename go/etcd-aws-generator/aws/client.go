package aws

import (
	"fmt"

	cfn "github.com/crewjam/go-cloudformation"
)

// MakeClientSecurityGroup creates a security group for clients of the etcd cluster
// and adds permissions that allow traffic into the cluster from there.
func MakeClientSecurityGroup(parameters *Parameters, t *cfn.Template) error {
	t.AddResource("ClientSecurityGroup", cfn.EC2SecurityGroup{
		GroupDescription: cfn.String(fmt.Sprintf("clients of master-%s", parameters.DnsName)),
		VpcId:            cfn.Ref("Vpc").String(),
	})
	t.AddResource("MasterSecurityGroupTcp2379FromClient", cfn.EC2SecurityGroupIngress{
		GroupId:               cfn.GetAtt("MasterSecurityGroup", "GroupId"),
		IpProtocol:            cfn.String("tcp"),
		FromPort:              cfn.Integer(2379),
		ToPort:                cfn.Integer(2379),
		SourceSecurityGroupId: cfn.GetAtt("ClientSecurityGroup", "GroupId"),
	})
	t.AddResource("MasterLoadBalancerSecurityGroupTcp2379FromClient", cfn.EC2SecurityGroupIngress{
		GroupId:               cfn.GetAtt("MasterLoadBalancerSecurityGroup", "GroupId"),
		IpProtocol:            cfn.String("tcp"),
		FromPort:              cfn.Integer(2379),
		ToPort:                cfn.Integer(2379),
		SourceSecurityGroupId: cfn.GetAtt("ClientSecurityGroup", "GroupId"),
	})
	return nil
}
