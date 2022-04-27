package steps

import (
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewSearchAPIFeature() *SearchAPIFeature {
	f := &SearchAPIFeature{
		FakeSearchAPI: httpfake.New(),
	}

	return f
}

type SearchAPIFeature struct {
	ErrorFeature
	FakeSearchAPI *httpfake.HTTPFake
}

func (f *SearchAPIFeature) Restart() {
	f.FakeSearchAPI = httpfake.New()
}

func (f *SearchAPIFeature) Reset() {
	f.FakeSearchAPI.Reset()
}

func (f *SearchAPIFeature) Close() {
	f.FakeSearchAPI.Close()
}

func (f *SearchAPIFeature) successfulSearchAPIResponse() error {
	f.FakeSearchAPI.NewHandler().Post("/search").Reply(201).BodyString(`{ "IndexName": "ons1638363874110115"}`)
	return nil
}

func (f *SearchAPIFeature) unsuccessfulSearchAPIResponse() error {
	f.FakeSearchAPI.NewHandler().Post("/search").Reply(500).BodyString(`internal server error`)
	return nil
}

func (f *SearchAPIFeature) restartFakeSearchAPI() error {
	f.Restart()
	return nil
}

// RegisterSteps defines the steps within a specific SearchFeature cucumber test.
func (f *SearchAPIFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the search api is working correctly$`, f.successfulSearchAPIResponse)
	ctx.Step(`^the search api is not working correctly$`, f.unsuccessfulSearchAPIResponse)
	ctx.Step(`^restart the search api$`, f.restartFakeSearchAPI)
}
