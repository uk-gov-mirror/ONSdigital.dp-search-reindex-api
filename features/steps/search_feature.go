package steps

import (
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewSearchFeature() *SearchFeature {
	f := &SearchFeature{
		FakeSearchAPI: httpfake.New(),
	}

	return f
}

type SearchFeature struct {
	ErrorFeature
	FakeSearchAPI *httpfake.HTTPFake
}

func (f *SearchFeature) Reset() {
	f.FakeSearchAPI.Reset()
}

func (f *SearchFeature) Close() {
	f.FakeSearchAPI.Close()
}

func (f *SearchFeature) theSearchAPIIsWorkingCorrectly() error {
	f.FakeSearchAPI.NewHandler().Post("/search").Reply(201).BodyString(`{ "IndexName": "ons1638363874110115"}`)
	return nil
}

func (f *SearchFeature) theSearchReindexAPIIsNotWorkingCorrectlyBecauseItIsPointingToTheOldVersionOfES() error {
	f.FakeSearchAPI.NewHandler().Post("/search").Reply(500).BodyString(`internal server error`)
	return nil
}

// RegisterSteps defines the steps within a specific SearchFeature cucumber test.
func (f *SearchFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the search api is working correctly$`, f.theSearchAPIIsWorkingCorrectly)
	ctx.Step(`^the search reindex api is not working correctly because it is pointing to the old version of ES$`, f.theSearchReindexAPIIsNotWorkingCorrectlyBecauseItIsPointingToTheOldVersionOfES)
}
