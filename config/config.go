package config

import (
	"time"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/kelseyhightower/envconfig"
)

var cfg *Config

// Config represents service configuration for dp-search-reindex-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	LatestVersion              string        `envconfig:"LATEST_VERSION"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	MaxReindexJobRuntime       time.Duration `envconfig:"MAX_REINDEX_JOB_RUNTIME"`
	MongoConfig                MongoConfig
	DefaultMaxLimit            int    `envconfig:"DEFAULT_MAXIMUM_LIMIT"`
	DefaultLimit               int    `envconfig:"DEFAULT_LIMIT"`
	DefaultOffset              int    `envconfig:"DEFAULT_OFFSET"`
	ZebedeeURL                 string `envconfig:"ZEBEDEE_URL"`
	TaskNameValues             string `envconfig:"TASK_NAME_VALUES"`
	SearchAPIURL               string `envconfig:"SEARCH_API_URL"`
	ServiceAuthToken           string `envconfig:"SERVICE_AUTH_TOKEN"   json:"-"`
	KafkaConfig                KafkaConfig
}

// MongoConfig contains the config required to connect to DocumentDB.
type MongoConfig struct {
	mongodriver.MongoDriverConfig
}

// KafkaConfig contains the config required to connect to Kafka
type KafkaConfig struct {
	Brokers               []string `envconfig:"KAFKA_ADDR"                            json:"-"`
	Version               string   `envconfig:"KAFKA_VERSION"`
	SecProtocol           string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts            string   `envconfig:"KAFKA_SEC_CA_CERTS"`
	SecClientKey          string   `envconfig:"KAFKA_SEC_CLIENT_KEY"                  json:"-"`
	SecClientCert         string   `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	SecSkipVerify         bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ReindexRequestedTopic string   `envconfig:"KAFKA_REINDEX_REQUESTED_TOPIC"`
}

const (
	JobsCollection          = "JobsCollection"
	LocksCollection         = "LocksCollection"
	TasksCollection         = "TasksCollection"
)

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:25700",
		LatestVersion:              "v1",
		GracefulShutdownTimeout:    20 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		MaxReindexJobRuntime:       3600 * time.Second,
		MongoConfig: MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint:               "localhost:27017",
				Username:                      "",
				Password:                      "",
				Database:                      "search",
				Collections:                   map[string]string{JobsCollection: "jobs", LocksCollection: "jobs_locks", TasksCollection: "tasks"},
				ReplicaSet:                    "",
				IsStrongReadConcernEnabled:    false,
				IsWriteConcernMajorityEnabled: true,
				ConnectTimeout:                5 * time.Second,
				QueryTimeout:                  15 * time.Second,
				TLSConnectionConfig: mongodriver.TLSConnectionConfig{
					IsSSL: false,
				},
			},
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
	}

	return cfg, envconfig.Process("", cfg)
}
