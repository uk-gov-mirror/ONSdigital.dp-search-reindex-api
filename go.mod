module github.com/ONSdigital/dp-search-reindex-api

go 1.23

// to avoid the following vulnerabilities:
//     - CVE-2023-32731 # pkg:google.golang.org/grpc
replace google.golang.org/grpc => google.golang.org/grpc v1.55.0

// [CVE-2024-24786] CWE-835: Loop with Unreachable Exit Condition ('Infinite Loop')
replace google.golang.org/protobuf => google.golang.org/protobuf v1.33.0

// [CVE-2023-48795] CWE-354: Improper Validation of Integrity Check Value
replace golang.org/x/crypto => golang.org/x/crypto v0.17.0

// [CVE-2024-37298] CWE-770: Allocation of Resources Without Limits or Throttling
replace github.com/gorilla/schema => github.com/gorilla/schema v1.4.1

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.261.0
	github.com/ONSdigital/dp-authorisation v0.3.0
	github.com/ONSdigital/dp-component-test v0.9.2
	github.com/ONSdigital/dp-healthcheck v1.6.3
	github.com/ONSdigital/dp-kafka/v2 v2.8.0
	github.com/ONSdigital/dp-mongodb/v3 v3.7.0
	github.com/ONSdigital/dp-net/v2 v2.11.2
	github.com/ONSdigital/log.go/v2 v2.4.3
	github.com/cucumber/godog v0.12.6
	github.com/cucumber/messages-go/v16 v16.0.1
	github.com/gorilla/mux v1.8.1
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/maxcnunes/httpfake v1.2.4
	github.com/pkg/errors v0.9.1
	github.com/rdumont/assistdog v0.0.0-20201106100018-168b06230d14
	github.com/satori/go.uuid v1.2.0
	github.com/smartystreets/goconvey v1.8.1
	github.com/stretchr/testify v1.9.0
	go.mongodb.org/mongo-driver v1.17.1
)

require (
	github.com/ONSdigital/dp-mongodb-in-memory v1.7.0 // indirect
	github.com/Shopify/sarama v1.38.1 // indirect
	github.com/chromedp/cdproto v0.0.0-20230419194459-b5ff65bc57a3 // indirect
	github.com/chromedp/chromedp v0.9.1 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.5.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.2.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/square/mongo-lock v0.0.0-20220601164918-701ecf357cd7 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
