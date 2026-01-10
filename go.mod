module github.com/theopenlane/core

go 1.25.5

tool (
	github.com/dave/jennifer
	github.com/invopop/jsonschema
	github.com/invopop/yaml
	github.com/vektra/mockery/v3
	gotest.tools/gotestsum
)

require (
	ariga.io/entcache v0.1.0
	cloud.google.com/go/secretmanager v1.16.0
	cloud.google.com/go/securitycenter v1.38.1
	entgo.io/contrib v0.7.0
	entgo.io/ent v0.14.6-0.20260108211107-2eb36418a02e
	github.com/99designs/gqlgen v0.17.85
	github.com/AfterShip/email-verifier v1.4.1
	github.com/Yamashou/gqlgenc v0.33.0
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/alitto/pond/v2 v2.6.0
	github.com/aws/aws-sdk-go-v2 v1.41.0
	github.com/aws/aws-sdk-go-v2/config v1.32.6
	github.com/aws/aws-sdk-go-v2/credentials v1.19.6
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.20.18
	github.com/aws/aws-sdk-go-v2/service/s3 v1.95.0
	github.com/brianvoe/gofakeit/v7 v7.14.0
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/didasy/tldr v0.7.0
	github.com/elimity-com/scim v0.0.0-20240320110924-172bf2aee9c8
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.18.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/gabriel-vasile/mimetype v1.4.12
	github.com/gertd/go-pluralize v0.2.1
	github.com/getkin/kin-openapi v0.133.0
	github.com/go-jose/go-jose/v4 v4.1.3
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/go-webauthn/webauthn v0.15.0
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/goccy/go-yaml v1.19.2
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gomarkdown/markdown v0.0.0-20250810172220-2e2c11897d1a
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/hashicorp/go-multierror v1.1.1
	github.com/invopop/jsonschema v0.13.0
	github.com/invopop/yaml v0.3.1
	github.com/jackc/pgx/v5 v5.8.0
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/env/v2 v2.0.0
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/providers/posflag v1.0.1
	github.com/knadh/koanf/v2 v2.3.0
	github.com/labstack/echo-contrib v0.17.4
	github.com/labstack/echo/v4 v4.15.0
	github.com/labstack/gommon v0.4.2
	github.com/lestrrat-go/httprc/v3 v3.0.3
	github.com/lestrrat-go/jwx/v3 v3.0.12
	github.com/lib/pq v1.10.9
	github.com/manifoldco/promptui v0.9.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/nyaruka/phonenumbers v1.6.8
	github.com/oklog/ulid/v2 v2.1.1
	github.com/openfga/go-sdk v0.7.3
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/pquerna/otp v1.5.0
	github.com/pressly/goose/v3 v3.26.0
	github.com/prometheus/client_golang v1.23.2
	github.com/ravilushqa/otelgqlgen v0.19.0
	github.com/redis/go-redis/v9 v9.17.2
	github.com/riverqueue/river v0.29.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.29.0
	github.com/rs/zerolog v1.34.0
	github.com/samber/lo v1.52.0
	github.com/spf13/cobra v1.10.2
	github.com/stoewer/go-strcase v1.3.1
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v84 v84.1.0
	github.com/theopenlane/beacon v0.3.0
	github.com/theopenlane/core/common v1.0.0
	github.com/theopenlane/echo-prometheus v0.1.0
	github.com/theopenlane/echox v0.2.4
	github.com/theopenlane/emailtemplates v0.3.0
	github.com/theopenlane/entx v0.20.3
	github.com/theopenlane/go-client v0.4.3
	github.com/theopenlane/gqlgen-plugins v0.13.2
	github.com/theopenlane/httpsling v0.3.0
	github.com/theopenlane/iam v0.23.2
	github.com/theopenlane/newman v0.2.1
	github.com/theopenlane/riverboat v0.7.4
	github.com/theopenlane/utils v0.6.2
	github.com/tmc/langchaingo v0.1.14
	github.com/urfave/cli/v3 v3.6.1
	github.com/vektah/gqlparser/v2 v2.5.31
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/zitadel/oidc/v3 v3.45.1
	gocloud.dev v0.44.0
	golang.org/x/crypto v0.46.0
	golang.org/x/mod v0.32.0
	golang.org/x/oauth2 v0.34.0
	golang.org/x/sync v0.19.0
	golang.org/x/text v0.33.0
	golang.org/x/tools v0.40.0
	google.golang.org/api v0.259.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.2
)

require (
	cloud.google.com/go/longrunning v0.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.4 // indirect
	github.com/clipperhouse/displaywidth v0.6.2 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.0 // indirect
	github.com/di-wu/parser v0.3.0 // indirect
	github.com/di-wu/xsd-datetime v1.0.0 // indirect
	github.com/google/go-github/v80 v80.0.0 // indirect
	github.com/riverqueue/river/rivertype v0.29.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/scim2/filter-parser/v2 v2.2.0 // indirect
	github.com/slack-go/slack v0.17.3 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	golang.org/x/term v0.38.0 // indirect
)

require (
	ariga.io/atlas v1.0.0
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.18.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/XSAM/otelsql v0.41.0 // indirect
	github.com/Yiling-J/theine-go v0.6.2 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alixaxel/pagerank v0.0.0-20200105181019-900657b89dcb // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.5 // indirect
	github.com/aws/smithy-go v1.24.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/brunoga/deep v1.2.5 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/dave/jennifer v1.7.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/docker v28.5.2+incompatible // indirect
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.7.3 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/knadh/koanf/providers/structs v1.0.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.1.3 // indirect
	github.com/onsi/gomega v1.38.2 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.25.9 // indirect
	github.com/testcontainers/testcontainers-go v0.40.0
	github.com/testcontainers/testcontainers-go/modules/openfga v0.40.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/vektra/mockery/v3 v3.6.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zclconf/go-cty-yaml v1.2.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	github.com/zitadel/logging v0.6.2 // indirect
	github.com/zitadel/schema v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.64.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20251222181119-0a764e51fe1b // indirect
	gotest.tools/gotestsum v1.13.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/inflect v0.21.5 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-webauthn/x v0.1.27 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.16.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.4 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hbollon/go-edlib v1.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/wrap v0.2.0 // indirect
	github.com/nguyenthenguyen/docx v0.0.0-20230621112118-9c8e795a11db
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/tablewriter v1.1.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.3.3 // indirect
	github.com/openfga/api/proto v0.0.0-20251105142303-feed3db3d69d // indirect
	github.com/openfga/language/pkg/go v0.2.0-beta.2.0.20251027165255-0f8f255e5f6c // indirect
	github.com/openfga/openfga v1.11.2 // indirect
	github.com/ory/dockertest v3.3.5+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/resend/resend-go/v2 v2.28.0 // indirect
	github.com/riverqueue/river/riverdriver v0.29.0 // indirect
	github.com/riverqueue/river/rivershared v0.29.0 // indirect
	github.com/riverqueue/rivercontrib/otelriver v0.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/samber/mo v1.16.0
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/theopenlane/dbx v0.1.3 // indirect
	github.com/theopenlane/eddy v0.1.0
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/woodsbury/decimal128 v1.4.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/zclconf/go-cty v1.17.0 // indirect
	go.opentelemetry.io/contrib v1.39.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.64.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
