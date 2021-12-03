package steps

import (
	"github.com/cucumber/godog"
	"github.com/maxcnunes/httpfake"
)

func NewSearchFeature() *SearchFeature {
	f := &SearchFeature{
		FakeSearchApi: httpfake.New(),
	}

	return f
}

type SearchFeature struct {
	ErrorFeature
	FakeSearchApi *httpfake.HTTPFake
}

func (f *SearchFeature) Reset() {
	f.FakeSearchApi.Reset()
}

func (f *SearchFeature) Close() {
	f.FakeSearchApi.Close()
}

func (f *SearchFeature) searchApiCreatesANewIndex() error {
	f.FakeSearchApi.NewHandler().Post("/search").Reply(201).BodyString(`{ "IndexName": "ons1638363874110115"}`)
	return nil
}

func (f *SearchFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^the search api is working correctly$`, f.searchApiCreatesANewIndex)
}
