package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()
	var err error
	var configuration *Config

	Convey("Given an environment with no environment variables set", t, func() {
		Convey("Then cfg should be nil", func() {
			So(cfg, ShouldBeNil)
		})

		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned, and values are as expected", func() {
				configuration, err = Get() // This Get() is only called once, when inside this function
				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &Config{
					BindAddr:                   "localhost:25700",
					GracefulShutdownTimeout:    20 * time.Second,
					HealthCheckInterval:        30 * time.Second,
					HealthCheckCriticalTimeout: 90 * time.Second,
					MaxReindexJobRuntime:       3600 * time.Second,
					MongoConfig: MongoConfig{
						BindAddr:        "localhost:27017",
						JobsCollection:  "jobs",
						LocksCollection: "jobs_locks",
						TasksCollection: "tasks",
						Database:        "search",
					},
					DefaultMaxLimit:  1000,
					DefaultLimit:     20,
					DefaultOffset:    0,
					ZebedeeURL:       "http://localhost:8082",
					TaskNameValues:   "dataset-api,zebedee",
					SearchAPIURL:     "http://localhost:23900",
					ServiceAuthToken: "",
					KafkaConfig: KafkaConfig{
						Brokers:               []string{"localhost:9092", "localhost:9093", "localhost:9094"},
						Version:               "1.0.2",
						SecProtocol:           "",
						SecCACerts:            "",
						SecClientCert:         "",
						SecClientKey:          "",
						SecSkipVerify:         false,
						ReindexRequestedTopic: "reindex-requested",
					},
				})
			})

			Convey("Then a second call to config should return the same config", func() {
				// This achieves code coverage of the first return in the Get() function.
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
