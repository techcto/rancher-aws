package aws

import (
	"fmt"
	"strconv"

	cfn "github.com/crewjam/go-cloudformation"
)

// MakeVPC creates a VPC with a subnet for each of up to four availability zones.
func MakeVPC(parameters *Parameters, t *cfn.Template) error {
	t.AddResource("Vpc", cfn.EC2VPC{
		CidrBlock:          cfn.String("10.0.0.0/16"),
		InstanceTenancy:    cfn.String("default"),
		EnableDnsSupport:   cfn.Bool(true),
		EnableDnsHostnames: cfn.Bool(true),
		Tags: []cfn.ResourceTag{
			cfn.ResourceTag{Key: cfn.String("application"), Value: cfn.Ref("Application").String()},
			cfn.ResourceTag{Key: cfn.String("Name"), Value: cfn.String(parameters.DnsName)},
		},
	})

	cidrBlocks := []string{
		"10.0.0.0/18",
		"10.0.64.0/18",
		"10.0.128.0/18",
		"10.0.192.0/18",
	}

	for index := 0; index < 3; index++ {
		t.AddResource(fmt.Sprintf("VpcSubnet%d", index), cfn.EC2Subnet{
			CidrBlock:        cfn.String(cidrBlocks[index]),
			AvailabilityZone: cfn.FindInMap("AvailablityZones", cfn.Ref("AWS::Region"), cfn.String(strconv.Itoa(index))),
			VpcId:            cfn.Ref("Vpc"),
			Tags: []cfn.ResourceTag{
				cfn.ResourceTag{Key: cfn.String("application"), Value: cfn.Ref("Application").String()},
				cfn.ResourceTag{Key: cfn.String("Name"), Value: cfn.String(fmt.Sprintf("%s-%d", parameters.DnsName, index))},
			},
		})
		t.AddResource(fmt.Sprintf("VpcSubnetRouteTableAssociation%d", index), cfn.EC2SubnetRouteTableAssociation{
			SubnetId:     cfn.Ref(fmt.Sprintf("VpcSubnet%d", index)).String(),
			RouteTableId: cfn.Ref("VpcRouteTable").String(),
		})
		t.AddResource(fmt.Sprintf("VpcSubnetAcl%d", index), cfn.EC2SubnetNetworkAclAssociation{
			NetworkAclId: cfn.Ref("VpcNetworkAcl").String(),
			SubnetId:     cfn.Ref(fmt.Sprintf("VpcSubnet%d", index)).String(),
		})

		parameters.VpcSubnets = append(parameters.VpcSubnets,
			cfn.Ref(fmt.Sprintf("VpcSubnet%d", index)).String())
	}

	t.AddResource("VpcInternetGateway", cfn.EC2InternetGateway{
		Tags: []cfn.ResourceTag{
			cfn.ResourceTag{Key: cfn.String("application"), Value: cfn.Ref("Application").String()},
			cfn.ResourceTag{Key: cfn.String("Name"), Value: cfn.String(fmt.Sprintf("InternetGateway-%s", parameters.DnsName))},
		},
	})
	t.AddResource("VpcInternetGatewayAttachment", cfn.EC2VPCGatewayAttachment{
		VpcId:             cfn.Ref("Vpc").String(),
		InternetGatewayId: cfn.Ref("VpcInternetGateway").String(),
	})
	t.Resources["VpcDefaultRoute"] = &cfn.Resource{
		DependsOn: []string{"VpcInternetGatewayAttachment"},
		Properties: cfn.EC2Route{
			DestinationCidrBlock: cfn.String("0.0.0.0/0"),
			RouteTableId:         cfn.Ref("VpcRouteTable").String(),
			GatewayId:            cfn.Ref("VpcInternetGateway").String(),
		},
	}
	t.AddResource("VpcRouteTable", cfn.EC2RouteTable{
		VpcId: cfn.Ref("Vpc").String(),
	})
	t.AddResource("VpcNetworkAcl", cfn.EC2NetworkAcl{
		VpcId: cfn.Ref("Vpc").String(),
	})
	t.AddResource("VpcNetworkAclAllowAllEgress", cfn.EC2NetworkAclEntry{
		CidrBlock:    cfn.String("0.0.0.0/0"),
		Egress:       cfn.Bool(true),
		Protocol:     cfn.Integer(-1),
		RuleAction:   cfn.String("allow"),
		RuleNumber:   cfn.Integer(100),
		NetworkAclId: cfn.Ref("VpcNetworkAcl").String(),
	})
	t.AddResource("VpcNetworkAclAllowAllIngress", cfn.EC2NetworkAclEntry{
		CidrBlock:    cfn.String("0.0.0.0/0"),
		Protocol:     cfn.Integer(-1),
		RuleAction:   cfn.String("allow"),
		RuleNumber:   cfn.Integer(100),
		NetworkAclId: cfn.Ref("VpcNetworkAcl").String(),
	})
	t.AddResource("VpcDhcpOptions", cfn.EC2DHCPOptions{
		DomainNameServers: cfn.StringList(cfn.String("AmazonProvidedDNS")),
	})
	t.AddResource("VpcDhpcOptionsAssociation", cfn.EC2VPCDHCPOptionsAssociation{
		VpcId:         cfn.Ref("Vpc").String(),
		DhcpOptionsId: cfn.Ref("VpcDhcpOptions").String(),
	})
	return nil
}
