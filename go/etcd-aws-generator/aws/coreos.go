package aws

import (
	"encoding/json"
	"net/http"

	cfn "github.com/crewjam/go-cloudformation"
)

var Arch = cfn.FindInMap("VirtType", cfn.Ref("InstanceType"), cfn.String("64"))

var AMI = cfn.FindInMap("CoreOSAMI", cfn.Ref("AWS::Region").String(), Arch)

func MakeMapping(parameters *Parameters, t *cfn.Template) error {
	t.Mappings["VirtType"] = &cfn.Mapping{
		"t1.micro":    map[string]string{"64": "pv"},
		"t2.micro":    map[string]string{"64": "hvm"},
		"t2.small":    map[string]string{"64": "hvm"},
		"t2.medium":   map[string]string{"64": "hvm"},
		"m1.small":    map[string]string{"64": "pv"},
		"m1.medium":   map[string]string{"64": "pv"},
		"m1.large":    map[string]string{"64": "pv"},
		"m1.xlarge":   map[string]string{"64": "pv"},
		"m2.xlarge":   map[string]string{"64": "pv"},
		"m2.2xlarge":  map[string]string{"64": "pv"},
		"m2.4xlarge":  map[string]string{"64": "pv"},
		"m3.medium":   map[string]string{"64": "hvm"},
		"m3.large":    map[string]string{"64": "hvm"},
		"m3.xlarge":   map[string]string{"64": "hvm"},
		"m3.2xlarge":  map[string]string{"64": "hvm"},
		"c1.medium":   map[string]string{"64": "pv"},
		"c1.xlarge":   map[string]string{"64": "pv"},
		"c3.large":    map[string]string{"64": "hvm"},
		"c3.xlarge":   map[string]string{"64": "hvm"},
		"c3.2xlarge":  map[string]string{"64": "hvm"},
		"c3.4xlarge":  map[string]string{"64": "hvm"},
		"c3.8xlarge":  map[string]string{"64": "hvm"},
		"c4.large":    map[string]string{"64": "hvm"},
		"c4.xlarge":   map[string]string{"64": "hvm"},
		"c4.2xlarge":  map[string]string{"64": "hvm"},
		"c4.4xlarge":  map[string]string{"64": "hvm"},
		"c4.8xlarge":  map[string]string{"64": "hvm"},
		"g2.2xlarge":  map[string]string{"64": "HVMG2"},
		"r3.large":    map[string]string{"64": "hvm"},
		"r3.xlarge":   map[string]string{"64": "hvm"},
		"r3.2xlarge":  map[string]string{"64": "hvm"},
		"r3.4xlarge":  map[string]string{"64": "hvm"},
		"r3.8xlarge":  map[string]string{"64": "hvm"},
		"i2.xlarge":   map[string]string{"64": "hvm"},
		"i2.2xlarge":  map[string]string{"64": "hvm"},
		"i2.4xlarge":  map[string]string{"64": "hvm"},
		"i2.8xlarge":  map[string]string{"64": "hvm"},
		"hi1.4xlarge": map[string]string{"64": "hvm"},
		"hs1.8xlarge": map[string]string{"64": "hvm"},
		"cr1.8xlarge": map[string]string{"64": "hvm"},
		"cc2.8xlarge": map[string]string{"64": "hvm"},
	}

	resp, err := http.Get("https://coreos.com/dist/aws/aws-stable.json")
	if err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	rv := cfn.Mapping{}
	for regionName, regionData := range data {
		if regionName == "release_info" {
			continue
		}
		rv[regionName] = map[string]string{}
		rd := regionData.(map[string]interface{})
		for arch, ami := range rd {
			rv[regionName][arch] = ami.(string)
		}
	}
	t.Mappings["CoreOSAMI"] = &rv

	return nil
}
