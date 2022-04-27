package steps

import (
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
