package aws

import cfn "github.com/crewjam/go-cloudformation"

type PolicyDocument struct {
	Version   string `json:",omitempty"`
	Statement []Policy
}

type Policy struct {
	Sid            string              `json:",omitempty"`
	Effect         string              `json:",omitempty"`
	Principal      *Principal          `json:",omitempty"`
	Action         *cfn.StringListExpr `json:",omitempty"`
	Resource       *cfn.StringListExpr `json:",omitempty"`
	ConditionBlock interface{}         `json:",omitempty"`
}

type Principal struct {
	Service *cfn.StringListExpr `json:",omitempty"`
}
