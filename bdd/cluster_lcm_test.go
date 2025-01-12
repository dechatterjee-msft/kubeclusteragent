package main

import (
	"github.com/cucumber/godog"
	"kubeclusteragent/stf/features"
	"testing"
)

func TestCreateKubeadmCluster(t *testing.T) {
	suite := godog.TestSuite{
		TestSuiteInitializer: func(context *godog.TestSuiteContext) {
			context.BeforeSuite(func() {})
			context.AfterSuite(func() {})
		},
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^Lcm of single node kubeadm cluster$`, features.LcmOfSingleNodeKubeadmCluster(t))
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
