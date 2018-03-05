package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/crewjam/awsregion"
	etcdaws "github.com/crewjam/etcd-aws/aws"
	"github.com/crewjam/go-cloudformation/deploycfn"
)

func Main() error {
	application := flag.String("app", "", "The name of the application")
	dnsName := flag.String("dns-name", "myapp.example.com", "The top level DNS name of the cluster.")
	keyPair := flag.String("key-pair", "", "the name of the EC2 SSH keypair to use")
	instanceType := flag.String("instance-type", "t1.micro",
		"The type of instance to use for master instances.")
	stackName := flag.String("stack", "", "The name of the CloudFormation stack. The default is derived from the DNS name.")
	doDryRun := flag.Bool("dry-run", false, "just print the cloudformation template, don't deploy it.")
	flag.Parse()

	awsSession := session.New()
	if region := os.Getenv("AWS_REGION"); region != "" {
		awsSession.Config.WithRegion(region)
	}
	awsregion.GuessRegion(awsSession.Config)

	parameters := map[string]string{}
	if *keyPair != "" {
		parameters["KeyPair"] = *keyPair
	}
	if *application != "" {
		parameters["Application"] = *application
	}
	if *instanceType != "" {
		parameters["InstanceType"] = *instanceType
	}

	templateParameters := etcdaws.Parameters{
		DnsName: *dnsName,
	}
	template, err := etcdaws.MakeTemplate(awsSession, &templateParameters)
	if err != nil {
		return err
	}

	if *doDryRun {
		templateBuf, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return err
		}
		os.Stdout.Write(templateBuf)
		return nil
	}

	if *stackName == "" {
		*stackName = regexp.MustCompile("[^A-Za-z0-9\\-]").ReplaceAllString(
			templateParameters.DnsName, "")
	}

	if err := deploycfn.Deploy(deploycfn.DeployInput{
		Session:    awsSession,
		StackName:  *stackName,
		Template:   template,
		Parameters: parameters,
	}); err != nil {
		return fmt.Errorf("deploy: %s", err)
	}
	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}
