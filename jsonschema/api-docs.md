# object

Config contains the configuration for the core server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**domain**|`string`|Domain provides a global domain value for other modules to inherit<br/>||
|**refreshinterval**|`integer`|RefreshInterval determines how often to reload the config<br/>||
|[**server**](#server)|`object`|Server settings for the echo server<br/>|yes|
|[**entconfig**](#entconfig)|`object`|Config holds the configuration for the ent server<br/>||
|[**auth**](#auth)|`object`|Auth settings including oauth2 providers and token configuration<br/>|yes|
|[**authz**](#authz)|`object`||yes|
|[**db**](#db)|`object`||yes|
|[**jobqueue**](#jobqueue)|`object`|||
|[**redis**](#redis)|`object`|||
|[**tracer**](#tracer)|`object`|||
|[**email**](#email)|`object`|||
|[**sessions**](#sessions)|`object`|||
|[**totp**](#totp)|`object`|||
|[**ratelimit**](#ratelimit)|`object`|Config defines the configuration settings for the rate limiter middleware.<br/>||
|[**objectstorage**](#objectstorage)|`object`|ProviderConfig contains configuration for object storage providers<br/>||
|[**subscription**](#subscription)|`object`|||
|[**keywatcher**](#keywatcher)|`object`|KeyWatcher contains settings for the key watcher that manages JWT signing keys<br/>||
|[**slack**](#slack)|`object`|Slack contains settings for Slack notifications<br/>||
|[**integrationoauthprovider**](#integrationoauthprovider)|`object`|IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations.<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "server": {
        "tls": {},
        "cors": {
            "prefixes": {}
        },
        "secure": {},
        "redirects": {
            "redirects": {}
        },
        "cachecontrol": {
            "nocacheheaders": {}
        },
        "mime": {},
        "graphpool": {},
        "csrfprotection": {},
        "fieldlevelencryption": {}
    },
    "entconfig": {
        "summarizer": {
            "llm": {
                "anthropic": {},
                "cloudflare": {},
                "openai": {}
            }
        },
        "windmill": {},
        "modules": {},
        "emailvalidation": {
            "allowedemailtypes": {}
        },
        "billing": {},
        "notifications": {}
    },
    "auth": {
        "token": {
            "keys": {},
            "redis": {
                "config": {}
            },
            "apitokens": {
                "keys": {}
            }
        },
        "providers": {
            "github": {},
            "google": {},
            "webauthn": {}
        }
    },
    "authz": {
        "credentials": {}
    },
    "db": {},
    "jobqueue": {
        "riverconf": {
            "Logger": {},
            "PeriodicJobs": [
                {}
            ],
            "Queues": {},
            "Test": {},
            "Workers": {}
        },
        "metrics": {}
    },
    "redis": {},
    "tracer": {
        "stdout": {},
        "otlp": {}
    },
    "email": {
        "urls": {}
    },
    "sessions": {},
    "totp": {},
    "ratelimit": {
        "options": [
            {}
        ]
    },
    "objectstorage": {
        "providers": {
            "s3": {
                "credentials": {}
            },
            "r2": {
                "credentials": {}
            },
            "disk": {
                "credentials": {}
            },
            "database": {
                "credentials": {}
            }
        }
    },
    "subscription": {
        "stripewebhooksecrets": {}
    },
    "keywatcher": {},
    "slack": {},
    "integrationoauthprovider": {}
}
```

<a name="server"></a>
## server: object

Server settings for the echo server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**dev**|`boolean`|Dev enables echo's dev mode options<br/>|no|
|**listen**|`string`|Listen sets the listen address to serve the echo server on<br/>|yes|
|**metricsport**|`string`|MetricsPort sets the port for the metrics endpoint<br/>|no|
|**shutdowngraceperiod**|`integer`|ShutdownGracePeriod sets the grace period for in flight requests before shutting down<br/>|no|
|**readtimeout**|`integer`|ReadTimeout sets the maximum duration for reading the entire request including the body<br/>|no|
|**writetimeout**|`integer`|WriteTimeout sets the maximum duration before timing out writes of the response<br/>|no|
|**idletimeout**|`integer`|IdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled<br/>|no|
|**readheadertimeout**|`integer`|ReadHeaderTimeout sets the amount of time allowed to read request headers<br/>|no|
|[**tls**](#servertls)|`object`|TLS settings for the server for secure connections<br/>|no|
|[**cors**](#servercors)|`object`|Config holds the cors configuration settings<br/>|no|
|[**secure**](#serversecure)|`object`|Config contains the types used in the mw middleware<br/>|no|
|[**redirects**](#serverredirects)|`object`|Config contains the types used in executing redirects via the redirect middleware<br/>|no|
|[**cachecontrol**](#servercachecontrol)|`object`|Config is the config values for the cache-control middleware<br/>|no|
|[**mime**](#servermime)|`object`|Config defines the config for Mime middleware<br/>|no|
|[**graphpool**](#servergraphpool)|`object`|PondPool contains the settings for the goroutine pool<br/>|no|
|**enablegraphextensions**|`boolean`|EnableGraphExtensions enables the graph extensions for the graph resolvers<br/>|no|
|**enablegraphsubscriptions**|`boolean`|EnableGraphSubscriptions enables graphql subscriptions to the server using websockets or sse<br/>|no|
|**complexitylimit**|`integer`|ComplexityLimit sets the maximum complexity allowed for a query<br/>|no|
|**maxresultlimit**|`integer`|MaxResultLimit sets the maximum number of results allowed for a query<br/>|no|
|[**csrfprotection**](#servercsrfprotection)|`object`|Config defines configuration for the CSRF middleware wrapper.<br/>|no|
|**secretmanager**|`string`|SecretManagerSecret is the name of the GCP Secret Manager secret containing the JWT signing key<br/>|no|
|**defaulttrustcenterdomain**|`string`|DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set<br/>|no|
|[**fieldlevelencryption**](#serverfieldlevelencryption)|`object`||no|
|**trustcentercnametarget**|`string`|TrustCenterCnameTarget is the cname target for the trust center<br/>Used for mapping the vanity domains to the trust centers<br/>|no|
|**trustcenterpreviewzoneid**|`string`|TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain<br/>|no|

**Additional Properties:** not allowed  
**Example**

```json
{
    "tls": {},
    "cors": {
        "prefixes": {}
    },
    "secure": {},
    "redirects": {
        "redirects": {}
    },
    "cachecontrol": {
        "nocacheheaders": {}
    },
    "mime": {},
    "graphpool": {},
    "csrfprotection": {},
    "fieldlevelencryption": {}
}
```

<a name="servertls"></a>
### server\.tls: object

TLS settings for the server for secure connections


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled turns on TLS settings for the server<br/>||
|**certfile**|`string`|CertFile location for the TLS server<br/>||
|**certkey**|`string`|CertKey file location for the TLS server<br/>||
|**autocert**|`boolean`|AutoCert generates the cert with letsencrypt, this does not work on localhost<br/>||

**Additional Properties:** not allowed  
<a name="servercors"></a>
### server\.cors: object

Config holds the cors configuration settings


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enable or disable the CORS middleware<br/>||
|[**prefixes**](#servercorsprefixes)|`object`|||
|[**alloworigins**](#servercorsalloworigins)|`string[]`|||
|**cookieinsecure**|`boolean`|CookieInsecure sets the cookie to be insecure<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "prefixes": {}
}
```

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
#### server\.cors\.alloworigins: array

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
**Example**

```json
{
    "redirects": {}
}
```

<a name="serverredirectsredirects"></a>
#### server\.redirects\.redirects: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="servercachecontrol"></a>
### server\.cachecontrol: object

Config is the config values for the cache-control middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**nocacheheaders**](#servercachecontrolnocacheheaders)|`object`|||
|[**etagheaders**](#servercachecontroletagheaders)|`string[]`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "nocacheheaders": {}
}
```

<a name="servercachecontrolnocacheheaders"></a>
#### server\.cachecontrol\.nocacheheaders: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="servercachecontroletagheaders"></a>
#### server\.cachecontrol\.etagheaders: array

**Items**

**Item Type:** `string`  
<a name="servermime"></a>
### server\.mime: object

Config defines the config for Mime middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the mime middleware should be enabled<br/>||
|**mimetypesfile**|`string`|MimeTypesFile is the file to load mime types from<br/>||
|**defaultcontenttype**|`string`|DefaultContentType is the default content type to set if no mime type is found<br/>||

**Additional Properties:** not allowed  
<a name="servergraphpool"></a>
### server\.graphpool: object

PondPool contains the settings for the goroutine pool


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**maxworkers**|`integer`|MaxWorkers is the maximum number of workers in the pool<br/>||

**Additional Properties:** not allowed  
<a name="servercsrfprotection"></a>
### server\.csrfprotection: object

Config defines configuration for the CSRF middleware wrapper.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether CSRF protection is enabled.<br/>||
|**header**|`string`|Header specifies the header name to look for the CSRF token.<br/>||
|**cookie**|`string`|Cookie specifies the cookie name used to store the CSRF token.<br/>||
|**secure**|`boolean`|Secure sets the Secure flag on the CSRF cookie.<br/>||
|**samesite**|`string`|SameSite configures the SameSite attribute on the CSRF cookie. Valid<br/>values are "Lax", "Strict", "None" and "Default".<br/>||
|**cookiehttponly**|`boolean`|CookieHTTPOnly indicates whether the CSRF cookie is HTTP only.<br/>||
|**cookiedomain**|`string`|CookieDomain specifies the domain for the CSRF cookie, default to no domain<br/>||
|**cookiepath**|`string`|CookiePath specifies the path for the CSRF cookie, default to "/"<br/>||

**Additional Properties:** not allowed  
<a name="serverfieldlevelencryption"></a>
### server\.fieldlevelencryption: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether Tink encryption is enabled<br/>||
|**keyset**|`string`|Keyset is the base64-encoded Tink keyset used for encryption<br/>||

**Additional Properties:** not allowed  
<a name="entconfig"></a>
## entconfig: object

Config holds the configuration for the ent server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**entitytypes**](#entconfigentitytypes)|`string[]`|||
|[**summarizer**](#entconfigsummarizer)|`object`|Config holds configuration for the text summarization functionality<br/>||
|[**windmill**](#entconfigwindmill)|`object`|Windmill holds configuration for the Windmill workflow automation platform<br/>||
|**maxpoolsize**|`integer`|MaxPoolSize is the max pond pool workers that can be used by the ent client<br/>||
|[**modules**](#entconfigmodules)|`object`|Modules settings for features access<br/>||
|**maxschemaimportsize**|`integer`|MaxSchemaImportSize is the maximum size allowed for schema imports in bytes<br/>||
|[**emailvalidation**](#entconfigemailvalidation)|`object`|EmailVerificationConfig is the configuration for email verification<br/>||
|[**billing**](#entconfigbilling)|`object`|Billing settings for feature access<br/>||
|[**notifications**](#entconfignotifications)|`object`|Notifications settings for notifications sent to users based on events<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "summarizer": {
        "llm": {
            "anthropic": {},
            "cloudflare": {},
            "openai": {}
        }
    },
    "windmill": {},
    "modules": {},
    "emailvalidation": {
        "allowedemailtypes": {}
    },
    "billing": {},
    "notifications": {}
}
```

<a name="entconfigentitytypes"></a>
### entconfig\.entitytypes: array

**Items**

**Item Type:** `string`  
<a name="entconfigsummarizer"></a>
### entconfig\.summarizer: object

Config holds configuration for the text summarization functionality


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**type**|`string`|Type specifies the summarization algorithm to use<br/>||
|[**llm**](#entconfigsummarizerllm)|`object`|LLM contains configuration for multiple LLM providers<br/>||
|**maximumsentences**|`integer`|MaximumSentences specifies the maximum number of sentences in the summary<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "llm": {
        "anthropic": {},
        "cloudflare": {},
        "openai": {}
    }
}
```

<a name="entconfigsummarizerllm"></a>
#### entconfig\.summarizer\.llm: object

LLM contains configuration for multiple LLM providers


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**provider**|`string`|Provider specifies which LLM service to use<br/>||
|[**anthropic**](#entconfigsummarizerllmanthropic)|`object`|AnthropicConfig contains Anthropic specific configuration<br/>||
|[**cloudflare**](#entconfigsummarizerllmcloudflare)|`object`|CloudflareConfig contains Cloudflare specific configuration<br/>||
|[**openai**](#entconfigsummarizerllmopenai)|`object`|OpenAIConfig contains OpenAI specific configuration<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "anthropic": {},
    "cloudflare": {},
    "openai": {}
}
```

<a name="entconfigsummarizerllmanthropic"></a>
##### entconfig\.summarizer\.llm\.anthropic: object

AnthropicConfig contains Anthropic specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**betaheader**|`string`|BetaHeader specifies the beta API features to enable<br/>||
|**legacytextcompletion**|`boolean`|LegacyTextCompletion enables legacy text completion API<br/>||
|**baseurl**|`string`|BaseURL specifies the API endpoint<br/>||
|**model**|`string`|Model specifies the model name to use<br/>||
|**apikey**|`string`|APIKey contains the authentication key for the service<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmcloudflare"></a>
##### entconfig\.summarizer\.llm\.cloudflare: object

CloudflareConfig contains Cloudflare specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apikey**|`string`|APIKey contains the authentication key for the service<br/>||
|**accountid**|`string`|AccountID specifies the Cloudflare account ID<br/>||
|**serverurl**|`string`|ServerURL specifies the API endpoint<br/>||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmopenai"></a>
##### entconfig\.summarizer\.llm\.openai: object

OpenAIConfig contains OpenAI specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apikey**|`string`|APIKey contains the authentication key for the service<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||
|**organizationid**|`string`|OrganizationID specifies the OpenAI organization ID<br/>||

**Additional Properties:** not allowed  
<a name="entconfigwindmill"></a>
### entconfig\.windmill: object

Windmill holds configuration for the Windmill workflow automation platform


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled specifies whether Windmill integration is enabled<br/>||
|**baseurl**|`string`|BaseURL is the base URL of the Windmill instance<br/>||
|**workspace**|`string`|Workspace is the Windmill workspace to use<br/>||
|**token**|`string`|Token is the API token for authentication with Windmill<br/>||
|**defaulttimeout**|`string`|DefaultTimeout is the default timeout for API requests<br/>||
|**timezone**|`string`|Timezone for scheduled jobs<br/>||
|**onfailurescript**|`string`|OnFailureScript script to run when a scheduled job fails<br/>||
|**onsuccessscript**|`string`|OnSuccessScript script to run when a scheduled job succeeds<br/>||

**Additional Properties:** not allowed  
<a name="entconfigmodules"></a>
### entconfig\.modules: object

Modules settings for features access


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether to check and verify module access<br/>||
|**usesandbox**|`boolean`|UseSandbox indicates whether to use the sandbox catalog for module access checks<br/>||

**Additional Properties:** not allowed  
<a name="entconfigemailvalidation"></a>
### entconfig\.emailvalidation: object

EmailVerificationConfig is the configuration for email verification


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether email verification is enabled<br/>||
|**enableautoupdatedisposable**|`boolean`|EnableAutoUpdateDisposable indicates whether to automatically update disposable email addresses<br/>||
|**enablegravatarcheck**|`boolean`|EnableGravatarCheck indicates whether to check for Gravatar existence<br/>||
|**enablesmtpcheck**|`boolean`|EnableSMTPCheck indicates whether to check email by smtp<br/>||
|[**allowedemailtypes**](#entconfigemailvalidationallowedemailtypes)|`object`|AllowedEmailTypes defines the allowed email types for verification<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "allowedemailtypes": {}
}
```

<a name="entconfigemailvalidationallowedemailtypes"></a>
#### entconfig\.emailvalidation\.allowedemailtypes: object

AllowedEmailTypes defines the allowed email types for verification


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**disposable**|`boolean`|Disposable indicates whether disposable email addresses are allowed<br/>||
|**free**|`boolean`|Free indicates whether free email addresses are allowed<br/>||
|**role**|`boolean`|Role indicates whether role-based email addresses are allowed<br/>||

**Additional Properties:** not allowed  
<a name="entconfigbilling"></a>
### entconfig\.billing: object

Billing settings for feature access


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**requirepaymentmethod**|`boolean`|RequirePaymentMethod indicates whether to check if a payment method<br/>exists for orgs before they can access some resource<br/>||
|[**bypassemaildomains**](#entconfigbillingbypassemaildomains)|`string[]`|||

**Additional Properties:** not allowed  
<a name="entconfigbillingbypassemaildomains"></a>
#### entconfig\.billing\.bypassemaildomains: array

**Items**

**Item Type:** `string`  
<a name="entconfignotifications"></a>
### entconfig\.notifications: object

Notifications settings for notifications sent to users based on events


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleurl**|`string`|ConsoleURL for ui links used in notifications<br/>||

**Additional Properties:** not allowed  
<a name="auth"></a>
## auth: object

Auth settings including oauth2 providers and token configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled authentication on the server, not recommended to disable<br/>|no|
|[**token**](#authtoken)|`object`||yes|
|[**supportedproviders**](#authsupportedproviders)|`string[]`||no|
|[**providers**](#authproviders)|`object`|OauthProviderConfig represents the configuration for OAuth providers such as Github and Google<br/>|no|

**Additional Properties:** not allowed  
**Example**

```json
{
    "token": {
        "keys": {},
        "redis": {
            "config": {}
        },
        "apitokens": {
            "keys": {}
        }
    },
    "providers": {
        "github": {},
        "google": {},
        "webauthn": {}
    }
}
```

<a name="authtoken"></a>
### auth\.token: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**kid**|`string`||yes|
|**audience**|`string`||yes|
|**refreshaudience**|`string`||no|
|**issuer**|`string`||yes|
|**accessduration**|`integer`||no|
|**refreshduration**|`integer`||no|
|**refreshoverlap**|`integer`||no|
|**jwksendpoint**|`string`||no|
|[**keys**](#authtokenkeys)|`object`||yes|
|**generatekeys**|`boolean`||no|
|**jwkscachettl**|`integer`||no|
|[**redis**](#authtokenredis)|`object`||no|
|[**apitokens**](#authtokenapitokens)|`object`||no|
|**assessmentaccessduration**|`integer`||no|

**Additional Properties:** not allowed  
**Example**

```json
{
    "keys": {},
    "redis": {
        "config": {}
    },
    "apitokens": {
        "keys": {}
    }
}
```

<a name="authtokenkeys"></a>
#### auth\.token\.keys: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="authtokenredis"></a>
#### auth\.token\.redis: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**config**](#authtokenredisconfig)|`object`|||
|**blacklistprefix**|`string`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "config": {}
}
```

<a name="authtokenredisconfig"></a>
##### auth\.token\.redis\.config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**address**|`string`|||
|**name**|`string`|||
|**username**|`string`|||
|**password**|`string`|||
|**db**|`integer`|||
|**dialtimeout**|`integer`|||
|**readtimeout**|`integer`|||
|**writetimeout**|`integer`|||
|**maxretries**|`integer`|||
|**minidleconns**|`integer`|||
|**maxidleconns**|`integer`|||
|**maxactiveconns**|`integer`|||

**Additional Properties:** not allowed  
<a name="authtokenapitokens"></a>
#### auth\.token\.apitokens: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**envprefix**|`string`|||
|[**keys**](#authtokenapitokenskeys)|`object`|||
|**secretsize**|`integer`|||
|**delimiter**|`string`|||
|**prefix**|`string`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "keys": {}
}
```

<a name="authtokenapitokenskeys"></a>
##### auth\.token\.apitokens\.keys: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#authtokenapitokenskeysadditionalproperties)|`object`|||

<a name="authtokenapitokenskeysadditionalproperties"></a>
###### auth\.token\.apitokens\.keys\.additionalProperties: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**secret**|`string`|||
|**status**|`string`|||

**Additional Properties:** not allowed  
<a name="authsupportedproviders"></a>
### auth\.supportedproviders: array

**Items**

**Item Type:** `string`  
<a name="authproviders"></a>
### auth\.providers: object

OauthProviderConfig represents the configuration for OAuth providers such as Github and Google


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**redirecturl**|`string`|RedirectURL is the URL that the OAuth2 client will redirect to after authentication is complete<br/>||
|[**github**](#authprovidersgithub)|`object`||yes|
|[**google**](#authprovidersgoogle)|`object`||yes|
|[**webauthn**](#authproviderswebauthn)|`object`||yes|

**Additional Properties:** not allowed  
**Example**

```json
{
    "github": {},
    "google": {},
    "webauthn": {}
}
```

<a name="authprovidersgithub"></a>
#### auth\.providers\.github: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`||yes|
|**clientsecret**|`string`||yes|
|**clientendpoint**|`string`||no|
|[**scopes**](#authprovidersgithubscopes)|`string[]`||yes|
|**redirecturl**|`string`||yes|

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
|**clientid**|`string`||yes|
|**clientsecret**|`string`||yes|
|**clientendpoint**|`string`||no|
|[**scopes**](#authprovidersgooglescopes)|`string[]`||yes|
|**redirecturl**|`string`||yes|

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
|**displayname**|`string`||yes|
|**relyingpartyid**|`string`||yes|
|[**requestorigins**](#authproviderswebauthnrequestorigins)|`string[]`||yes|
|**maxdevices**|`integer`||no|
|**enforcetimeout**|`boolean`||no|
|**timeout**|`integer`||no|
|**debug**|`boolean`||no|

**Additional Properties:** not allowed  
<a name="authproviderswebauthnrequestorigins"></a>
##### auth\.providers\.webauthn\.requestorigins: array

**Items**

**Item Type:** `string`  
<a name="authz"></a>
## authz: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|enables authorization checks with openFGA<br/>|no|
|**storename**|`string`|name of openFGA store<br/>|no|
|**hosturl**|`string`|host url with scheme of the openFGA API<br/>|yes|
|**storeid**|`string`|id of openFGA store<br/>|no|
|**modelid**|`string`|id of openFGA model<br/>|no|
|**createnewmodel**|`boolean`|force create a new model<br/>|no|
|**modelfile**|`string`|path to the fga model file<br/>|no|
|[**credentials**](#authzcredentials)|`object`||no|
|**maxbatchwritesize**|`integer`|maximum number of writes per batch in a transaction<br/>|no|

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="authzcredentials"></a>
### authz\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**apitoken**|`string`|api token for the openFGA client<br/>||
|**clientid**|`string`|client id for the openFGA client<br/>||
|**clientsecret**|`string`|client secret for the openFGA client<br/>||
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
|**databasename**|`string`|the name of the database to use with otel tracing<br/>|no|
|**drivername**|`string`|sql driver name<br/>|no|
|**multiwrite**|`boolean`|enables writing to two databases simultaneously<br/>|no|
|**primarydbsource**|`string`|dsn of the primary database<br/>|yes|
|**secondarydbsource**|`string`|dsn of the secondary database if multi-write is enabled<br/>|no|
|**cachettl**|`integer`|cache results for subsequent requests<br/>|no|
|**runmigrations**|`boolean`|run migrations on startup<br/>|no|
|**migrationprovider**|`string`|migration provider to use for running migrations<br/>|no|
|**enablehistory**|`boolean`|enable history data to be logged to the database<br/>|no|
|**maxconnections**|`integer`|maximum number of connections to the database<br/>|no|
|**maxidleconnections**|`integer`|maximum number of idle connections to the database<br/>|no|

**Additional Properties:** not allowed  
<a name="jobqueue"></a>
## jobqueue: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**connectionuri**|`string`|||
|**runmigrations**|`boolean`|||
|[**riverconf**](#jobqueueriverconf)|`object`|||
|[**metrics**](#jobqueuemetrics)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "riverconf": {
        "Logger": {},
        "PeriodicJobs": [
            {}
        ],
        "Queues": {},
        "Test": {},
        "Workers": {}
    },
    "metrics": {}
}
```

<a name="jobqueueriverconf"></a>
### jobqueue\.riverconf: object

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
**Example**

```json
{
    "Logger": {},
    "PeriodicJobs": [
        {}
    ],
    "Queues": {},
    "Test": {},
    "Workers": {}
}
```

<a name="jobqueueriverconfjobinsertmiddleware"></a>
#### jobqueue\.riverconf\.JobInsertMiddleware: array

**Items**

<a name="jobqueueriverconfhooks"></a>
#### jobqueue\.riverconf\.Hooks: array

**Items**

<a name="jobqueueriverconflogger"></a>
#### jobqueue\.riverconf\.Logger: object

**No properties.**

**Additional Properties:** not allowed  
<a name="jobqueueriverconfmiddleware"></a>
#### jobqueue\.riverconf\.Middleware: array

**Items**

<a name="jobqueueriverconfperiodicjobs"></a>
#### jobqueue\.riverconf\.PeriodicJobs: array

**Items**

**Example**

```json
[
    {}
]
```

<a name="jobqueueriverconfqueues"></a>
#### jobqueue\.riverconf\.Queues: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#jobqueueriverconfqueuesadditionalproperties)|`object`|||

<a name="jobqueueriverconfqueuesadditionalproperties"></a>
##### jobqueue\.riverconf\.Queues\.additionalProperties: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**FetchCooldown**|`integer`|||
|**FetchPollInterval**|`integer`|||
|**MaxWorkers**|`integer`|||

**Additional Properties:** not allowed  
<a name="jobqueueriverconftest"></a>
#### jobqueue\.riverconf\.Test: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**DisableUniqueEnforcement**|`boolean`|||
|**Time**||||

**Additional Properties:** not allowed  
<a name="jobqueueriverconfworkers"></a>
#### jobqueue\.riverconf\.Workers: object

**No properties.**

**Additional Properties:** not allowed  
<a name="jobqueueriverconfworkermiddleware"></a>
#### jobqueue\.riverconf\.WorkerMiddleware: array

**Items**

<a name="jobqueuemetrics"></a>
### jobqueue\.metrics: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enablemetrics**|`boolean`|||
|**metricsdurationunit**|`string`|||
|**enablesemanticmetrics**|`boolean`|||

**Additional Properties:** not allowed  
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
|**dialtimeout**|`integer`|||
|**readtimeout**|`integer`|||
|**writetimeout**|`integer`|||
|**maxretries**|`integer`|||
|**minidleconns**|`integer`|||
|**maxidleconns**|`integer`|||
|**maxactiveconns**|`integer`|||

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
**Example**

```json
{
    "stdout": {},
    "otlp": {}
}
```

<a name="tracerstdout"></a>
### tracer\.stdout: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**pretty**|`boolean`|||
|**disabletimestamp**|`boolean`|||

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
|**companyname**|`string`|||
|**companyaddress**|`string`|||
|**corporation**|`string`|||
|**year**|`integer`|||
|**fromemail**|`string`|||
|**supportemail**|`string`|||
|**logourl**|`string`|||
|[**urls**](#emailurls)|`object`|||
|**templatespath**|`string`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "urls": {}
}
```

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
|**verifysubscriber**|`string`|||
|**verifybilling**|`string`|||
|**questionnaire**|`string`|||

**Additional Properties:** not allowed  
<a name="sessions"></a>
## sessions: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**signingkey**|`string`|||
|**encryptionkey**|`string`|||
|**domain**|`string`|||
|**maxage**|`integer`|||
|**secure**|`boolean`|||
|**httponly**|`boolean`|||
|**samesite**|`string`|||

**Additional Properties:** not allowed  
<a name="totp"></a>
## totp: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**codelength**|`integer`|||
|**issuer**|`string`|||
|**redis**|`boolean`|||
|**secret**|`string`|||
|**recoverycodecount**|`integer`|||
|**recoverycodelength**|`integer`|||

**Additional Properties:** not allowed  
<a name="ratelimit"></a>
## ratelimit: object

Config defines the configuration settings for the rate limiter middleware.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**options**](#ratelimitoptions)|`array`|||
|[**headers**](#ratelimitheaders)|`string[]`|||
|**forwardedindexfrombehind**|`integer`|ForwardedIndexFromBehind selects which IP from X-Forwarded-For should be used.<br/>0 means the closest client, 1 the proxy behind it, etc.<br/>||
|**includepath**|`boolean`|IncludePath appends the request path to the limiter key when true.<br/>||
|**includemethod**|`boolean`|IncludeMethod appends the request method to the limiter key when true.<br/>||
|**keyprefix**|`string`|KeyPrefix allows scoping the limiter key space with a static prefix.<br/>||
|**denystatus**|`integer`|DenyStatus overrides the HTTP status code returned when a rate limit is exceeded.<br/>||
|**denymessage**|`string`|DenyMessage customises the error payload when a rate limit is exceeded.<br/>||
|**sendretryafterheader**|`boolean`|SendRetryAfterHeader toggles whether the Retry-After header should be added when available.<br/>||
|**dryrun**|`boolean`|DryRun enables logging rate limit decisions without blocking requests.<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "options": [
        {}
    ]
}
```

<a name="ratelimitoptions"></a>
### ratelimit\.options: array

**Items**

**Example**

```json
[
    {}
]
```

<a name="ratelimitheaders"></a>
### ratelimit\.headers: array

**Items**

**Item Type:** `string`  
<a name="objectstorage"></a>
## objectstorage: object

ProviderConfig contains configuration for object storage providers


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if object storage is enabled<br/>||
|[**keys**](#objectstoragekeys)|`string[]`|||
|**maxsizemb**|`integer`|MaxSizeMB is the maximum file size allowed in MB<br/>||
|**maxmemorymb**|`integer`|MaxMemoryMB is the maximum memory to use for file uploads in MB<br/>||
|**devmode**|`boolean`|DevMode automatically configures a local disk storage provider (and ensures directories exist) and ignores other provider configs<br/>||
|[**providers**](#objectstorageproviders)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "providers": {
        "s3": {
            "credentials": {}
        },
        "r2": {
            "credentials": {}
        },
        "disk": {
            "credentials": {}
        },
        "database": {
            "credentials": {}
        }
    }
}
```

<a name="objectstoragekeys"></a>
### objectstorage\.keys: array

**Items**

**Item Type:** `string`  
<a name="objectstorageproviders"></a>
### objectstorage\.providers: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**s3**](#objectstorageproviderss3)|`object`|ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future<br/>||
|[**r2**](#objectstorageprovidersr2)|`object`|ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future<br/>||
|[**disk**](#objectstorageprovidersdisk)|`object`|ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future<br/>||
|[**database**](#objectstorageprovidersdatabase)|`object`|ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "s3": {
        "credentials": {}
    },
    "r2": {
        "credentials": {}
    },
    "disk": {
        "credentials": {}
    },
    "database": {
        "credentials": {}
    }
}
```

<a name="objectstorageproviderss3"></a>
#### objectstorage\.providers\.s3: object

ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if this provider is enabled<br/>||
|**ensureavailable**|`boolean`|EnsureAvailable enforces provider availability before completing server startup<br/>||
|**region**|`string`|Region for cloud providers<br/>||
|**bucket**|`string`|Bucket name for cloud providers<br/>||
|**endpoint**|`string`|Endpoint for custom endpoints<br/>||
|**proxypresignenabled**|`boolean`|ProxyPresignEnabled toggles proxy-signed download URL generation<br/>||
|**baseurl**|`string`|BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files).<br/>||
|[**credentials**](#objectstorageproviderss3credentials)|`object`|ProviderCredentials contains credentials for a storage provider<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageproviderss3credentials"></a>
##### objectstorage\.providers\.s3\.credentials: object

ProviderCredentials contains credentials for a storage provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|AccessKeyID for cloud providers<br/>||
|**secretaccesskey**|`string`|SecretAccessKey for cloud providers<br/>||
|**projectid**|`string`|ProjectID for GCS<br/>||
|**accountid**|`string`|AccountID for Cloudflare R2<br/>||
|**apitoken**|`string`|APIToken for Cloudflare R2<br/>||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersr2"></a>
#### objectstorage\.providers\.r2: object

ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if this provider is enabled<br/>||
|**ensureavailable**|`boolean`|EnsureAvailable enforces provider availability before completing server startup<br/>||
|**region**|`string`|Region for cloud providers<br/>||
|**bucket**|`string`|Bucket name for cloud providers<br/>||
|**endpoint**|`string`|Endpoint for custom endpoints<br/>||
|**proxypresignenabled**|`boolean`|ProxyPresignEnabled toggles proxy-signed download URL generation<br/>||
|**baseurl**|`string`|BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files).<br/>||
|[**credentials**](#objectstorageprovidersr2credentials)|`object`|ProviderCredentials contains credentials for a storage provider<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersr2credentials"></a>
##### objectstorage\.providers\.r2\.credentials: object

ProviderCredentials contains credentials for a storage provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|AccessKeyID for cloud providers<br/>||
|**secretaccesskey**|`string`|SecretAccessKey for cloud providers<br/>||
|**projectid**|`string`|ProjectID for GCS<br/>||
|**accountid**|`string`|AccountID for Cloudflare R2<br/>||
|**apitoken**|`string`|APIToken for Cloudflare R2<br/>||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersdisk"></a>
#### objectstorage\.providers\.disk: object

ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if this provider is enabled<br/>||
|**ensureavailable**|`boolean`|EnsureAvailable enforces provider availability before completing server startup<br/>||
|**region**|`string`|Region for cloud providers<br/>||
|**bucket**|`string`|Bucket name for cloud providers<br/>||
|**endpoint**|`string`|Endpoint for custom endpoints<br/>||
|**proxypresignenabled**|`boolean`|ProxyPresignEnabled toggles proxy-signed download URL generation<br/>||
|**baseurl**|`string`|BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files).<br/>||
|[**credentials**](#objectstorageprovidersdiskcredentials)|`object`|ProviderCredentials contains credentials for a storage provider<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersdiskcredentials"></a>
##### objectstorage\.providers\.disk\.credentials: object

ProviderCredentials contains credentials for a storage provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|AccessKeyID for cloud providers<br/>||
|**secretaccesskey**|`string`|SecretAccessKey for cloud providers<br/>||
|**projectid**|`string`|ProjectID for GCS<br/>||
|**accountid**|`string`|AccountID for Cloudflare R2<br/>||
|**apitoken**|`string`|APIToken for Cloudflare R2<br/>||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersdatabase"></a>
#### objectstorage\.providers\.database: object

ProviderConfigs contains configuration for all storage providers This is structured to allow easy extension for additional providers in the future


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if this provider is enabled<br/>||
|**ensureavailable**|`boolean`|EnsureAvailable enforces provider availability before completing server startup<br/>||
|**region**|`string`|Region for cloud providers<br/>||
|**bucket**|`string`|Bucket name for cloud providers<br/>||
|**endpoint**|`string`|Endpoint for custom endpoints<br/>||
|**proxypresignenabled**|`boolean`|ProxyPresignEnabled toggles proxy-signed download URL generation<br/>||
|**baseurl**|`string`|BaseURL is the prefix for proxy download URLs (e.g., http://localhost:17608/v1/files).<br/>||
|[**credentials**](#objectstorageprovidersdatabasecredentials)|`object`|ProviderCredentials contains credentials for a storage provider<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersdatabasecredentials"></a>
##### objectstorage\.providers\.database\.credentials: object

ProviderCredentials contains credentials for a storage provider


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|AccessKeyID for cloud providers<br/>||
|**secretaccesskey**|`string`|SecretAccessKey for cloud providers<br/>||
|**projectid**|`string`|ProjectID for GCS<br/>||
|**accountid**|`string`|AccountID for Cloudflare R2<br/>||
|**apitoken**|`string`|APIToken for Cloudflare R2<br/>||

**Additional Properties:** not allowed  
<a name="subscription"></a>
## subscription: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled determines if the entitlements service is enabled<br/>||
|**privatestripekey**|`string`|PrivateStripeKey is the key for the stripe service<br/>||
|**stripewebhooksecret**|`string`|StripeWebhookSecret is the secret for the stripe service (legacy, use StripeWebhookSecrets for version-specific secrets)<br/>||
|[**stripewebhooksecrets**](#subscriptionstripewebhooksecrets)|`object`|||
|**stripewebhookurl**|`string`|StripeWebhookURL is the URL for the stripe webhook<br/>||
|**stripebillingportalsuccessurl**|`string`|StripeBillingPortalSuccessURL<br/>||
|**stripecancellationreturnurl**|`string`|StripeCancellationReturnURL is the URL for the stripe cancellation return<br/>||
|[**stripewebhookevents**](#subscriptionstripewebhookevents)|`string[]`|||
|**stripewebhookapiversion**|`string`|StripeWebhookAPIVersion is the Stripe API version currently accepted by the webhook handler<br/>||
|**stripewebhookdiscardapiversion**|`string`|StripeWebhookDiscardAPIVersion is the Stripe API version to discard during migration<br/>||

**Additional Properties:** not allowed  
**Example**

```json
{
    "stripewebhooksecrets": {}
}
```

<a name="subscriptionstripewebhooksecrets"></a>
### subscription\.stripewebhooksecrets: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

<a name="subscriptionstripewebhookevents"></a>
### subscription\.stripewebhookevents: array

**Items**

**Item Type:** `string`  
<a name="keywatcher"></a>
## keywatcher: object

KeyWatcher contains settings for the key watcher that manages JWT signing keys


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether the key watcher is enabled<br/>||
|**keydir**|`string`|KeyDir is the path to the directory containing PEM keys for JWT signing<br/>||
|**externalsecretsintegration**|`boolean`|ExternalSecretsIntegration enables integration with external secret management systems (specifically GCP secret manager today)<br/>||
|**secretmanager**|`string`|SecretManagerSecret is the name of the GCP Secret Manager secret containing the JWT signing key<br/>||

**Additional Properties:** not allowed  
<a name="slack"></a>
## slack: object

Slack contains settings for Slack notifications


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**webhookurl**|`string`|WebhookURL is the Slack webhook to post messages to<br/>||
|**newsubscribermessagefile**|`string`|NewSubscriberMessageFile is the path to the template used for new subscriber notifications<br/>||
|**newusermessagefile**|`string`|NewUserMessageFile is the path to the template used for new user notifications<br/>||

**Additional Properties:** not allowed  
<a name="integrationoauthprovider"></a>
## integrationoauthprovider: object

IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled toggles initialization of the integration provider registry.<br/>||
|**successredirecturl**|`string`|SuccessRedirectURL is the URL to redirect to after successful OAuth integration.<br/>||

**Additional Properties:** not allowed  

