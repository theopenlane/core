module github.com/theopenlane/core

go 1.26.2

tool (
	github.com/dave/jennifer
	github.com/invopop/jsonschema
	github.com/invopop/yaml
	github.com/vektra/mockery/v3
	gotest.tools/gotestsum
)

replace github.com/theopenlane/core/common => ./common

require (
	ariga.io/entcache v0.1.0
	cloud.google.com/go/securitycenter v1.42.0
	entgo.io/contrib v0.7.0
	entgo.io/ent v0.14.6
	github.com/99designs/gqlgen v0.17.90
	github.com/AfterShip/email-verifier v1.4.1
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity v0.14.0
	github.com/alicebob/miniredis/v2 v2.37.0
	github.com/alitto/pond/v2 v2.7.1
	github.com/aws/aws-sdk-go-v2 v1.41.7
	github.com/aws/aws-sdk-go-v2/config v1.32.16
	github.com/aws/aws-sdk-go-v2/credentials v1.19.15
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.22.16
	github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager v0.1.18
	github.com/aws/aws-sdk-go-v2/service/configservice v1.62.2
	github.com/aws/aws-sdk-go-v2/service/iam v1.53.8
	github.com/aws/aws-sdk-go-v2/service/s3 v1.101.0
	github.com/brianvoe/gofakeit/v7 v7.14.1
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/cloudflare/cloudflare-go/v6 v6.10.0
	github.com/didasy/tldr v0.7.0
	github.com/elimity-com/scim v0.0.0-20260506142751-830e1caafcc3
	github.com/fatih/camelcase v1.0.0
	github.com/fatih/color v1.19.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/fumiama/go-docx v0.0.0-20250506085032-0c30fd09304b
	github.com/gabriel-vasile/mimetype v1.4.13
	github.com/gertd/go-pluralize v0.2.1
	github.com/getkin/kin-openapi v0.137.0
	github.com/go-jose/go-jose/v4 v4.1.4
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/go-webauthn/webauthn v0.17.2
	github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
	github.com/goccy/go-yaml v1.19.2
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/go-github/v84 v84.0.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/gqlgo/gqlgenc v0.36.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/invopop/jsonschema v0.14.0
	github.com/invopop/yaml v0.3.1
	github.com/jackc/pgx/v5 v5.9.2
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/env/v2 v2.0.0
	github.com/knadh/koanf/providers/file v1.2.1
	github.com/knadh/koanf/providers/posflag v1.0.1
	github.com/knadh/koanf/v2 v2.3.4
	github.com/labstack/echo-contrib v0.50.1
	github.com/labstack/echo/v4 v4.15.1
	github.com/labstack/gommon v0.5.0
	github.com/lestrrat-go/httprc/v3 v3.0.5
	github.com/lestrrat-go/jwx/v3 v3.1.0
	github.com/lib/pq v1.12.3
	github.com/manifoldco/promptui v0.9.0
	github.com/mcuadros/go-defaults v1.2.0
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/microsoft/kiota-authentication-azure-go v1.3.1
	github.com/microsoftgraph/msgraph-sdk-go v1.97.0
	github.com/nyaruka/phonenumbers v1.7.2
	github.com/oklog/ulid/v2 v2.1.1
	github.com/okta/okta-sdk-golang/v6 v6.1.6
	github.com/openfga/go-sdk v0.8.0
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/pquerna/otp v1.5.0
	github.com/pressly/goose/v3 v3.27.1
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.19.0
	github.com/riverqueue/river v0.32.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.32.0
	github.com/riverqueue/river/rivertype v0.32.0
	github.com/rs/zerolog v1.35.1
	github.com/samber/do/v2 v2.0.0
	github.com/samber/lo v1.53.0
	github.com/shurcooL/githubv4 v0.0.0-20260209031235-2402fdf4a9ed
	github.com/slack-go/slack v0.23.0
	github.com/spf13/cobra v1.10.2
	github.com/stoewer/go-strcase v1.3.1
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v84 v84.4.1
	github.com/theopenlane/core/common v1.0.21
	github.com/theopenlane/echo-prometheus v0.1.0
	github.com/theopenlane/echox v0.3.0
	github.com/theopenlane/entx v0.27.1
	github.com/theopenlane/go-client v0.10.0
	github.com/theopenlane/gqlgen-plugins v0.14.7
	github.com/theopenlane/httpsling v0.3.0
	github.com/theopenlane/iam v0.28.0
	github.com/theopenlane/newman v0.4.0
	github.com/theopenlane/riverboat v0.8.8
	github.com/theopenlane/utils v0.7.0
	github.com/tmc/langchaingo v0.1.14
	github.com/urfave/cli/v3 v3.8.0
	github.com/vektah/gqlparser/v2 v2.5.33
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/yuin/goldmark v1.8.2
	github.com/zitadel/oidc/v3 v3.46.0
	gocloud.dev v0.45.0
	golang.org/x/crypto v0.50.0
	golang.org/x/mod v0.35.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.20.0
	golang.org/x/text v0.36.0
	golang.org/x/tools v0.44.0
	google.golang.org/api v0.277.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.2
)

require (
	github.com/AzureAD/microsoft-authentication-library-for-go v1.7.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/PuerkitoBio/goquery v1.12.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/inbucket/html2text v1.0.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/microsoft/kiota-abstractions-go v1.9.4 // indirect
	github.com/microsoft/kiota-http-go v1.5.4 // indirect
	github.com/microsoft/kiota-serialization-form-go v1.1.3 // indirect
	github.com/microsoft/kiota-serialization-json-go v1.1.2 // indirect
	github.com/microsoft/kiota-serialization-multipart-go v1.1.2 // indirect
	github.com/microsoft/kiota-serialization-text-go v1.1.3 // indirect
	github.com/microsoftgraph/msgraph-sdk-go-core v1.4.0 // indirect
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/std-uritemplate/std-uritemplate/go/v2 v2.0.3 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/vanng822/css v1.0.1 // indirect
	github.com/vanng822/go-premailer v1.33.0 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
)

require (
	cloud.google.com/go/longrunning v0.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.12.0 // indirect
	github.com/Yamashou/gqlgenc v0.33.0
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.10 // indirect
	github.com/clipperhouse/displaywidth v0.6.2 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.4.0 // indirect
	github.com/di-wu/parser v0.3.0 // indirect
	github.com/di-wu/xsd-datetime v1.0.0 // indirect
	github.com/fumiama/imgsz v0.0.4 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/samber/go-type-to-string v1.8.0 // indirect
	github.com/scim2/filter-parser/v2 v2.2.0 // indirect
	github.com/shurcooL/graphql v0.0.0-20240915155400-7ee5256398cf
	github.com/spf13/cast v1.10.0 // indirect
	github.com/theopenlane/oscalot v0.1.0 // indirect
	github.com/wundergraph/astjson v1.1.0
	github.com/wundergraph/go-arena v0.0.0-20251008210416-55cb97e6f68f // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/image v0.38.0 // indirect
	golang.org/x/term v0.42.0 // indirect
)

require (
	ariga.io/atlas v1.2.0
	cel.dev/expr v0.25.1 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.7.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/XSAM/otelsql v0.42.0 // indirect
	github.com/Yiling-J/theine-go v0.6.2 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alixaxel/pagerank v0.0.0-20200105181019-900657b89dcb // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.0
	github.com/aws/smithy-go v1.25.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/boombuler/barcode v1.1.0 // indirect
	github.com/brunoga/deep v1.3.1 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
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
	github.com/dave/jennifer v1.7.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/docker/docker v28.5.2+incompatible // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/jedib0t/go-pretty/v6 v6.7.8 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/knadh/koanf/providers/structs v1.0.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.2.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/oasdiff/yaml v0.0.9 // indirect
	github.com/oasdiff/yaml3 v0.0.12 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.1.4-0.20260115111900-9e59c2286df0 // indirect
	github.com/onsi/gomega v1.38.2 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.26.2 // indirect
	github.com/testcontainers/testcontainers-go v0.41.0
	github.com/testcontainers/testcontainers-go/modules/openfga v0.41.0 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/vektra/mockery/v3 v3.7.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zclconf/go-cty-yaml v1.2.0 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	github.com/zitadel/logging v0.7.0 // indirect
	github.com/zitadel/schema v1.3.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/genproto v0.0.0-20260319201613-d00831a3d3e7 // indirect
	gotest.tools/gotestsum v1.13.0 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/service/securityhub v1.70.0
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/go-connections v0.7.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/inflect v0.21.5 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-webauthn/x v0.2.3 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.28.0
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.15 // indirect
	github.com/googleapis/gax-go/v2 v2.22.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hbollon/go-edlib v1.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/env v1.1.0
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/hashstructure v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/natefinch/wrap v0.2.0 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/tablewriter v1.1.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.3.3 // indirect
	github.com/openfga/api/proto v0.0.0-20260319214821-f153694bfc20 // indirect
	github.com/openfga/language/pkg/go v0.2.1 // indirect
	github.com/openfga/openfga v1.15.0 // indirect
	github.com/ory/dockertest v3.3.5+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/resend/resend-go/v3 v3.6.0
	github.com/riverqueue/river/riverdriver v0.32.0 // indirect
	github.com/riverqueue/river/rivershared v0.32.0 // indirect
	github.com/riverqueue/rivercontrib/otelriver v0.7.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/samber/mo v1.16.0
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/pflag v1.0.10
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
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/zclconf/go-cty v1.18.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.11
)
