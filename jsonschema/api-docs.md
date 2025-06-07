# object

Config contains the configuration for the core server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**refreshInterval**|`integer`|RefreshInterval determines how often to reload the config<br/>||
|[**server**](#server)|`object`|Server settings for the echo server<br/>|yes|
|[**entConfig**](#entconfig)|`object`|Config holds the configuration for the ent server<br/>||
|[**auth**](#auth)|`object`|Auth settings including oauth2 providers and token configuration<br/>|yes|
|[**authz**](#authz)|`object`||yes|
|[**db**](#db)|`object`||yes|
|[**jobQueue**](#jobqueue)|`object`|||
|[**redis**](#redis)|`object`|||
|[**tracer**](#tracer)|`object`|||
|[**email**](#email)|`object`|||
|[**sessions**](#sessions)|`object`|||
|[**totp**](#totp)|`object`|||
|[**ratelimit**](#ratelimit)|`object`|Config defines the configuration settings for the default rate limiter<br/>||
|[**objectStorage**](#objectstorage)|`object`|Config is the configuration for the object store<br/>||
|[**subscription**](#subscription)|`object`|||

**Additional Properties:** not allowed  
<a name="server"></a>
## server: object

Server settings for the echo server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**debug**|`boolean`|Debug enables debug mode for the server<br/>|no|
|**dev**|`boolean`|Dev enables echo's dev mode options<br/>|no|
|**listen**|`string`|Listen sets the listen address to serve the echo server on<br/>|yes|
|**metricsPort**|`string`|MetricsPort sets the port for the metrics endpoint<br/>|no|
|**shutdownGracePeriod**|`integer`|ShutdownGracePeriod sets the grace period for in flight requests before shutting down<br/>|no|
|**readTimeout**|`integer`|ReadTimeout sets the maximum duration for reading the entire request including the body<br/>|no|
|**writeTimeout**|`integer`|WriteTimeout sets the maximum duration before timing out writes of the response<br/>|no|
|**idleTimeout**|`integer`|IdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled<br/>|no|
|**readHeaderTimeout**|`integer`|ReadHeaderTimeout sets the amount of time allowed to read request headers<br/>|no|
|[**tls**](#servertls)|`object`|TLS settings for the server for secure connections<br/>|no|
|[**cors**](#servercors)|`object`|Config holds the cors configuration settings<br/>|no|
|[**secure**](#serversecure)|`object`|Config contains the types used in the mw middleware<br/>|no|
|[**redirects**](#serverredirects)|`object`|Config contains the types used in executing redirects via the redirect middleware<br/>|no|
|[**cacheControl**](#servercachecontrol)|`object`|Config is the config values for the cache-control middleware<br/>|no|
|[**mime**](#servermime)|`object`|Config defines the config for Mime middleware<br/>|no|
|[**graphPool**](#servergraphpool)|`object`|PondPool contains the settings for the goroutine pool<br/>|no|
|**enableGraphExtensions**|`boolean`|EnableGraphExtensions enables the graph extensions for the graph resolvers<br/>|no|
|**complexityLimit**|`integer`|ComplexityLimit sets the maximum complexity allowed for a query<br/>|no|
|**maxResultLimit**|`integer`|MaxResultLimit sets the maximum number of results allowed for a query<br/>|no|
|[**csrfProtection**](#servercsrfprotection)|`object`|Config defines configuration for the CSRF middleware wrapper.<br/>|no|

**Additional Properties:** not allowed  
<a name="servertls"></a>
### server\.tls: object

TLS settings for the server for secure connections


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled turns on TLS settings for the server<br/>||
|**certFile**|`string`|CertFile location for the TLS server<br/>||
|**certKey**|`string`|CertKey file location for the TLS server<br/>||
|**autoCert**|`boolean`|AutoCert generates the cert with letsencrypt, this does not work on localhost<br/>||

**Additional Properties:** not allowed  
<a name="servercors"></a>
### server\.cors: object

Config holds the cors configuration settings


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enable or disable the CORS middleware<br/>||
|[**prefixes**](#servercorsprefixes)|`object`|||
|[**allowOrigins**](#servercorsalloworigins)|`string[]`|||
|**cookieInsecure**|`boolean`|CookieInsecure sets the cookie to be insecure<br/>||

**Additional Properties:** not allowed  
<a name="servercorsprefixes"></a>
#### server\.cors\.prefixes: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#servercorsprefixesadditionalproperties)|`string[]`|||

<a name="servercorsprefixesadditionalproperties"></a>
##### server\.cors\.prefixes\.additionalProperties: array

**Items**

**Item Type:** `string`  
<a name="servercorsalloworigins"></a>
#### server\.cors\.allowOrigins: array

**Items**

**Item Type:** `string`  
<a name="serversecure"></a>
### server\.secure: object

Config contains the types used in the mw middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the secure middleware should be enabled<br/>||
|**xssprotection**|`string`|XSSProtection is the value to set the X-XSS-Protection header to - default is 1; mode=block<br/>||
|**contenttypenosniff**|`string`|ContentTypeNosniff is the value to set the X-Content-Type-Options header to - default is nosniff<br/>||
|**xframeoptions**|`string`|XFrameOptions is the value to set the X-Frame-Options header to - default is SAMEORIGIN<br/>||
|**hstspreloadenabled**|`boolean`|HSTSPreloadEnabled is a boolean to enable HSTS preloading - default is false<br/>||
|**hstsmaxage**|`integer`|HSTSMaxAge is the max age to set the HSTS header to - default is 31536000<br/>||
|**contentsecuritypolicy**|`string`|ContentSecurityPolicy is the value to set the Content-Security-Policy header to - default is default-src 'self'<br/>||
|**referrerpolicy**|`string`|ReferrerPolicy is the value to set the Referrer-Policy header to - default is same-origin<br/>||
|**cspreportonly**|`boolean`|CSPReportOnly is a boolean to enable the Content-Security-Policy-Report-Only header - default is false<br/>||

**Additional Properties:** not allowed  
<a name="serverredirects"></a>
### server\.redirects: object

Config contains the types used in executing redirects via the redirect middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the redirect middleware should be enabled<br/>||
|[**redirects**](#serverredirectsredirects)|`object`|||
|**code**|`integer`|Code is the HTTP status code to use for the redirect<br/>||

**Additional Properties:** not allowed  
<a name="serverredirectsredirects"></a>
#### server\.redirects\.redirects: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="servercachecontrol"></a>
### server\.cacheControl: object

Config is the config values for the cache-control middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**noCacheHeaders**](#servercachecontrolnocacheheaders)|`object`|||
|[**etagHeaders**](#servercachecontroletagheaders)|`string[]`|||

**Additional Properties:** not allowed  
<a name="servercachecontrolnocacheheaders"></a>
#### server\.cacheControl\.noCacheHeaders: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="servercachecontroletagheaders"></a>
#### server\.cacheControl\.etagHeaders: array

**Items**

**Item Type:** `string`  
<a name="servermime"></a>
### server\.mime: object

Config defines the config for Mime middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the mime middleware should be enabled<br/>||
|**mimeTypesFile**|`string`|MimeTypesFile is the file to load mime types from<br/>||
|**defaultContentType**|`string`|DefaultContentType is the default content type to set if no mime type is found<br/>||

**Additional Properties:** not allowed  
<a name="servergraphpool"></a>
### server\.graphPool: object

PondPool contains the settings for the goroutine pool


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**maxWorkers**|`integer`|MaxWorkers is the maximum number of workers in the pool<br/>||

**Additional Properties:** not allowed  
<a name="servercsrfprotection"></a>
### server\.csrfProtection: object

Config defines configuration for the CSRF middleware wrapper.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether CSRF protection is enabled.<br/>||
|**header**|`string`|Header specifies the header name to look for the CSRF token.<br/>||
|**cookie**|`string`|Cookie specifies the cookie name used to store the CSRF token.<br/>||
|**secure**|`boolean`|Secure sets the Secure flag on the CSRF cookie.<br/>||
|**sameSite**|`string`|SameSite configures the SameSite attribute on the CSRF cookie. Valid<br/>values are "Lax", "Strict", "None" and "Default".<br/>||

**Additional Properties:** not allowed  
<a name="entconfig"></a>
## entConfig: object

Config holds the configuration for the ent server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**entityTypes**](#entconfigentitytypes)|`string[]`|||
|[**summarizer**](#entconfigsummarizer)|`object`|Summarizer holds configuration for the text summarization functionality<br/>||

**Additional Properties:** not allowed  
<a name="entconfigentitytypes"></a>
### entConfig\.entityTypes: array

**Items**

**Item Type:** `string`  
<a name="entconfigsummarizer"></a>
### entConfig\.summarizer: object

Summarizer holds configuration for the text summarization functionality


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**type**|`string`|Type specifies the summarization algorithm to use<br/>||
|[**llm**](#entconfigsummarizerllm)|`object`|SummarizerLLM contains configuration for multiple LLM providers<br/>||
|**maximumSentences**|`integer`|MaximumSentences specifies the maximum number of sentences in the summary<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllm"></a>
#### entConfig\.summarizer\.llm: object

SummarizerLLM contains configuration for multiple LLM providers


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**provider**|`string`|Provider specifies which LLM service to use<br/>||
|[**anthropic**](#entconfigsummarizerllmanthropic)|`object`|AnthropicConfig contains Anthropic specific configuration<br/>||
|[**mistral**](#entconfigsummarizerllmmistral)|`object`|MistralConfig contains Mistral specific configuration<br/>||
|[**gemini**](#entconfigsummarizerllmgemini)|`object`|GeminiConfig contains Google Gemini specific configuration<br/>||
|[**huggingFace**](#entconfigsummarizerllmhuggingface)|`object`|HuggingFaceConfig contains HuggingFace specific configuration<br/>||
|[**ollama**](#entconfigsummarizerllmollama)|`object`|OllamaConfig contains Ollama specific configuration<br/>||
|[**cloudflare**](#entconfigsummarizerllmcloudflare)|`object`|CloudflareConfig contains Cloudflare specific configuration<br/>||
|[**openai**](#entconfigsummarizerllmopenai)|`object`|OpenAIConfig contains OpenAI specific configuration<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmanthropic"></a>
##### entConfig\.summarizer\.llm\.anthropic: object

AnthropicConfig contains Anthropic specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**betaHeader**|`string`|BetaHeader specifies the beta API features to enable<br/>||
|**legacyTextCompletion**|`boolean`|LegacyTextCompletion enables legacy text completion API<br/>||
|**baseURL**|`string`|BaseURL specifies the API endpoint<br/>||
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmmistral"></a>
##### entConfig\.summarizer\.llm\.mistral: object

MistralConfig contains Mistral specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmgemini"></a>
##### entConfig\.summarizer\.llm\.gemini: object

GeminiConfig contains Google Gemini specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||
|**credentialsPath**|`string`|CredentialsPath is the path to Google Cloud credentials file<br/>||
|**credentialsJSON**|`string`|CredentialsJSON contains Google Cloud credentials as JSON string<br/>||
|**maxTokens**|`integer`|MaxTokens specifies the maximum tokens for response<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmhuggingface"></a>
##### entConfig\.summarizer\.llm\.huggingFace: object

HuggingFaceConfig contains HuggingFace specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmollama"></a>
##### entConfig\.summarizer\.llm\.ollama: object

OllamaConfig contains Ollama specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model to use<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmcloudflare"></a>
##### entConfig\.summarizer\.llm\.cloudflare: object

CloudflareConfig contains Cloudflare specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||
|**accountID**|`string`|AccountID specifies the Cloudflare account ID<br/>||
|**serverURL**|`string`|ServerURL specifies the API endpoint<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmopenai"></a>
##### entConfig\.summarizer\.llm\.openai: object

OpenAIConfig contains OpenAI specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apiKey**|`string`|APIKey contains the authentication key for the service<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||
|**organizationID**|`string`|OrganizationID specifies the OpenAI organization ID<br/>||

**Additional Properties:** not allowed  
<a name="auth"></a>
## auth: object

Auth settings including oauth2 providers and token configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled authentication on the server, not recommended to disable<br/>|no|
|[**token**](#authtoken)|`object`||yes|
|[**supportedProviders**](#authsupportedproviders)|`string[]`||no|
|[**providers**](#authproviders)|`object`|OauthProviderConfig represents the configuration for OAuth providers such as Github and Google<br/>|no|

**Additional Properties:** not allowed  
<a name="authtoken"></a>
### auth\.token: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**kid**|`string`||yes|
|**audience**|`string`||yes|
|**refreshAudience**|`string`||no|
|**issuer**|`string`||yes|
|**accessDuration**|`integer`||no|
|**refreshDuration**|`integer`||no|
|**refreshOverlap**|`integer`||no|
|**jwksEndpoint**|`string`||no|
|[**keys**](#authtokenkeys)|`object`||yes|
|**generateKeys**|`boolean`||no|

**Additional Properties:** not allowed  
<a name="authtokenkeys"></a>
#### auth\.token\.keys: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="authsupportedproviders"></a>
### auth\.supportedProviders: array

**Items**

**Item Type:** `string`  
<a name="authproviders"></a>
### auth\.providers: object

OauthProviderConfig represents the configuration for OAuth providers such as Github and Google


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**redirectUrl**|`string`|RedirectURL is the URL that the OAuth2 client will redirect to after authentication is complete<br/>||
|[**github**](#authprovidersgithub)|`object`||yes|
|[**google**](#authprovidersgoogle)|`object`||yes|
|[**webauthn**](#authproviderswebauthn)|`object`||yes|

**Additional Properties:** not allowed  
<a name="authprovidersgithub"></a>
#### auth\.providers\.github: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientId**|`string`||yes|
|**clientSecret**|`string`||yes|
|**clientEndpoint**|`string`||no|
|[**scopes**](#authprovidersgithubscopes)|`string[]`||yes|
|**redirectUrl**|`string`||yes|

**Additional Properties:** not allowed  
<a name="authprovidersgithubscopes"></a>
##### auth\.providers\.github\.scopes: array

**Items**

**Item Type:** `string`  
<a name="authprovidersgoogle"></a>
#### auth\.providers\.google: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientId**|`string`||yes|
|**clientSecret**|`string`||yes|
|**clientEndpoint**|`string`||no|
|[**scopes**](#authprovidersgooglescopes)|`string[]`||yes|
|**redirectUrl**|`string`||yes|

**Additional Properties:** not allowed  
<a name="authprovidersgooglescopes"></a>
##### auth\.providers\.google\.scopes: array

**Items**

**Item Type:** `string`  
<a name="authproviderswebauthn"></a>
#### auth\.providers\.webauthn: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`||no|
|**displayName**|`string`||yes|
|**relyingPartyId**|`string`||yes|
|[**requestOrigins**](#authproviderswebauthnrequestorigins)|`string[]`||yes|
|**maxDevices**|`integer`||no|
|**enforceTimeout**|`boolean`||no|
|**timeout**|`integer`||no|
|**debug**|`boolean`||no|

**Additional Properties:** not allowed  
<a name="authproviderswebauthnrequestorigins"></a>
##### auth\.providers\.webauthn\.requestOrigins: array

**Items**

**Item Type:** `string`  
<a name="authz"></a>
## authz: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|enables authorization checks with openFGA<br/>|no|
|**storeName**|`string`|name of openFGA store<br/>|no|
|**hostUrl**|`string`|host url with scheme of the openFGA API<br/>|yes|
|**storeId**|`string`|id of openFGA store<br/>|no|
|**modelId**|`string`|id of openFGA model<br/>|no|
|**createNewModel**|`boolean`|force create a new model<br/>|no|
|**modelFile**|`string`|path to the fga model file<br/>|no|
|[**credentials**](#authzcredentials)|`object`||no|
|**ignoreDuplicateKeyError**|`boolean`|ignore duplicate key error<br/>|no|

**Additional Properties:** not allowed  
<a name="authzcredentials"></a>
### authz\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**apiToken**|`string`|api token for the openFGA client<br/>||
|**clientId**|`string`|client id for the openFGA client<br/>||
|**clientSecret**|`string`|client secret for the openFGA client<br/>||
|**audience**|`string`|audience for the openFGA client<br/>||
|**issuer**|`string`|issuer for the openFGA client<br/>||
|**scopes**|`string`|scopes for the openFGA client<br/>||

**Additional Properties:** not allowed  
<a name="db"></a>
## db: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**debug**|`boolean`|debug enables printing the debug database logs<br/>|no|
|**databaseName**|`string`|the name of the database to use with otel tracing<br/>|no|
|**driverName**|`string`|sql driver name<br/>|no|
|**multiWrite**|`boolean`|enables writing to two databases simultaneously<br/>|no|
|**primaryDbSource**|`string`|dsn of the primary database<br/>|yes|
|**secondaryDbSource**|`string`|dsn of the secondary database if multi-write is enabled<br/>|no|
|**cacheTTL**|`integer`|cache results for subsequent requests<br/>|no|
|**runMigrations**|`boolean`|run migrations on startup<br/>|no|
|**migrationProvider**|`string`|migration provider to use for running migrations<br/>|no|
|**enableHistory**|`boolean`|enable history data to be logged to the database<br/>|no|

**Additional Properties:** not allowed  
<a name="jobqueue"></a>
## jobQueue: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**connectionURI**|`string`|||
|**runMigrations**|`boolean`|||
|[**riverConf**](#jobqueueriverconf)|`object`|||

**Additional Properties:** not allowed  
<a name="jobqueueriverconf"></a>
### jobQueue\.riverConf: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**AdvisoryLockPrefix**|`integer`|||
|**CancelledJobRetentionPeriod**|`integer`|||
|**CompletedJobRetentionPeriod**|`integer`|||
|**DiscardedJobRetentionPeriod**|`integer`|||
|**ErrorHandler**||||
|**FetchCooldown**|`integer`|||
|**FetchPollInterval**|`integer`|||
|**ID**|`string`|||
|**JobCleanerTimeout**|`integer`|||
|[**JobInsertMiddleware**](#jobqueueriverconfjobinsertmiddleware)|`array`|||
|**JobTimeout**|`integer`|||
|[**Hooks**](#jobqueueriverconfhooks)|`array`|||
|[**Logger**](#jobqueueriverconflogger)|`object`|||
|**MaxAttempts**|`integer`|||
|[**Middleware**](#jobqueueriverconfmiddleware)|`array`|||
|[**PeriodicJobs**](#jobqueueriverconfperiodicjobs)|`array`|||
|**PollOnly**|`boolean`|||
|[**Queues**](#jobqueueriverconfqueues)|`object`|||
|**ReindexerSchedule**||||
|**ReindexerTimeout**|`integer`|||
|**RescueStuckJobsAfter**|`integer`|||
|**RetryPolicy**||||
|**Schema**|`string`|||
|**SkipJobKindValidation**|`boolean`|||
|**SkipUnknownJobCheck**|`boolean`|||
|[**Test**](#jobqueueriverconftest)|`object`|||
|**TestOnly**|`boolean`|||
|[**Workers**](#jobqueueriverconfworkers)|`object`|||
|[**WorkerMiddleware**](#jobqueueriverconfworkermiddleware)|`array`|||

**Additional Properties:** not allowed  
<a name="jobqueueriverconfjobinsertmiddleware"></a>
#### jobQueue\.riverConf\.JobInsertMiddleware: array

**Items**

<a name="jobqueueriverconfhooks"></a>
#### jobQueue\.riverConf\.Hooks: array

**Items**

<a name="jobqueueriverconflogger"></a>
#### jobQueue\.riverConf\.Logger: object

**No properties.**

**Additional Properties:** not allowed  
<a name="jobqueueriverconfmiddleware"></a>
#### jobQueue\.riverConf\.Middleware: array

**Items**

<a name="jobqueueriverconfperiodicjobs"></a>
#### jobQueue\.riverConf\.PeriodicJobs: array

**Items**

<a name="jobqueueriverconfqueues"></a>
#### jobQueue\.riverConf\.Queues: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#jobqueueriverconfqueuesadditionalproperties)|`object`|||

<a name="jobqueueriverconfqueuesadditionalproperties"></a>
##### jobQueue\.riverConf\.Queues\.additionalProperties: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**MaxWorkers**|`integer`|||

**Additional Properties:** not allowed  
<a name="jobqueueriverconftest"></a>
#### jobQueue\.riverConf\.Test: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**DisableUniqueEnforcement**|`boolean`|||
|**Time**||||

**Additional Properties:** not allowed  
<a name="jobqueueriverconfworkers"></a>
#### jobQueue\.riverConf\.Workers: object

**No properties.**

**Additional Properties:** not allowed  
<a name="jobqueueriverconfworkermiddleware"></a>
#### jobQueue\.riverConf\.WorkerMiddleware: array

**Items**

<a name="redis"></a>
## redis: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**address**|`string`|||
|**name**|`string`|||
|**username**|`string`|||
|**password**|`string`|||
|**db**|`integer`|||
|**dialTimeout**|`integer`|||
|**readTimeout**|`integer`|||
|**writeTimeout**|`integer`|||
|**maxRetries**|`integer`|||
|**minIdleConns**|`integer`|||
|**maxIdleConns**|`integer`|||
|**maxActiveConns**|`integer`|||

**Additional Properties:** not allowed  
<a name="tracer"></a>
## tracer: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**provider**|`string`|||
|**environment**|`string`|||
|[**stdout**](#tracerstdout)|`object`|||
|[**otlp**](#tracerotlp)|`object`|||

**Additional Properties:** not allowed  
<a name="tracerstdout"></a>
### tracer\.stdout: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**pretty**|`boolean`|||
|**disableTimestamp**|`boolean`|||

**Additional Properties:** not allowed  
<a name="tracerotlp"></a>
### tracer\.otlp: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**endpoint**|`string`|||
|**insecure**|`boolean`|||
|**certificate**|`string`|||
|[**headers**](#tracerotlpheaders)|`string[]`|||
|**compression**|`string`|||
|**timeout**|`integer`|||

**Additional Properties:** not allowed  
<a name="tracerotlpheaders"></a>
#### tracer\.otlp\.headers: array

**Items**

**Item Type:** `string`  
<a name="email"></a>
## email: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**companyName**|`string`|||
|**companyAddress**|`string`|||
|**corporation**|`string`|||
|**year**|`integer`|||
|**fromEmail**|`string`|||
|**supportEmail**|`string`|||
|**logoURL**|`string`|||
|[**urls**](#emailurls)|`object`|||
|**templatesPath**|`string`|||

**Additional Properties:** not allowed  
<a name="emailurls"></a>
### email\.urls: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**root**|`string`|||
|**product**|`string`|||
|**docs**|`string`|||
|**verify**|`string`|||
|**invite**|`string`|||
|**reset**|`string`|||
|**verifySubscriber**|`string`|||
|**verifyBilling**|`string`|||

**Additional Properties:** not allowed  
<a name="sessions"></a>
## sessions: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**signingKey**|`string`|||
|**encryptionKey**|`string`|||
|**domain**|`string`|||
|**maxAge**|`integer`|||

**Additional Properties:** not allowed  
<a name="totp"></a>
## totp: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**codeLength**|`integer`|||
|**issuer**|`string`|||
|**redis**|`boolean`|||
|**secret**|`string`|||
|**recoveryCodeCount**|`integer`|||
|**recoveryCodeLength**|`integer`|||

**Additional Properties:** not allowed  
<a name="ratelimit"></a>
## ratelimit: object

Config defines the configuration settings for the default rate limiter


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**limit**|`number`|||
|**burst**|`integer`|||
|**expires**|`integer`|||

**Additional Properties:** not allowed  
<a name="objectstorage"></a>
## objectStorage: object

Config is the configuration for the object store


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the store is enabled<br/>||
|**provider**|`string`|Provider is the name of the provider, eg. disk, s3, will default to disk if nothing is set<br/>||
|**accessKey**|`string`|AccessKey is the access key for the storage provider<br/>||
|**region**|`string`|Region is the region for the storage provider<br/>||
|**secretKey**|`string`|SecretKey is the secret key for the storage provider<br/>||
|**credentialsJSON**|`string`|CredentialsJSON is the credentials JSON for the storage provider<br/>||
|**defaultBucket**|`string`|DefaultBucket is the default bucket name for the storage provider, if not set, it will use the default<br/>this is the local path for disk storage or the bucket name for S3<br/>||
|**localURL**|`string`|LocalURL is the URL to use for the "presigned" URL for the file when using local storage<br/>e.g for local development, this can be http://localhost:17608/files/<br/>||
|[**keys**](#objectstoragekeys)|`string[]`|||
|**maxSizeMB**|`integer`|MaxUploadSizeMB is the maximum size of file uploads to accept in megabytes<br/>||
|**maxMemoryMB**|`integer`|MaxUploadMemoryMB is the maximum memory in megabytes to use when parsing a multipart form<br/>||

**Additional Properties:** not allowed  
<a name="objectstoragekeys"></a>
### objectStorage\.keys: array

**Items**

**Item Type:** `string`  
<a name="subscription"></a>
## subscription: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**publicStripeKey**|`string`|||
|**privateStripeKey**|`string`|||
|**stripeWebhookSecret**|`string`|||
|**trialSubscriptionPriceID**|`string`|||
|**personalOrgSubscriptionPriceID**|`string`|||
|**stripeWebhookURL**|`string`|||
|**stripeBillingPortalSuccessURL**|`string`|||
|**stripeCancellationReturnURL**|`string`|||
|[**saasPricingTiers**](#subscriptionsaaspricingtiers)|`array`|||
|[**features**](#subscriptionfeatures)|`array`|||

**Additional Properties:** not allowed  
<a name="subscriptionsaaspricingtiers"></a>
### subscription\.saasPricingTiers: array

**Items**

<a name="subscriptionfeatures"></a>
### subscription\.features: array

**Items**


