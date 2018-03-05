package aws

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	cfn "github.com/crewjam/go-cloudformation"
)

type Parameters struct {
	DnsName    string
	VpcSubnets []cfn.Stringable
}

var Regions = []string{
	"us-east-1",      // US East (N. Virginia)
	"us-west-2",      // US West (Oregon)
	"us-west-1",      // US West (N. California)
	"eu-west-1",      // EU (Ireland)
	"eu-central-1",   // EU (Frankfurt)
	"ap-southeast-1", // Asia Pacific (Singapore)
	"ap-northeast-1", // Asia Pacific (Tokyo)
	"ap-southeast-2", // Asia Pacific (Sydney)
	"ap-northeast-2", // Asia Pacific (Seoul)
	"sa-east-1",      // South America (Sao Paulo)
}

func MakeAvailabilityZonesMap(awsSession *session.Session, t *cfn.Template) error {
	rv := cfn.Mapping{}
	for _, region := range Regions {
		s := *awsSession
		s.Config = s.Config.Copy()
		s.Config.Region = aws.String(region)

		azinfo, err := ec2.New(&s).DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
		if err != nil {
			return err
		}

		goodZones := []string{}
		badZones := []string{}
		for _, availabilityZone := range azinfo.AvailabilityZones {
			if *availabilityZone.State == ec2.AvailabilityZoneStateAvailable {
				goodZones = append(goodZones, *availabilityZone.ZoneName)
			} else {
				badZones = append(badZones, *availabilityZone.ZoneName)
			}
		}
		zones := append(goodZones, badZones...)

		rv[region] = map[string]string{}
		for i, zone := range zones {
			rv[region][strconv.Itoa(i)] = zone
		}
	}

	t.Mappings["AvailablityZones"] = &rv
	return nil
}

// MakeTemplate returns a Cloudformation template for the project infrastructure
func MakeTemplate(awsSession *session.Session, parameters *Parameters) (*cfn.Template, error) {
	t := cfn.NewTemplate()
	t.Parameters = map[string]*cfn.Parameter{
		"MasterScale": &cfn.Parameter{
			Description: "Number of master instances",
			Type:        "String",
			Default:     "3",
		},
		"KeyPair": &cfn.Parameter{
			Description: "Number of master instances",
			Type:        "AWS::EC2::KeyPair::KeyName",
			Default:     "",
		},
		"Application": &cfn.Parameter{
			Description: "Number of master instances",
			Type:        "String",
			Default:     "",
		},
		"InstanceType": &cfn.Parameter{
			Description: "The type of instance",
			Type:        "String",
			Default:     "t1.micro",
		},
	}
	if err := MakeAvailabilityZonesMap(awsSession, t); err != nil {
		return nil, err
	}

	if err := MakeMapping(parameters, t); err != nil {
		return nil, err
	}
	if err := MakeVPC(parameters, t); err != nil {
		return nil, err
	}

	if err := MakeMaster(parameters, t); err != nil {
		return nil, err
	}

	if err := MakeMasterLoadBalancer(parameters, t); err != nil {
		return nil, err
	}

	if err := MakeHealthCheck(parameters, t); err != nil {
		return nil, err
	}

	return t, nil
}
