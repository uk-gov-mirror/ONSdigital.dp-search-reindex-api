package main

import (
	"flag"
	"fmt"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-search-reindex-api/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"os"
	"testing"
)

// Mongo version here is overridden in the pipeline by the URL provided in the component.sh
const MongoVersion = "4.0.23"
const DatabaseName = "testing"

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	MongoFeature *componentTest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	authorizationFeature := componentTest.NewAuthorizationFeature()
	jobsFeature, err := steps.NewJobsFeature(f.MongoFeature)
	if err != nil {
		os.Exit(1)
	}
	apiFeature := jobsFeature.InitAPIFeature()

	ctx.BeforeScenario(func(*godog.Scenario) {
		apiFeature.Reset()
		jobsFeature.Reset(false)
		authorizationFeature.Reset()
	})
	ctx.AfterScenario(func(*godog.Scenario, error) {
		err := jobsFeature.Close()
		if err != nil {
			os.Exit(1)
		}
		authorizationFeature.Close()
	})
	jobsFeature.RegisterSteps(ctx)
	apiFeature.RegisterSteps(ctx)
	authorizationFeature.RegisterSteps(ctx)
}
func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		f.MongoFeature = componentTest.NewMongoFeature(componentTest.MongoOptions{MongoVersion: MongoVersion, DatabaseName: DatabaseName})
	})
	ctx.AfterSuite(func() {
		err := f.MongoFeature.Close()
		if err != nil {
			os.Exit(1)
		}
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
