package steps

import (
	"github.com/maxcnunes/httpfake"
)

func NewFakeSearchAPI() *FakeAPI {
	f := &FakeAPI{
		fakeHTTP: httpfake.New(),
	}

	return f
}

type FakeAPI struct {
	fakeHTTP *httpfake.HTTPFake
}

func (f *FakeAPI) Restart() {
	f.fakeHTTP = httpfake.New()
}

func (f *FakeAPI) Reset() {
	f.fakeHTTP.Reset()
}

func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}
