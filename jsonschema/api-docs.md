# object

Config contains the configuration for the datum server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**refreshInterval**|`integer`|RefreshInterval determines how often to reload the config<br/>||
|[**server**](#server)|`object`|Server settings for the echo server<br/>|yes|
|[**entConfig**](#entconfig)|`object`|Config holds the configuration for the ent server<br/>||
|[**auth**](#auth)|`object`|Auth settings including oauth2 providers and datum token configuration<br/>|yes|
|[**authz**](#authz)|`object`||yes|
|[**db**](#db)|`object`||yes|
|[**geodetic**](#geodetic)|`object`|||
|[**redis**](#redis)|`object`|Config for the redis client used to store key-value pairs<br/>||
|[**tracer**](#tracer)|`object`|Config defines the configuration settings for opentelemetry tracing<br/>||
|[**email**](#email)|`object`|Config for sending emails via SendGrid and managing marketing contacts<br/>||
|[**sessions**](#sessions)|`object`|Config contains the configuration for the session store<br/>||
|[**posthog**](#posthog)|`object`|Config is the configuration for PostHog<br/>||
|[**totp**](#totp)|`object`|||
|[**ratelimit**](#ratelimit)|`object`|Config defines the configuration settings for the default rate limiter<br/>||
|[**publisherConfig**](#publisherconfig)|`object`|Config is the configuration for the Kafka event source<br/>||

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
|**maxCapacity**|`integer`|MaxCapacity is the maximum number of tasks that can be queued<br/>||

**Additional Properties:** not allowed  
<a name="entconfig"></a>
## entConfig: object

Config holds the configuration for the ent server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**flags**](#entconfigflags)|`object`|Flags contains the flags for the server to allow use to test different code paths<br/>||
|[**entityTypes**](#entconfigentitytypes)|`string[]`|||

**Additional Properties:** not allowed  
<a name="entconfigflags"></a>
### entConfig\.flags: object

Flags contains the flags for the server to allow use to test different code paths


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**useListUserService**|`boolean`|use list services endpoint for object access<br/>||
|**useListObjectServices**|`boolean`|use list object services endpoint for object access<br/>||

**Additional Properties:** not allowed  
<a name="entconfigentitytypes"></a>
### entConfig\.entityTypes: array

**Items**

**Item Type:** `string`  
<a name="auth"></a>
## auth: object

Auth settings including oauth2 providers and datum token configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled authentication on the server, not recommended to disable<br/>|no|
|[**token**](#authtoken)|`object`|Config defines the configuration settings for authentication tokens used in the server<br/>|yes|
|[**supportedProviders**](#authsupportedproviders)|`string[]`||no|
|[**providers**](#authproviders)|`object`|OauthProviderConfig represents the configuration for OAuth providers such as Github and Google<br/>|no|

**Additional Properties:** not allowed  
<a name="authtoken"></a>
### auth\.token: object

Config defines the configuration settings for authentication tokens used in the server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**kid**|`string`|KID represents the Key ID used in the configuration.<br/>|yes|
|**audience**|`string`|Audience represents the target audience for the tokens.<br/>|yes|
|**refreshAudience**|`string`|RefreshAudience represents the audience for refreshing tokens.<br/>|no|
|**issuer**|`string`|Issuer represents the issuer of the tokens<br/>|yes|
|**accessDuration**|`integer`|AccessDuration represents the duration of the access token is valid for<br/>|no|
|**refreshDuration**|`integer`|RefreshDuration represents the duration of the refresh token is valid for<br/>|no|
|**refreshOverlap**|`integer`|RefreshOverlap represents the overlap time for a refresh and access token<br/>|no|
|**jwksEndpoint**|`string`|JWKSEndpoint represents the endpoint for the JSON Web Key Set<br/>|no|
|[**keys**](#authtokenkeys)|`object`||yes|
|**generateKeys**|`boolean`|GenerateKeys is a boolean to determine if the keys should be generated<br/>|no|

**Additional Properties:** not allowed  
<a name="authtokenkeys"></a>
#### auth\.token\.keys: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|

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
|**redirectUrl**|`string`|RedirectURL is the URL that the OAuth2 client will redirect to after authentication with datum<br/>||
|[**github**](#authprovidersgithub)|`object`|ProviderConfig represents the configuration settings for a Github Oauth Provider<br/>|yes|
|[**google**](#authprovidersgoogle)|`object`|ProviderConfig represents the configuration settings for a Google Oauth Provider<br/>|yes|
|[**webauthn**](#authproviderswebauthn)|`object`|ProviderConfig represents the configuration settings for a Webauthn Provider<br/>|yes|

**Additional Properties:** not allowed  
<a name="authprovidersgithub"></a>
#### auth\.providers\.github: object

ProviderConfig represents the configuration settings for a Github Oauth Provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientId**|`string`|ClientID is the public identifier for the GitHub oauth2 client<br/>|yes|
|**clientSecret**|`string`|ClientSecret is the secret for the GitHub oauth2 client<br/>|yes|
|**clientEndpoint**|`string`|ClientEndpoint is the endpoint for the GitHub oauth2 client<br/>|no|
|[**scopes**](#authprovidersgithubscopes)|`string[]`||yes|
|**redirectUrl**|`string`|RedirectURL is the URL that the GitHub oauth2 client will redirect to after authentication with Github<br/>|yes|

**Additional Properties:** not allowed  
<a name="authprovidersgithubscopes"></a>
##### auth\.providers\.github\.scopes: array

**Items**

**Item Type:** `string`  
<a name="authprovidersgoogle"></a>
#### auth\.providers\.google: object

ProviderConfig represents the configuration settings for a Google Oauth Provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientId**|`string`|ClientID is the public identifier for the Google oauth2 client<br/>|yes|
|**clientSecret**|`string`|ClientSecret is the secret for the Google oauth2 client<br/>|yes|
|**clientEndpoint**|`string`|ClientEndpoint is the endpoint for the Google oauth2 client<br/>|no|
|[**scopes**](#authprovidersgooglescopes)|`string[]`||yes|
|**redirectUrl**|`string`|RedirectURL is the URL that the Google oauth2 client will redirect to after authentication with Google<br/>|yes|

**Additional Properties:** not allowed  
<a name="authprovidersgooglescopes"></a>
##### auth\.providers\.google\.scopes: array

**Items**

**Item Type:** `string`  
<a name="authproviderswebauthn"></a>
#### auth\.providers\.webauthn: object

ProviderConfig represents the configuration settings for a Webauthn Provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled is the provider enabled<br/>|no|
|**displayName**|`string`|DisplayName is the site display name<br/>|yes|
|**relyingPartyId**|`string`|RelyingPartyID is the relying party identifier<br/>set to localhost for development, no port<br/>|yes|
|[**requestOrigins**](#authproviderswebauthnrequestorigins)|`string[]`||yes|
|**maxDevices**|`integer`|MaxDevices is the maximum number of devices that can be associated with a user<br/>|no|
|**enforceTimeout**|`boolean`|EnforceTimeout at the Relying Party / Server. This means if enabled and the user takes too long that even if the browser does not<br/>enforce a timeout, the server will<br/>|no|
|**timeout**|`integer`|Timeout is the timeout in seconds<br/>|no|
|**debug**|`boolean`|Debug enables debug mode<br/>|no|

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
<a name="geodetic"></a>
## geodetic: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enable the geodetic client<br/>||
|**baseUrl**|`string`|Base URL for the geodetic service<br/>||
|**endpoint**|`string`|Endpoint for the graphql api<br/>||
|**debug**|`boolean`|Enable debug mode<br/>||

**Additional Properties:** not allowed  
<a name="redis"></a>
## redis: object

Config for the redis client used to store key-value pairs


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled to enable redis client in the server<br/>||
|**address**|`string`|Address is the host:port to connect to redis<br/>||
|**name**|`string`|Name of the connecting client<br/>||
|**username**|`string`|Username to connect to redis<br/>||
|**password**|`string`|Password, must match the password specified in the server configuration<br/>||
|**db**|`integer`|DB to be selected after connecting to the server, 0 uses the default<br/>||
|**dialTimeout**|`integer`|Dial timeout for establishing new connections, defaults to 5s<br/>||
|**readTimeout**|`integer`|Timeout for socket reads. If reached, commands will fail<br/>with a timeout instead of blocking. Supported values:<br/>  - `0` - default timeout (3 seconds).<br/>  - `-1` - no timeout (block indefinitely).<br/>  - `-2` - disables SetReadDeadline calls completely.<br/>||
|**writeTimeout**|`integer`|Timeout for socket writes. If reached, commands will fail<br/>with a timeout instead of blocking.  Supported values:<br/>  - `0` - default timeout (3 seconds).<br/>  - `-1` - no timeout (block indefinitely).<br/>  - `-2` - disables SetWriteDeadline calls completely.<br/>||
|**maxRetries**|`integer`|MaxRetries before giving up.<br/>Default is 3 retries; -1 (not 0) disables retries.<br/>||
|**minIdleConns**|`integer`|MinIdleConns is useful when establishing new connection is slow.<br/>Default is 0. the idle connections are not closed by default.<br/>||
|**maxIdleConns**|`integer`|Maximum number of idle connections.<br/>Default is 0. the idle connections are not closed by default.<br/>||
|**maxActiveConns**|`integer`|Maximum number of connections allocated by the pool at a given time.<br/>When zero, there is no limit on the number of connections in the pool.<br/>||

**Additional Properties:** not allowed  
<a name="tracer"></a>
## tracer: object

Config defines the configuration settings for opentelemetry tracing


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled to enable tracing<br/>||
|**provider**|`string`|Provider to use for tracing<br/>||
|**environment**|`string`|Environment to set for the service<br/>||
|[**stdout**](#tracerstdout)|`object`|StdOut settings for the stdout provider<br/>||
|[**otlp**](#tracerotlp)|`object`|OTLP settings for the otlp provider<br/>||

**Additional Properties:** not allowed  
<a name="tracerstdout"></a>
### tracer\.stdout: object

StdOut settings for the stdout provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**pretty**|`boolean`|Pretty enables pretty printing of the output<br/>||
|**disableTimestamp**|`boolean`|DisableTimestamp disables the timestamp in the output<br/>||

**Additional Properties:** not allowed  
<a name="tracerotlp"></a>
### tracer\.otlp: object

OTLP settings for the otlp provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**endpoint**|`string`|Endpoint to send the traces to<br/>||
|**insecure**|`boolean`|Insecure to disable TLS<br/>||
|**certificate**|`string`|Certificate to use for TLS<br/>||
|[**headers**](#tracerotlpheaders)|`string[]`|||
|**compression**|`string`|Compression to use for the request<br/>||
|**timeout**|`integer`|Timeout for the request<br/>||

**Additional Properties:** not allowed  
<a name="tracerotlpheaders"></a>
#### tracer\.otlp\.headers: array

**Items**

**Item Type:** `string`  
<a name="email"></a>
## email: object

Config for sending emails via SendGrid and managing marketing contacts


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**sendGridApiKey**|`string`|SendGridAPIKey is the SendGrid API key to authenticate with the service<br/>||
|**fromEmail**|`string`|FromEmail is the default email to send from<br/>||
|**testing**|`boolean`|Testing is a bool flag to indicate we shouldn't be sending live emails, and instead should be writing out fixtures<br/>||
|**archive**|`string`|Archive is only supported in testing mode and is what is tied through the mock to write out fixtures<br/>||
|**datumListId**|`string`|DatumListID is the UUID SendGrid spits out when you create marketing lists<br/>||
|**adminEmail**|`string`|AdminEmail is an internal group email configured within datum for email testing and visibility<br/>||
|[**consoleUrl**](#emailconsoleurl)|`object`|ConsoleURLConfig for the datum registration<br/>||
|[**marketingUrl**](#emailmarketingurl)|`object`|MarketingURLConfig for the datum marketing emails<br/>||

**Additional Properties:** not allowed  
<a name="emailconsoleurl"></a>
### email\.consoleUrl: object

ConsoleURLConfig for the datum registration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleBase**|`string`|ConsoleBase is the base URL used for URL links in emails<br/>||
|**verify**|`string`|Verify is the path to the verify endpoint used in verification emails<br/>||
|**invite**|`string`|Invite is the path to the invite endpoint used in invite emails<br/>||
|**reset**|`string`|Reset is the path to the reset endpoint used in password reset emails<br/>||

**Additional Properties:** not allowed  
<a name="emailmarketingurl"></a>
### email\.marketingUrl: object

MarketingURLConfig for the datum marketing emails


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**marketingBase**|`string`|MarketingBase is the base URL used for marketing links in emails<br/>||
|**subscriberVerify**|`string`|SubscriberVerify is the path to the subscriber verify endpoint used in verification emails<br/>||

**Additional Properties:** not allowed  
<a name="sessions"></a>
## sessions: object

Config contains the configuration for the session store


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**signingKey**|`string`|SigningKey must be a 16, 32, or 64 character string used to encode the cookie<br/>||
|**encryptionKey**|`string`|EncryptionKey must be a 16, 32, or 64 character string used to encode the cookie<br/>||
|**domain**|`string`|Domain is the domain for the cookie, leave empty to use the default value of the server<br/>||

**Additional Properties:** not allowed  
<a name="posthog"></a>
## posthog: object

Config is the configuration for PostHog


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled is a flag to enable or disable PostHog<br/>||
|**apiKey**|`string`|APIKey is the PostHog API Key<br/>||
|**host**|`string`|Host is the PostHog API Host<br/>||

**Additional Properties:** not allowed  
<a name="totp"></a>
## totp: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled is a flag to enable or disable the OTP service<br/>||
|**codeLength**|`integer`|CodeLength is the length of the OTP code<br/>||
|**issuer**|`string`|Issuer is the issuer for TOTP codes<br/>||
|**redis**|`boolean`|WithRedis configures the service with a redis client<br/>||
|**secret**|`string`|Secret stores a versioned secret key for cryptography functions<br/>||
|**recoveryCodeCount**|`integer`|RecoveryCodeCount is the number of recovery codes to generate<br/>||
|**recoveryCodeLength**|`integer`|RecoveryCodeLength is the length of a recovery code<br/>||

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
<a name="publisherconfig"></a>
## publisherConfig: object

Config is the configuration for the Kafka event source


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled is a flag to determine if the Kafka event source is enabled<br/>||
|**appName**|`string`|AppName is the name of the application that is publishing events<br/>||
|**address**|`string`|Address is the address of the Kafka broker<br/>||
|[**addresses**](#publisherconfigaddresses)|`string[]`|||
|**debug**|`boolean`|Debug is a flag to determine if the Kafka client should run in debug mode<br/>||

**Additional Properties:** not allowed  
<a name="publisherconfigaddresses"></a>
### publisherConfig\.addresses: array

**Items**

**Item Type:** `string`  

