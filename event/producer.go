package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-search-reindex-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out ./mock/producer.go -pkg mock . Marshaller

// Marshaller defines a type for marshalling the requested object into a stream of bytes which can be sent to the kafka
// producer channel
type Marshaller interface {
	Marshal(s interface{}) ([]byte, error)
}

// ReindexRequestedProducer is responsible for sending a ReindexRequested event object to the supplied kafka producer,
// as well as cleanly closing the producer and exposing health check information
type ReindexRequestedProducer struct {
	Marshaller Marshaller
	Producer   kafka.IProducer
}

// ProduceReindexRequested produce a kafka message for an instance which has been successfully processed.
func (p ReindexRequestedProducer) ProduceReindexRequested(ctx context.Context, event models.ReindexRequested) error {
	logData := log.Data{"event": event}
	log.Info(ctx, "sending reindex-requested event", logData)

	if err := p.ensureDependencies(ctx); err != nil {
		log.Error(ctx, "failed to ensure dependencies", err, logData)
		return err
	}

	bytes, err := p.Marshaller.Marshal(event)
	if err != nil {
		log.Fatal(ctx, "failed to marshal reindex-requested event", err, logData)
		return fmt.Errorf(fmt.Sprintf("Marshaller.Marshal returned an error: event=%v: %%w", event), err)
	}

	var timeout = time.Second * 5
	delay := time.NewTimer(timeout)
	select {
	case p.Producer.Channels().Output <- bytes:
		// Ensure timer is stopped and its resources are freed
		if !delay.Stop() {
			// if the timer has been stopped then read from the channel
			<-delay.C
		}
		logData["bytes_sent"] = bytes
		logData["package"] = "event.ReindexRequestedProducer"
		log.Info(ctx, "reindex-requested event sent", logData)
		return nil

	case <-delay.C:
		logData["timeout"] = timeout
		log.Fatal(ctx, "producer output channel failed to read bytes", err, logData)
		return fmt.Errorf(fmt.Sprintf("producer output channel failed to read reindex-requested event: event=%v: %%w", event), err)
	}
}

func (p ReindexRequestedProducer) ensureDependencies(ctx context.Context) error {
	if p.Marshaller == nil {
		err := errors.New("marshaller is not provided")
		log.Error(ctx, "marshaller is nil", err)
		return err
	}
	if p.Producer == nil {
		err := errors.New("producer is not provided")
		log.Error(ctx, "producer is nil", err)
		return err
	}

	return nil
}

// Close is called when the service shuts down gracefully
func (p ReindexRequestedProducer) Close(ctx context.Context) error {
	if err := p.ensureDependencies(ctx); err != nil {
		log.Error(ctx, "failed to ensure dependencies", err)
		return err
	}

	return p.Producer.Close(ctx)
}

// Checker is called by the healthcheck library to check the health state of this kafka producer instance
func (p *ReindexRequestedProducer) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if err := p.ensureDependencies(ctx); err != nil {
		log.Error(ctx, "failed to ensure dependencies", err)
		return err
	}

	return p.Producer.Checker(ctx, state)
}
