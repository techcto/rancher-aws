package aws

import (
	"fmt"

	cfn "github.com/crewjam/go-cloudformation"
)

// MakeHealthCheck creates CloudWatch alarms that monitor the health of the cluster.
// Notifications go to the HealthTopic. You should subscribe manually to this topic
// if you care about the health of the cluster.
func MakeHealthCheck(parameters *Parameters, t *cfn.Template) error {
	t.AddResource("HealthTopic", cfn.SNSTopic{
		DisplayName: cfn.String(fmt.Sprintf("cloudwatch notifications for %s", parameters.DnsName)),
	})

	t.AddResource("MasterLoadBalancerHealthAlarm", cfn.CloudWatchAlarm{
		ActionsEnabled:     cfn.Bool(true),
		AlarmActions:       cfn.StringList(cfn.Ref("HealthTopic").String()),
		OKActions:          cfn.StringList(cfn.Ref("HealthTopic").String()),
		AlarmDescription:   cfn.String("master instance health"),
		AlarmName:          cfn.String("MasterInstanceHealth"),
		ComparisonOperator: cfn.String("LessThanThreshold"),
		EvaluationPeriods:  cfn.String("1"),
		Dimensions: &cfn.CloudWatchMetricDimensionList{
			cfn.CloudWatchMetricDimension{
				Name:  cfn.String("LoadBalancerName"),
				Value: cfn.Ref("MasterLoadBalancer").String(),
			},
		},
		MetricName: cfn.String("HealthyHostCount"),
		Namespace:  cfn.String("AWS/ELB"),
		Period:     cfn.String("60"),
		Statistic:  cfn.String("Minimum"),
		Unit:       cfn.String("Count"),

		// Note: for scale=3 we should have no fewer than 1 healthy instance
		// *PER AVAILABILITY ZONE*. This is confusing, I know.
		Threshold: cfn.String("1"),
	})
	t.Resources["MasterLoadBalancerHealthAlarm"].DependsOn = []string{"HealthTopic"}

	// this alarm is triggered (mostly) by the requirement that data be present.
	// if it isn't for 300 seconds, then the backups are failing and the check goes
	// into the INSUFFICIENT_DATA state and we are alerted.
	t.AddResource("MasterBackupHealthAlarm", cfn.CloudWatchAlarm{
		ActionsEnabled:          cfn.Bool(true),
		AlarmActions:            cfn.StringList(cfn.Ref("HealthTopic").String()),
		InsufficientDataActions: cfn.StringList(cfn.Ref("HealthTopic").String()),
		OKActions:               cfn.StringList(cfn.Ref("HealthTopic").String()),
		AlarmDescription:        cfn.String("key backup count"),
		AlarmName:               cfn.String("MasterBackupKeyCount"),
		ComparisonOperator:      cfn.String("LessThanThreshold"),
		EvaluationPeriods:       cfn.String("1"),
		Dimensions: &cfn.CloudWatchMetricDimensionList{
			cfn.CloudWatchMetricDimension{
				Name:  cfn.String("AutoScalingGroupName"),
				Value: cfn.Ref("MasterAutoscale").String(),
			},
		},
		MetricName: cfn.String("BackupKeyCount"),
		Namespace:  cfn.String("Local/etcd"),
		Period:     cfn.String("300"),
		Statistic:  cfn.String("Minimum"),
		Unit:       cfn.String("Count"),
		Threshold:  cfn.String("1"),
	})

	return nil
}
