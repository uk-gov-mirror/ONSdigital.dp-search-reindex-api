package main

import (
	"flag"
	"fmt"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-search-reindex-api/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"os"
	"testing"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := componenttest.NewAuthorizationFeature()
	jobsFeature, err := steps.NewJobsFeature()
	if err != nil {
		os.Exit(1)
	}
	apiFeature := jobsFeature.InitAPIFeature()
	if err != nil {
		os.Exit(1)
	}

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		jobsFeature.Reset()
		authorizationFeature.Reset()
	})
	ctx.AfterScenario(func(*godog.Scenario, error) {
		jobsFeature.Close()
		authorizationFeature.Close()
	})
	jobsFeature.RegisterSteps(ctx)
	apiFeature.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}
func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})
	ctx.AfterSuite(func() {
	})
}
func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0
		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}
		f := &ComponentTest{}
		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()
		fmt.Println("=================================")
		fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")
		if status != 0 {
			t.FailNow()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
