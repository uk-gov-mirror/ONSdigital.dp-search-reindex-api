package event_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-reindex-api/event"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProducer(t *testing.T) {
	fakeKafkaProducer := &FakeKafkaProducer{}
	reindexRequestedEvent := models.ReindexRequested{JobID: "job id", SearchIndex: "search index", TraceID: "trace id"}
	checkState := healthcheck.NewCheckState("test")
	fakeMarshaller := &FakeMarshaller{}

	Convey("Given the producer is initialized without a marshaller", t, func() {
		sut := event.ReindexRequestedProducer{Marshaller: nil, Producer: fakeKafkaProducer}

		Convey("When ProduceReindexRequested is called then an error should be returned", func() {
			err := sut.ProduceReindexRequested(context.Background(), reindexRequestedEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "marshaller is not provided")
		})

		Convey("When Close is called then an error should be called", func() {
			err := sut.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "marshaller is not provided")
		})

		Convey("When Checker is called then an error should be called", func() {
			err := sut.Checker(context.Background(), checkState)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "marshaller is not provided")
		})
	})

	Convey("Given the producer is initialized without a kafka provider", t, func() {
		producer := event.ReindexRequestedProducer{Marshaller: fakeMarshaller, Producer: nil}
		const expectedErrorMessage = "producer is not provided"

		Convey("When ProduceReindexRequested is called then an error should be returned", func() {
			err := producer.ProduceReindexRequested(context.Background(), reindexRequestedEvent)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, expectedErrorMessage)
		})

		Convey("When Close is called then an error should be called", func() {
			err := producer.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, expectedErrorMessage)
		})

		Convey("When Checker is called then an error should be called", func() {
			err := producer.Checker(context.Background(), checkState)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, expectedErrorMessage)
		})
	})

	Convey("Given an initialized producer", t, func() {
		fakeKafkaProducer = &FakeKafkaProducer{ProducerChannels: kafka.CreateProducerChannels()}
		fakeKafkaProducer.ProducerChannels.Output = make(chan []byte, 9999) // output will buffer what it receives, so we can inspect it below

		sut := event.ReindexRequestedProducer{Marshaller: fakeMarshaller, Producer: fakeKafkaProducer}

		Convey("When Close is called, Then the kafka producer should be closed", func() {
			err := sut.Close(context.Background())
			So(err, ShouldBeNil)
			So(fakeKafkaProducer.CloseCalls, ShouldHaveLength, 1)
			So(fakeKafkaProducer.CloseCalls[0], ShouldEqual, context.Background())
		})

		Convey("When Close is called, But the kafka producer returns an error, Then the error should be returned", func() {
			kafkaProviderError := errors.New("an error")
			fakeKafkaProducer.ReturnError(kafkaProviderError)
			actualError := sut.Close(context.Background())
			So(actualError, ShouldEqual, kafkaProviderError)
		})

		Convey("When Checker is called, Then the kafka producer checker should be called", func() {
			err := sut.Checker(context.Background(), checkState)
			So(err, ShouldBeNil)
			So(fakeKafkaProducer.CheckerCalls, ShouldHaveLength, 1)
			So(fakeKafkaProducer.CheckerCalls[0], ShouldResemble, checkerCall{context.Background(), checkState})
		})

		Convey("When Checker is called, But the kafka producer returns an error, Then the error should be returned", func() {
			kafkaProviderError := errors.New("an error")
			fakeKafkaProducer.ReturnError(kafkaProviderError)
			actualError := sut.Checker(context.Background(), checkState)
			So(actualError, ShouldEqual, kafkaProviderError)
		})

		Convey("When produce reindex request is called", func() {
			fakeMarshalledRequest := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
			fakeMarshaller.ReturnBytes(fakeMarshalledRequest)

			err := sut.ProduceReindexRequested(context.Background(), reindexRequestedEvent)
			So(err, ShouldBeNil)

			Convey("Then it should marshall the request as expected", func() {
				So(fakeMarshaller.MarshalCalls, ShouldHaveLength, 1)
				So(fakeMarshaller.MarshalCalls[0], ShouldResemble, reindexRequestedEvent)
			})

			Convey("When produce reindex request is called, Then it should send the marshalled event to the producer's Output channel", func() {
				actualSentMessage := readFromOutputChannel(fakeKafkaProducer)
				So(actualSentMessage, ShouldResemble, fakeMarshalledRequest)
			})
		})

		Convey("When produce reindex request is called, But the marshaller returns an error", func() {
			fakeMarshaller.ReturnError(errors.New("marshal error"))

			actualError := sut.ProduceReindexRequested(context.Background(), reindexRequestedEvent)

			Convey("Then the producer returns an error regarding the marshaller", func() {
				So(actualError.Error(), ShouldContainSubstring, "marshal error")
			})

			Convey("Then it should not send a message to the kafka producer", func() {
				actualSentMessage := readFromOutputChannel(fakeKafkaProducer)
				So(actualSentMessage, ShouldBeNil)
			})
		})
	})

	Convey("Given an initialized producer that has an incorrectly configured Output channel", t, func() {
		fakeKafkaProducer := &FakeKafkaProducer{ProducerChannels: kafka.CreateProducerChannels()}

		// The buffer size is the number of elements that can be sent to the channel without the send blocking.
		// By giving the producer channel a buffer size of 0 this means that every single send will block.
		fakeKafkaProducer.ProducerChannels.Output = make(chan []byte)

		fakeMarshaller := &FakeMarshaller{}
		fakeMarshaller.ErrorToReturn = nil // The marshaller should work correctly

		sut := event.ReindexRequestedProducer{Marshaller: fakeMarshaller, Producer: fakeKafkaProducer}

		Convey("When produce reindex request is called", func() {
			fakeMarshalledRequest := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
			fakeMarshaller.ReturnBytes(fakeMarshalledRequest)

			produceEventErr := sut.ProduceReindexRequested(context.Background(), reindexRequestedEvent)

			Convey("Then it should marshall the request as expected", func() {
				So(fakeMarshaller.MarshalCalls, ShouldHaveLength, 1)
				So(fakeMarshaller.MarshalCalls[0], ShouldResemble, reindexRequestedEvent)
			})

			Convey("Then it should wait 5 seconds and fail to send the marshalled event to the producer's blocked Output channel", func() {
				So(produceEventErr.Error(), ShouldContainSubstring, "producer output channel failed to read reindex-requested event")
				actualSentMessage := readFromOutputChannel(fakeKafkaProducer)
				fakeEmptyRequest := []byte(nil)
				So(actualSentMessage, ShouldResemble, fakeEmptyRequest)
			})
		})
	})
}

func readFromOutputChannel(fakeKafkaProducer *FakeKafkaProducer) []byte {
	var actualSentMessage []byte
	select {
	case actualSentMessage = <-fakeKafkaProducer.ProducerChannels.Output:
		// we just need to read the message
	default:
		// do nothing
	}
	return actualSentMessage
}

// A mock marshaller
type FakeMarshaller struct {
	BytesToReturn []byte
	MarshalCalls  []interface{}
	ErrorToReturn error
}

func (f *FakeMarshaller) Marshal(s interface{}) ([]byte, error) {
	f.MarshalCalls = append(f.MarshalCalls, s)
	if f.ErrorToReturn != nil {
		return nil, f.ErrorToReturn
	}
	return f.BytesToReturn, nil
}

func (f *FakeMarshaller) ReturnBytes(request []byte) {
	f.BytesToReturn = request
}

func (f *FakeMarshaller) ReturnError(err error) {
	f.ErrorToReturn = err
}

// A mock IKafkaProducer
type FakeKafkaProducer struct {
	CloseCalls       []context.Context
	CheckerCalls     []checkerCall
	ErrorToReturn    error
	ProducerChannels *kafka.ProducerChannels
}

func (f *FakeKafkaProducer) Channels() *kafka.ProducerChannels {
	return f.ProducerChannels
}

func (f *FakeKafkaProducer) IsInitialised() bool {
	panic("implement me")
}

func (f *FakeKafkaProducer) Initialise(ctx context.Context) error {
	panic("implement me")
}

func (f *FakeKafkaProducer) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	f.CheckerCalls = append(f.CheckerCalls, checkerCall{ctx, state})
	return f.ErrorToReturn
}

func (f *FakeKafkaProducer) Close(ctx context.Context) (err error) {
	f.CloseCalls = append(f.CloseCalls, ctx)
	return f.ErrorToReturn
}

func (f *FakeKafkaProducer) ReturnError(err error) {
	f.ErrorToReturn = err
}

func (f *FakeKafkaProducer) AddHeader(key, value string) {
	return
}

// a fake type which represents a call to IKafkaProducer.Checker
type checkerCall struct {
	Context    context.Context
	CheckState *healthcheck.CheckState
}
