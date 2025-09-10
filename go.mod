module github.com/theopenlane/core

go 1.25.1

replace github.com/oNaiPs/go-generate-fast => github.com/golanglemonade/go-generate-fast v0.0.0-20250908221431-37f3149f009b

tool (
	github.com/dave/jennifer
	github.com/invopop/jsonschema
	github.com/invopop/yaml
	github.com/vektra/mockery/v3
	gotest.tools/gotestsum
)

require (
	ariga.io/entcache v0.1.0
	cloud.google.com/go/secretmanager v1.15.0
	entgo.io/contrib v0.7.0
	entgo.io/ent v0.14.5
	github.com/99designs/gqlgen v0.17.78
	github.com/JesusIslam/tldr v0.6.0
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2
	github.com/Yamashou/gqlgenc v0.33.0
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/alitto/pond/v2 v2.5.0
	github.com/aws/aws-sdk-go-v2 v1.39.0
	github.com/aws/aws-sdk-go-v2/config v1.31.8
	github.com/aws/aws-sdk-go-v2/credentials v1.18.12
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.19.6
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.1
	github.com/brianvoe/gofakeit/v7 v7.6.0
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/cloudflare/cloudflare-go/v6 v6.0.0
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.18.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/gabriel-vasile/mimetype v1.4.10
	github.com/gertd/go-pluralize v0.2.1
	github.com/getkin/kin-openapi v0.133.0
	github.com/go-jose/go-jose/v4 v4.1.2
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/go-webauthn/webauthn v0.13.4
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/goccy/go-yaml v1.18.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/hashicorp/go-multierror v1.1.1
	github.com/invopop/jsonschema v0.13.0
	github.com/invopop/yaml v0.3.1
	github.com/jackc/pgx/v5 v5.7.6
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/knadh/koanf/providers/file v1.2.0
	github.com/knadh/koanf/providers/posflag v1.0.1
	github.com/knadh/koanf/v2 v2.2.2
	github.com/labstack/echo-contrib v0.17.4
	github.com/labstack/echo/v4 v4.13.4
	github.com/labstack/gommon v0.4.2
	github.com/lestrrat-go/httprc/v3 v3.0.1
	github.com/lestrrat-go/jwx/v3 v3.0.10
	github.com/lib/pq v1.10.9
	github.com/manifoldco/promptui v0.9.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nyaruka/phonenumbers v1.6.5
	github.com/oklog/ulid/v2 v2.1.1
	github.com/openfga/go-sdk v0.7.1
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/pquerna/otp v1.5.0
	github.com/pressly/goose/v3 v3.25.0
	github.com/prometheus/client_golang v1.23.2
	github.com/ravilushqa/otelgqlgen v0.19.0
	github.com/redis/go-redis/v9 v9.14.0
	github.com/riverqueue/river v0.24.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.24.0
	github.com/riverqueue/river/rivertype v0.24.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.34.0
	github.com/samber/lo v1.51.0
	github.com/sebdah/goldie/v2 v2.7.1
	github.com/spf13/cast v1.10.0
	github.com/spf13/cobra v1.10.1
	github.com/stoewer/go-strcase v1.3.1
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v82 v82.5.0
	github.com/theopenlane/beacon v0.2.0
	github.com/theopenlane/echo-prometheus v0.1.0
	github.com/theopenlane/echox v0.2.4
	github.com/theopenlane/emailtemplates v0.2.4
	github.com/theopenlane/entx v0.15.0
	github.com/theopenlane/gqlgen-plugins v0.9.0
	github.com/theopenlane/httpsling v0.2.2
	github.com/theopenlane/iam v0.17.0
	github.com/theopenlane/newman v0.2.0
	github.com/theopenlane/riverboat v0.4.0
	github.com/theopenlane/utils v0.5.0
	github.com/tink-crypto/tink-go/v2 v2.4.0
	github.com/tmc/langchaingo v0.1.13
	github.com/unidoc/unioffice v1.39.0
	github.com/urfave/cli/v3 v3.4.1
	github.com/vektah/gqlparser/v2 v2.5.30
	github.com/windmill-labs/windmill-go-client v1.541.1
	github.com/wundergraph/graphql-go-tools v1.67.4
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/zitadel/oidc/v3 v3.44.0
	gocloud.dev v0.43.0
	golang.org/x/crypto v0.42.0
	golang.org/x/mod v0.28.0
	golang.org/x/oauth2 v0.31.0
	golang.org/x/sync v0.17.0
	golang.org/x/term v0.35.0
	golang.org/x/text v0.29.0
	golang.org/x/tools v0.37.0
	google.golang.org/api v0.249.0
	gopkg.in/cheggaaa/pb.v2 v2.0.7
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.2
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	ariga.io/atlas v0.36.1
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.16.5 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/XSAM/otelsql v0.40.0 // indirect
	github.com/Yiling-J/theine-go v0.6.2 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alixaxel/pagerank v0.0.0-20200105181019-900657b89dcb // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.7 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.4 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/brunoga/deep v1.2.4 // indirect
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
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/dave/jennifer v1.7.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/docker v28.4.0+incompatible // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/go-github/v74 v74.0.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.6.7 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/knadh/koanf/providers/structs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250827001030-24949be3fa54 // indirect
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
	github.com/oapi-codegen/runtime v1.1.2 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.1.1 // indirect
	github.com/onsi/gomega v1.36.3 // indirect
	github.com/pkoukk/tiktoken-go v0.1.7 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.8 // indirect
	github.com/testcontainers/testcontainers-go v0.38.0 // indirect
	github.com/testcontainers/testcontainers-go/modules/openfga v0.38.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/vektra/mockery/v3 v3.5.4 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	github.com/zitadel/logging v0.6.2 // indirect
	github.com/zitadel/schema v1.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20250826171959-ef028d996bc1 // indirect
	gopkg.in/VividCortex/ewma.v1 v1.1.1 // indirect
	gopkg.in/fatih/color.v1 v1.7.0 // indirect
	gopkg.in/mattn/go-colorable.v0 v0.1.0 // indirect
	gopkg.in/mattn/go-isatty.v0 v0.0.4 // indirect
	gopkg.in/mattn/go-runewidth.v0 v0.0.4 // indirect
	gotest.tools/gotestsum v1.12.3 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/inflect v0.21.3 // indirect
	github.com/go-openapi/jsonpointer v0.22.0 // indirect
	github.com/go-openapi/swag/jsonname v0.24.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-webauthn/x v0.1.24 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/wrap v0.2.0 // indirect
	github.com/olekukonko/cat v0.0.0-20250908003013-b0de306c343b // indirect
	github.com/olekukonko/tablewriter v1.0.9 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.3.1 // indirect
	github.com/openfga/api/proto v0.0.0-20250814141243-c0b62b28b14d // indirect
	github.com/openfga/language/pkg/go v0.2.0-beta.2.0.20250428093642-7aeebe78bbfe // indirect
	github.com/openfga/openfga v1.9.5 // indirect
	github.com/ory/dockertest v3.3.5+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/resend/resend-go/v2 v2.23.0 // indirect
	github.com/richardlehane/msoleps v1.0.4 // indirect
	github.com/riverqueue/river/riverdriver v0.24.0 // indirect
	github.com/riverqueue/river/rivershared v0.24.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/theopenlane/dbx v0.1.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/woodsbury/decimal128 v1.4.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/zalando/go-keyring v0.2.6 // indirect
	github.com/zclconf/go-cty v1.17.0 // indirect
	go.opentelemetry.io/contrib v1.38.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250826171959-ef028d996bc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
	google.golang.org/grpc v1.75.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
