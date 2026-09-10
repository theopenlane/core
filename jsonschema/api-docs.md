# object

Config contains the configuration for the core server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**domain**|`string`|Domain provides a global domain value for other modules to inherit<br/>||
|**refreshinterval**|`integer`|RefreshInterval determines how often to reload the config<br/>||
|[**server**](#defsconfigserver)|`object`|Server settings for the echo server<br/>|yes|
|[**entconfig**](#defsentconfigconfig)|`object`|Config holds the configuration for the ent server<br/>||
|[**auth**](#defsconfigauth)|`object`|Auth settings including oauth2 providers and token configuration<br/>|yes|
|[**authz**](#defsfgaxconfig)|`object`||yes|
|[**db**](#defsentxconfig)|`object`||yes|
|[**jobqueue**](#defsriverqueueconfig)|`object`|||
|[**redis**](#defscacheconfig)|`object`|||
|[**sessions**](#defssessionsconfig)|`object`|||
|[**totp**](#defstotpconfig)|`object`|||
|[**ratelimit**](#defsratelimitconfig)|`object`|Config defines the configuration settings for the rate limiter middleware.<br/>||
|[**ratelimitunmatched**](#defsratelimitconfig)|`object`|Config defines the configuration settings for the rate limiter middleware.<br/>||
|[**objectstorage**](#defsstorageproviderconfig)|`object`|ProviderConfig contains configuration for object storage providers<br/>||
|[**subscription**](#defsentitlementsconfig)|`object`|||
|[**keywatcher**](#defsconfigkeywatcher)|`object`|KeyWatcher contains settings for the key watcher that manages JWT signing keys<br/>||
|[**integrations**](#defscatalogconfig)|`object`|||
|[**workflows**](#defsworkflowsconfig)|`object`|||
|[**cloudflare**](#defshandlerscloudflareconfig)|`object`|CloudflareConfig contains configuration for Cloudflare integration.<br/>||
|[**shortlinks**](#defsshortlinksconfig)|`object`|||
|[**backfill**](#defsconfigbackfill)|`object`|Backfill configures one-time startup data backfill routines that populate fields introduced by recent<br/>||

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
        "cachecontrol": {
            "nocacheheaders": {}
        },
        "mime": {},
        "graphpool": {},
        "csrfprotection": {}
    },
    "entconfig": {
        "summarizer": {
            "llm": {
                "anthropic": {},
                "cloudflare": {},
                "openai": {}
            }
        },
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
        },
        "supportaccess": {}
    },
    "authz": {
        "credentials": {},
        "parentcontextconditions": [
            {
                "context": {}
            }
        ]
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
    "sessions": {},
    "totp": {},
    "ratelimit": {
        "options": [
            {}
        ]
    },
    "ratelimitunmatched": {
        "options": [
            {}
        ]
    },
    "objectstorage": {
        "providers": {
            "s3": {
                "credentials": {},
                "backup": {}
            },
            "r2": {
                "credentials": {},
                "backup": {}
            },
            "disk": {
                "credentials": {},
                "backup": {}
            },
            "database": {
                "credentials": {},
                "backup": {}
            }
        }
    },
    "subscription": {
        "stripewebhooksecrets": {}
    },
    "keywatcher": {},
    "integrations": {
        "awssecurityhub": {},
        "cloudflareruntime": {
            "domainscan": {}
        },
        "githubapp": {},
        "slack": {},
        "slackruntime": {},
        "googledrive": {},
        "googleworkspace": {},
        "azureentraid": {},
        "microsoftteams": {},
        "onedrive": {},
        "oidclocal": {},
        "email": {},
        "paymentreminder": {
            "paymentmethodinterval": 30,
            "deletiondays": 7,
            "enabled": false,
            "dryrun": true
        },
        "organizationdelete": {
            "maxdeletesperrun": 25
        },
        "integrationlifecycle": {
            "enabled": true,
            "dryrun": true,
            "maxperrun": 100
        }
    },
    "workflows": {
        "cel": {},
        "gala": {}
    },
    "cloudflare": {},
    "shortlinks": {},
    "backfill": {}
}
```

   
<a name="defsconfigserver"></a>
## $defs/config\.Server: object

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
|[**tls**](#defsconfigtls)|`object`|TLS settings for the server for secure connections<br/>||
|[**cors**](#defscorsconfig)|`object`|Config holds the cors configuration settings<br/>||
|[**secure**](#defssecureconfig)|`object`|Config contains the types used in the mw middleware<br/>||
|[**cachecontrol**](#defscachecontrolconfig)|`object`|Config is the config values for the cache-control middleware<br/>||
|[**mime**](#defsmimeconfig)|`object`|Config defines the config for Mime middleware<br/>||
|[**graphpool**](#defsconfigpoolconfig)|`object`|PoolConfig contains the settings for the goroutine pool<br/>||
|**enablegraphextensions**|`boolean`|EnableGraphExtensions enables the graph extensions for the graph resolvers<br/>|no|
|**enablegraphsubscriptions**|`boolean`|EnableGraphSubscriptions enables graphql subscriptions to the server using websockets or sse<br/>|no|
|**complexitylimit**|`integer`|ComplexityLimit sets the maximum complexity allowed for a query<br/>|no|
|**maxresultlimit**|`integer`|MaxResultLimit sets the maximum number of results allowed for a query<br/>|no|
|[**csrfprotection**](#defscsrfconfig)|`object`|Config defines configuration for the CSRF middleware wrapper.<br/>||
|**secretmanager**|`string`|SecretManagerSecret is the name of the GCP Secret Manager secret containing the JWT signing key<br/>|no|
|**defaulttrustcenterdomain**|`string`|DefaultTrustCenterDomain is the default domain to use for the trust center if no custom domain is set<br/>|no|
|**trustcentercnametarget**|`string`|TrustCenterCnameTarget is the cname target for the trust center<br/>Used for mapping the vanity domains to the trust centers<br/>|no|
|**trustcenterpreviewcnametarget**|`string`|TrustCenterPreviewCnameTarget is the cname target for trust center preview domains<br/>|no|
|**notificationlookbackdays**|`integer`|NotificationLookbackDays is the number of days of read notifications to pull when starting a notification subscription<br/>Unread notifications are always pulled regardless of this setting<br/>|no|

**Additional Properties:** not allowed   
**Example**

```json
{
    "tls": {},
    "cors": {
        "prefixes": {}
    },
    "secure": {},
    "cachecontrol": {
        "nocacheheaders": {}
    },
    "mime": {},
    "graphpool": {},
    "csrfprotection": {}
}
```

   
<a name="defsconfigtls"></a>
### $defs/config\.TLS: object

TLS settings for the server for secure connections


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled turns on TLS settings for the server<br/>||
|**certfile**|`string`|CertFile location for the TLS server<br/>||
|**certkey**|`string`|CertKey file location for the TLS server<br/>||
|**autocert**|`boolean`|AutoCert generates the cert with letsencrypt, this does not work on localhost<br/>||

**Additional Properties:** not allowed   
   
<a name="defscorsconfig"></a>
### $defs/cors\.Config: object

Config holds the cors configuration settings


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enable or disable the CORS middleware<br/>||
|[**prefixes**](#defsmapstringstring)|`object`|||
|[**alloworigins**](#defsstring)|`string[]`|||
|**cookieinsecure**|`boolean`|CookieInsecure sets the cookie to be insecure<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "prefixes": {}
}
```

   
<a name="defsmapstringstring"></a>
#### $defs/map\[string\]\[\]string: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#defsstring)|`string[]`|||

   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defssecureconfig"></a>
### $defs/secure\.Config: object

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
   
<a name="defscachecontrolconfig"></a>
### $defs/cachecontrol\.Config: object

Config is the config values for the cache-control middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**nocacheheaders**](#defsmapstringstring)|`object`|||
|[**etagheaders**](#defsstring)|`string[]`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "nocacheheaders": {}
}
```

   
<a name="defsmapstringstring"></a>
#### $defs/map\[string\]string: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

   
<a name="defsstring"></a>
#### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsmimeconfig"></a>
### $defs/mime\.Config: object

Config defines the config for Mime middleware


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if the mime middleware should be enabled<br/>||
|**mimetypesfile**|`string`|MimeTypesFile is the file to load mime types from<br/>||
|**defaultcontenttype**|`string`|DefaultContentType is the default content type to set if no mime type is found<br/>||

**Additional Properties:** not allowed   
   
<a name="defsconfigpoolconfig"></a>
### $defs/config\.PoolConfig: object

PoolConfig contains the settings for the goroutine pool


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**maxworkers**|`integer`|MaxWorkers is the maximum number of workers in the pool<br/>||

**Additional Properties:** not allowed   
   
<a name="defscsrfconfig"></a>
### $defs/csrf\.Config: object

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
   
<a name="defsentconfigconfig"></a>
## $defs/entconfig\.Config: object

Config holds the configuration for the ent server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**entitytypes**](#defsstring)|`string[]`|||
|[**summarizer**](#defssummarizerconfig)|`object`|Config holds configuration for the text summarization functionality<br/>||
|**maxpoolsize**|`integer`|MaxPoolSize is the max worker pool size that can be used by the ent client<br/>||
|[**modules**](#defsentconfigmodules)|`object`|Modules settings for features access<br/>||
|**maxschemaimportsize**|`integer`|MaxSchemaImportSize is the maximum size allowed for schema imports in bytes<br/>||
|[**emailvalidation**](#defsvalidatoremailverificationconfig)|`object`|EmailVerificationConfig is the configuration for email verification<br/>||
|[**billing**](#defsentconfigbilling)|`object`|Billing settings for feature access<br/>||
|[**notifications**](#defsentconfignotifications)|`object`|Notifications settings for notifications sent to users based on events<br/>||
|**questionnaireproducturl**|`string`|QuestionnaireProductURL is the product URL used to build questionnaire access links<br/>||

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
    "modules": {},
    "emailvalidation": {
        "allowedemailtypes": {}
    },
    "billing": {},
    "notifications": {}
}
```

   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defssummarizerconfig"></a>
### $defs/summarizer\.Config: object

Config holds configuration for the text summarization functionality


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**type**|`string`|Type specifies the summarization algorithm to use<br/>||
|[**llm**](#defssummarizerllm)|`object`|LLM contains configuration for multiple LLM providers<br/>||
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

   
<a name="defssummarizerllm"></a>
#### $defs/summarizer\.LLM: object

LLM contains configuration for multiple LLM providers


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**provider**|`string`|Provider specifies which LLM service to use<br/>||
|[**anthropic**](#defssummarizeranthropicconfig)|`object`|AnthropicConfig contains Anthropic specific configuration<br/>||
|[**cloudflare**](#defssummarizercloudflareconfig)|`object`|CloudflareConfig contains Cloudflare specific configuration<br/>||
|[**openai**](#defssummarizeropenaiconfig)|`object`|OpenAIConfig contains OpenAI specific configuration<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "anthropic": {},
    "cloudflare": {},
    "openai": {}
}
```

   
<a name="defssummarizeranthropicconfig"></a>
##### $defs/summarizer\.AnthropicConfig: object

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
   
<a name="defssummarizercloudflareconfig"></a>
##### $defs/summarizer\.CloudflareConfig: object

CloudflareConfig contains Cloudflare specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apikey**|`string`|APIKey contains the authentication key for the service<br/>||
|**accountid**|`string`|AccountID specifies the Cloudflare account ID<br/>||
|**serverurl**|`string`|ServerURL specifies the API endpoint<br/>||

**Additional Properties:** not allowed   
   
<a name="defssummarizeropenaiconfig"></a>
##### $defs/summarizer\.OpenAIConfig: object

OpenAIConfig contains OpenAI specific configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|Model specifies the model name to use<br/>||
|**apikey**|`string`|APIKey contains the authentication key for the service<br/>||
|**url**|`string`|URL specifies the API endpoint<br/>||
|**organizationid**|`string`|OrganizationID specifies the OpenAI organization ID<br/>||

**Additional Properties:** not allowed   
   
<a name="defsentconfigmodules"></a>
### $defs/entconfig\.Modules: object

Modules settings for features access


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether to check and verify module access<br/>||
|**usesandbox**|`boolean`|UseSandbox indicates whether to use the sandbox catalog for module access checks<br/>||
|**devmode**|`boolean`|DevMode enables all modules for local development regardless of trial status<br/>||

**Additional Properties:** not allowed   
   
<a name="defsvalidatoremailverificationconfig"></a>
### $defs/validator\.EmailVerificationConfig: object

EmailVerificationConfig is the configuration for email verification


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether email verification is enabled<br/>||
|**enableautoupdatedisposable**|`boolean`|EnableAutoUpdateDisposable indicates whether to automatically update disposable email addresses<br/>||
|**enablegravatarcheck**|`boolean`|EnableGravatarCheck indicates whether to check for Gravatar existence<br/>||
|**enablesmtpcheck**|`boolean`|EnableSMTPCheck indicates whether to check email by smtp<br/>||
|[**allowedemailtypes**](#defsvalidatorallowedemailtypes)|`object`|AllowedEmailTypes defines the allowed email types for verification<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "allowedemailtypes": {}
}
```

   
<a name="defsvalidatorallowedemailtypes"></a>
#### $defs/validator\.AllowedEmailTypes: object

AllowedEmailTypes defines the allowed email types for verification


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**disposable**|`boolean`|Disposable indicates whether disposable email addresses are allowed<br/>||
|**free**|`boolean`|Free indicates whether free email addresses are allowed<br/>||
|**role**|`boolean`|Role indicates whether role-based email addresses are allowed<br/>||

**Additional Properties:** not allowed   
   
<a name="defsentconfigbilling"></a>
### $defs/entconfig\.Billing: object

Billing settings for feature access


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**requirepaymentmethod**|`boolean`|RequirePaymentMethod indicates whether to check if a payment method<br/>exists for orgs before they can access some resource<br/>||
|[**bypassemaildomains**](#defsstring)|`string[]`|||

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
#### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsentconfignotifications"></a>
### $defs/entconfig\.Notifications: object

Notifications settings for notifications sent to users based on events


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleurl**|`string`|ConsoleURL for ui links used in notifications<br/>||

**Additional Properties:** not allowed   
   
<a name="defsconfigauth"></a>
## $defs/config\.Auth: object

Auth settings including oauth2 providers and token configuration


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled authentication on the server, not recommended to disable<br/>|no|
|[**token**](#defstokensconfig)|`object`||yes|
|[**supportedproviders**](#defsstring)|`string[]`|||
|[**providers**](#defshandlersoauthproviderconfig)|`object`|OauthProviderConfig represents the configuration for OAuth providers such as Github and Google<br/>||
|[**supportaccess**](#defshandlerssupportaccessconfig)|`object`|SupportAccessConfig contains configuration for the Openlane support access flow. The support<br/>||

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
    },
    "supportaccess": {}
}
```

   
<a name="defstokensconfig"></a>
### $defs/tokens\.Config: object

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
|[**keys**](#defsmapstringstring)|`object`|||
|**generatekeys**|`boolean`||no|
|**jwkscachettl**|`integer`||no|
|[**redis**](#defstokensredisconfig)|`object`|||
|[**apitokens**](#defstokensapitokenconfig)|`object`|||
|**assessmentaccessduration**|`integer`||no|
|**trustcenterndarequestaccessduration**|`integer`||no|

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

   
<a name="defsmapstringstring"></a>
#### $defs/map\[string\]string: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

   
<a name="defstokensredisconfig"></a>
#### $defs/tokens\.RedisConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**config**](#defscacheconfig)|`object`|||
|**blacklistprefix**|`string`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "config": {}
}
```

   
<a name="defscacheconfig"></a>
##### $defs/cache\.Config: object

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
   
<a name="defstokensapitokenconfig"></a>
#### $defs/tokens\.APITokenConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**envprefix**|`string`|||
|[**keys**](#defsmapstringtokensapitokenkeyconfig)|`object`|||
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

   
<a name="defsmapstringtokensapitokenkeyconfig"></a>
##### $defs/map\[string\]tokens\.APITokenKeyConfig: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#defstokensapitokenkeyconfig)|`object`|||

   
<a name="defstokensapitokenkeyconfig"></a>
###### $defs/tokens\.APITokenKeyConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**secret**|`string`|||
|**status**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defshandlersoauthproviderconfig"></a>
### $defs/handlers\.OauthProviderConfig: object

OauthProviderConfig represents the configuration for OAuth providers such as Github and Google


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**redirecturl**|`string`|RedirectURL is the URL that the OAuth2 client will redirect to after authentication is complete<br/>||
|[**github**](#defsgithubproviderconfig)|`object`||yes|
|[**google**](#defsgoogleproviderconfig)|`object`||yes|
|[**webauthn**](#defswebauthnproviderconfig)|`object`||yes|

**Additional Properties:** not allowed   
**Example**

```json
{
    "github": {},
    "google": {},
    "webauthn": {}
}
```

   
<a name="defsgithubproviderconfig"></a>
#### $defs/github\.ProviderConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`||yes|
|**clientsecret**|`string`||yes|
|**clientendpoint**|`string`||no|
|[**scopes**](#defsstring)|`string[]`|||
|**redirecturl**|`string`||yes|

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsgoogleproviderconfig"></a>
#### $defs/google\.ProviderConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`||yes|
|**clientsecret**|`string`||yes|
|**clientendpoint**|`string`||no|
|[**scopes**](#defsstring)|`string[]`|||
|**redirecturl**|`string`||yes|

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defswebauthnproviderconfig"></a>
#### $defs/webauthn\.ProviderConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`||no|
|**displayname**|`string`||yes|
|**relyingpartyid**|`string`||yes|
|[**requestorigins**](#defsstring)|`string[]`|||
|**maxdevices**|`integer`||no|
|**enforcetimeout**|`boolean`||no|
|**timeout**|`integer`||no|
|**debug**|`boolean`||no|

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defshandlerssupportaccessconfig"></a>
### $defs/handlers\.SupportAccessConfig: object

SupportAccessConfig contains configuration for the Openlane support access flow. The support
identity is virtual and authenticated entirely from these values, never from the database. This is
the single place that holds the support identity, its shared password, and the second factor
identity provider configuration, since both authentications must occur together


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled toggles the support access endpoints<br/>||
|**email**|`string`|Email is the email of the virtual support identity, used as the first factor username<br/>||
|**displayname**|`string`|DisplayName is the display name of the virtual support identity, used for record attribution<br/>||
|**subjectid**|`string`|SubjectID is the stable subject id of the virtual support identity used for created_by/updated_by<br/>attribution. It must be a valid ULID and is consistent without a backing user row. Default should match auth.SupportSubjectID<br/>||
|**password**|`string`|Password is the shared password for the virtual support identity, validated against this value<br/>||
|**clientid**|`string`|ClientID is the client ID for the second factor identity provider<br/>||
|**clientsecret**|`string`|ClientSecret is the client secret for the second factor identity provider<br/>||
|**issuerurl**|`string`|IssuerURL is the issuer URL of the second factor identity provider<br/>||
|**discoveryendpoint**|`string`|DiscoveryEndpoint is the optional OIDC discovery endpoint of the second factor identity provider<br/>||
|**redirecturl**|`string`|RedirectURL is the callback URL registered with the second factor identity provider<br/>||
|**alloweddomain**|`string`|AllowedDomain restricts which email domain may complete the second factor (e.g. theopenlane.io)<br/>||

**Additional Properties:** not allowed   
   
<a name="defsfgaxconfig"></a>
## $defs/fgax\.Config: object

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
|**modulefile**|`string`|path to the fga module file<br/>|no|
|[**credentials**](#defsfgaxcredentials)|`object`|||
|**maxbatchwritesize**|`integer`|maximum number of writes per batch in a transaction<br/>|no|
|**enableparentcontext**|`boolean`|disables the automatic addition of parent context tuples<br/>|no|
|[**parentcontextskipkinds**](#defsstring)|`string[]`|||
|[**parentcontextconditions**](#defsfgaxparentcontextconditionconfig)|`array`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "credentials": {},
    "parentcontextconditions": [
        {
            "context": {}
        }
    ]
}
```

   
<a name="defsfgaxcredentials"></a>
### $defs/fgax\.Credentials: object

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
   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsfgaxparentcontextconditionconfig"></a>
### $defs/\[\]fgax\.ParentContextConditionConfig: array

**Items**

**Example**

```json
[
    {
        "context": {}
    }
]
```

   
<a name="defsentxconfig"></a>
## $defs/entx\.Config: object

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
   
<a name="defsriverqueueconfig"></a>
## $defs/riverqueue\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**connectionuri**|`string`|||
|**runmigrations**|`boolean`|||
|[**riverconf**](#defsriverconfig)|`object`|||
|[**metrics**](#defsriverqueuemetricsconfig)|`object`|||

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

   
<a name="defsriverconfig"></a>
### $defs/river\.Config: object

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
|[**JobInsertMiddleware**](#defsrivertypejobinsertmiddleware)|`array`|||
|**JobStuckHandler**||||
|**JobStuckThreshold**|`integer`|||
|**JobTimeout**|`integer`|||
|[**Hooks**](#defsrivertypehook)|`array`|||
|[**Logger**](#defssloglogger)|`object`|||
|**MaxAttempts**|`integer`|||
|[**Middleware**](#defsrivertypemiddleware)|`array`|||
|[**Plugins**](#defsrivertypeplugin)|`array`|||
|[**PeriodicJobs**](#defsriverperiodicjob)|`array`|||
|**PollOnly**|`boolean`|||
|[**Queues**](#defsmapstringriverqueueconfig)|`object`|||
|**ReindexerSchedule**||||
|[**ReindexerIndexNames**](#defsstring)|`string[]`|||
|**ReindexerTimeout**|`integer`|||
|**RescueStuckJobsAfter**|`integer`|||
|**RetryPolicy**||||
|**Schema**|`string`|||
|**SoftStopTimeout**|`integer`|||
|**SkipJobKindValidation**|`boolean`|||
|**SkipUnknownJobCheck**|`boolean`|||
|[**Test**](#defsrivertestconfig)|`object`|||
|**TestOnly**|`boolean`|||
|[**Workers**](#defsriverworkers)|`object`|||
|[**WorkerMiddleware**](#defsrivertypeworkermiddleware)|`array`|||

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

   
<a name="defsrivertypejobinsertmiddleware"></a>
#### $defs/\[\]rivertype\.JobInsertMiddleware: array

**Items**

   
<a name="defsrivertypehook"></a>
#### $defs/\[\]rivertype\.Hook: array

**Items**

   
<a name="defssloglogger"></a>
#### $defs/slog\.Logger: object

**No properties.**

**Additional Properties:** not allowed   
   
<a name="defsrivertypemiddleware"></a>
#### $defs/\[\]rivertype\.Middleware: array

**Items**

   
<a name="defsrivertypeplugin"></a>
#### $defs/\[\]rivertype\.Plugin: array

**Items**

   
<a name="defsriverperiodicjob"></a>
#### $defs/\[\]\*river\.PeriodicJob: array

**Items**

**Example**

```json
[
    {}
]
```

   
<a name="defsmapstringriverqueueconfig"></a>
#### $defs/map\[string\]river\.QueueConfig: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**Additional Properties**](#defsriverqueueconfig)|`object`|||

   
<a name="defsriverqueueconfig"></a>
##### $defs/river\.QueueConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**FetchCooldown**|`integer`|||
|**FetchPollInterval**|`integer`|||
|**MaxWorkers**|`integer`|||

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
#### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsrivertestconfig"></a>
#### $defs/river\.TestConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**DisableUniqueEnforcement**|`boolean`|||
|**Time**||||

**Additional Properties:** not allowed   
   
<a name="defsriverworkers"></a>
#### $defs/river\.Workers: object

**No properties.**

**Additional Properties:** not allowed   
   
<a name="defsrivertypeworkermiddleware"></a>
#### $defs/\[\]rivertype\.WorkerMiddleware: array

**Items**

   
<a name="defsriverqueuemetricsconfig"></a>
### $defs/riverqueue\.MetricsConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enablemetrics**|`boolean`|||
|**metricsdurationunit**|`string`|||
|**enablesemanticmetrics**|`boolean`|||

**Additional Properties:** not allowed   
   
<a name="defscacheconfig"></a>
##### $defs/cache\.Config: object

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
   
<a name="defssessionsconfig"></a>
## $defs/sessions\.Config: object

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
   
<a name="defstotpconfig"></a>
## $defs/totp\.Config: object

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
   
<a name="defsratelimitconfig"></a>
## $defs/ratelimit\.Config: object

Config defines the configuration settings for the rate limiter middleware.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**options**](#defsratelimitrateoption)|`array`|||
|[**headers**](#defsstring)|`string[]`|||
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

   
<a name="defsratelimitrateoption"></a>
### $defs/\[\]ratelimit\.RateOption: array

**Items**

**Example**

```json
[
    {}
]
```

   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsstorageproviderconfig"></a>
## $defs/storage\.ProviderConfig: object

ProviderConfig contains configuration for object storage providers


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if object storage is enabled<br/>||
|[**keys**](#defsstring)|`string[]`|||
|**maxsizemb**|`integer`|MaxSizeMB is the maximum file size allowed in MB<br/>||
|**maxmemorymb**|`integer`|MaxMemoryMB is the maximum memory to use for file uploads in MB<br/>||
|**devmode**|`boolean`|DevMode automatically configures a local disk storage provider (and ensures directories exist) and ignores other provider configs<br/>||
|[**providers**](#defsstorageproviders)|`object`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "providers": {
        "s3": {
            "credentials": {},
            "backup": {}
        },
        "r2": {
            "credentials": {},
            "backup": {}
        },
        "disk": {
            "credentials": {},
            "backup": {}
        },
        "database": {
            "credentials": {},
            "backup": {}
        }
    }
}
```

   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsstorageproviders"></a>
### $defs/storage\.Providers: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**s3**](#defsstorageproviderconfigs)|`object`|ProviderConfigs contains configuration for all storage providers<br/>||
|[**r2**](#defsstorageproviderconfigs)|`object`|ProviderConfigs contains configuration for all storage providers<br/>||
|[**disk**](#defsstorageproviderconfigs)|`object`|ProviderConfigs contains configuration for all storage providers<br/>||
|[**database**](#defsstorageproviderconfigs)|`object`|ProviderConfigs contains configuration for all storage providers<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "s3": {
        "credentials": {},
        "backup": {}
    },
    "r2": {
        "credentials": {},
        "backup": {}
    },
    "disk": {
        "credentials": {},
        "backup": {}
    },
    "database": {
        "credentials": {},
        "backup": {}
    }
}
```

   
<a name="defsstorageproviderconfigs"></a>
#### $defs/storage\.ProviderConfigs: object

ProviderConfigs contains configuration for all storage providers
This is structured to allow easy extension for additional providers in the future


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
|[**credentials**](#defsstorageprovidercredentials)|`object`|ProviderCredentials contains credentials for a storage provider<br/>||
|[**backup**](#defsstoragebackupconfig)|`object`|BackupConfig defines an asynchronous replication target for a provider's objects. It only<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "credentials": {},
    "backup": {}
}
```

   
<a name="defsstorageprovidercredentials"></a>
##### $defs/storage\.ProviderCredentials: object

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
   
<a name="defsstoragebackupconfig"></a>
##### $defs/storage\.BackupConfig: object

BackupConfig defines an asynchronous replication target for a provider's objects. It only
names whether backups run and where they go; the destination provider's own configuration
supplies the region, endpoint, and credentials


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates if this backup target is enabled<br/>||
|**provider**|`string`|Provider names the destination backend type, e.g. s3; empty replicates to the source<br/>provider itself, which writes the backup to the suffixed bucket alongside the live one<br/>||
|**readfrombackup**|`boolean`|ReadFromBackup serves reads from this backup target instead of the source provider, intended<br/>to be enabled during a disaster recovery event when the source provider storage is lost<br/>||
|**region**|`string`|Region optionally overrides the destination provider's region, so a backup can replicate<br/>into a region other than the one holding the live objects<br/>||

**Additional Properties:** not allowed   
   
<a name="defsentitlementsconfig"></a>
## $defs/entitlements\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled determines if the entitlements service is enabled<br/>||
|**privatestripekey**|`string`|PrivateStripeKey is the key for the stripe service<br/>||
|**stripewebhooksecret**|`string`|StripeWebhookSecret is the secret for the stripe service (legacy, use StripeWebhookSecrets for version-specific secrets)<br/>||
|[**stripewebhooksecrets**](#defsmapstringstring)|`object`|||
|**stripewebhookurl**|`string`|StripeWebhookURL is the URL for the stripe webhook<br/>||
|**stripebillingportalsuccessurl**|`string`|StripeBillingPortalSuccessURL<br/>||
|**stripecancellationreturnurl**|`string`|StripeCancellationReturnURL is the URL for the stripe cancellation return<br/>||
|[**stripewebhookevents**](#defsstring)|`string[]`|||
|**stripewebhookapiversion**|`string`|StripeWebhookAPIVersion is the Stripe API version currently accepted by the webhook handler<br/>||
|**stripewebhookdiscardapiversion**|`string`|StripeWebhookDiscardAPIVersion is the Stripe API version to discard during migration<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "stripewebhooksecrets": {}
}
```

   
<a name="defsmapstringstring"></a>
### $defs/map\[string\]string: object

**Additional Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**Additional Properties**|`string`|||

   
<a name="defsstring"></a>
### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsconfigkeywatcher"></a>
## $defs/config\.KeyWatcher: object

KeyWatcher contains settings for the key watcher that manages JWT signing keys


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled indicates whether the key watcher is enabled<br/>||
|**keydir**|`string`|KeyDir is the path to the directory containing PEM keys for JWT signing<br/>||

**Additional Properties:** not allowed   
   
<a name="defscatalogconfig"></a>
## $defs/catalog\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleintegrationpath**|`string`|||
|[**awssecurityhub**](#defsawssecurityhubconfig)|`object`|||
|[**cloudflareruntime**](#defscloudflareruntimeconfig)|`object`|||
|[**githubapp**](#defsgithubappconfig)|`object`|||
|[**slack**](#defsslackconfig)|`object`|||
|[**slackruntime**](#defsslackruntimeslackconfig)|`object`|||
|[**googledrive**](#defsgoogledriveconfig)|`object`|||
|[**googleworkspace**](#defsgoogleworkspaceconfig)|`object`|||
|[**azureentraid**](#defsazureentraidconfig)|`object`|||
|[**microsoftteams**](#defsmicrosoftteamsconfig)|`object`|||
|[**onedrive**](#defsonedriveconfig)|`object`|||
|[**oidclocal**](#defsoidclocalconfig)|`object`|||
|[**email**](#defsemailruntimeemailconfig)|`object`||yes|
|[**paymentreminder**](#defssystempaymentreminderconfig)|`object`|||
|[**organizationdelete**](#defssystemorganizationdeleteconfig)|`object`|||
|[**integrationlifecycle**](#defssystemintegrationlifecycleconfig)|`object`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "awssecurityhub": {},
    "cloudflareruntime": {
        "domainscan": {}
    },
    "githubapp": {},
    "slack": {},
    "slackruntime": {},
    "googledrive": {},
    "googleworkspace": {},
    "azureentraid": {},
    "microsoftteams": {},
    "onedrive": {},
    "oidclocal": {},
    "email": {},
    "paymentreminder": {
        "paymentmethodinterval": 30,
        "deletiondays": 7,
        "enabled": false,
        "dryrun": true
    },
    "organizationdelete": {
        "maxdeletesperrun": 25
    },
    "integrationlifecycle": {
        "enabled": true,
        "dryrun": true,
        "maxperrun": 100
    }
}
```

   
<a name="defsawssecurityhubconfig"></a>
### $defs/awssecurityhub\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**arn**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defscloudflareruntimeconfig"></a>
### $defs/cloudflare\.RuntimeConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**apitoken**|`string`|Cloudflare API token for the operator-owned account<br/>||
|**accountid**|`string`|Cloudflare account ID for the operator-owned account<br/>||
|[**domainscan**](#defsdomainscanreportconfig)|`object`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "domainscan": {}
}
```

   
<a name="defsdomainscanreportconfig"></a>
#### $defs/domainscan\.ReportConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**nonvendorcategories**](#defsstring)|`string[]`|||
|[**deniedvendornames**](#defsstring)|`string[]`|||
|**scanttl**|`integer`|||

**Additional Properties:** not allowed   
   
<a name="defsstring"></a>
##### $defs/\[\]string: array

**Items**

**Item Type:** `string`   
   
<a name="defsgithubappconfig"></a>
### $defs/githubapp\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**appid**|`string`|||
|**privatekey**|`string`|||
|**webhooksecret**|`string`|||
|**appslug**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsslackconfig"></a>
### $defs/slack\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**appid**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsslackruntimeslackconfig"></a>
### $defs/slack\.RuntimeSlackConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**webhookURL**|`string`|Slack incoming webhook URL for fire-and-forget system notifications<br/>||
|**botToken**|`string`|Bot User OAuth Token for full Web API access to the platform workspace<br/>||
|**defaultChannel**|`string`|Default channel id for system messages when no explicit channel is provided<br/>||

**Additional Properties:** not allowed   
   
<a name="defsgoogledriveconfig"></a>
### $defs/googledrive\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsgoogleworkspaceconfig"></a>
### $defs/googleworkspace\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsazureentraidconfig"></a>
### $defs/azureentraid\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**defaulttenant**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsmicrosoftteamsconfig"></a>
### $defs/microsoftteams\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsonedriveconfig"></a>
### $defs/onedrive\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**contentmode**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsoidclocalconfig"></a>
### $defs/oidclocal\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**discoveryurl**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsemailruntimeemailconfig"></a>
### $defs/email\.RuntimeEmailConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**testdir**|`string`|Directory for dev-mode email output<br/>|no|
|**resendsecret**|`string`|Resend webhook signing secret<br/>|no|
|**apikey**|`string`|Email provider API key<br/>|yes|
|**provider**|`string`|Email service provider<br/>Enum: `"resend"`<br/>|yes|
|**fromemail**|`string`|Sender email address<br/>|yes|
|**supportemail**|`string`|Support contact email address<br/>|no|
|**questionnaireemail**|`string`|Sender override for questionnaire auth emails<br/>|no|
|**rooturl**|`string`|Root application URL used to construct email action links<br/>|no|
|**producturl**|`string`|Product home URL<br/>|no|
|**docsurl**|`string`|Documentation URL<br/>|no|
|**apiurl**|`string`|Public base URL of the API for email links that hit the API directly<br/>|no|
|**CompanyName**|`string`|Company display name<br/>|no|
|**CompanyAddress**|`string`|Company mailing address<br/>|no|
|**Corporation**|`string`|Legal corporation name<br/>|no|
|**LogoURL**|`string`|Hero logo URL displayed in the email body<br/>|no|
|**HeaderLogoURL**|`string`|Small logo or icon displayed in the top header bar<br/>|no|
|**Copyright**|`string`|Copyright override for email footers; when empty the template renders a dynamic notice from Corporation and the current year<br/>|no|
|**TroubleText**|`string`|Fallback help text shown below action buttons; {ACTION} is replaced with the button text at render time<br/>|no|
|**TermsURL**|`string`|Terms of service link for email footers<br/>|no|
|**PrivacyURL**|`string`|Privacy policy link for email footers<br/>|no|
|**UnsubscribeURL**|`string`|Unsubscribe link override for email footers; when empty the template constructs one from ProductURL and the recipient email<br/>|no|
|**HeaderText**|`string`|Text displayed in the upper-right corner of the modern theme header<br/>|no|
|**CardStyle**|`string`|Card visual style<br/>Enum: `"flat"`, `"elevated"`<br/>|no|
|**BodyBackgroundColor**|`string`|Outer page background color<br/>|no|
|**CardBackgroundColor**|`string`|Card container background color<br/>|no|
|**HeroBackgroundColor**|`string`|Hero banner section background color<br/>|no|
|**ButtonColor**|`string`|Call-to-action button background color<br/>|no|
|**ButtonTextColor**|`string`|Call-to-action button text color<br/>|no|
|**HeadingColor**|`string`|Heading and title text color<br/>|no|
|**TextColor**|`string`|Body paragraph text color<br/>|no|
|**FooterTextColor**|`string`|Muted text color for headers footers and secondary content<br/>|no|
|**AccentBorderColor**|`string`|Decorative accent color applied to borders only<br/>|no|
|**Tagline**|`string`|Short descriptive footer line rendered above the social row in modern themes<br/>|no|

**Additional Properties:** not allowed   
   
<a name="defssystempaymentreminderconfig"></a>
### $defs/system\.PaymentReminderConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**paymentmethodinterval**|`integer`|Days after org creation before marking for deletion<br/>Default: `30`<br/>||
|**deletiondays**|`integer`|Days between marking and actual deletion<br/>Default: `7`<br/>||
|**enabled**|`boolean`|Whether the payment reminder listener is enabled<br/>Default: `false`<br/>||
|**dryrun**|`boolean`|If true only log organization IDs that would be processed<br/>Default: `true`<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "paymentmethodinterval": 30,
    "deletiondays": 7,
    "enabled": false,
    "dryrun": true
}
```

   
<a name="defssystemorganizationdeleteconfig"></a>
### $defs/system\.OrganizationDeleteConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**maxdeletesperrun**|`integer`|Maximum overdue organizations to delete per run<br/>Default: `25`<br/>||
|**enabled**|`boolean`|Whether the organization deletion listener is enabled<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "maxdeletesperrun": 25
}
```

   
<a name="defssystemintegrationlifecycleconfig"></a>
### $defs/system\.IntegrationLifecycleConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Whether the integration lifecycle sweep is enabled<br/>Default: `true`<br/>||
|**dryrun**|`boolean`|If true only log integration IDs and actions that would be dispatched<br/>Default: `true`<br/>||
|**maxperrun**|`integer`|Maximum integrations to evaluate per run<br/>Default: `100`<br/>||

**Additional Properties:** not allowed   
**Example**

```json
{
    "enabled": true,
    "dryrun": true,
    "maxperrun": 100
}
```

   
<a name="defsworkflowsconfig"></a>
## $defs/workflows\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**cel**](#defsworkflowscelconfig)|`object`|||
|[**gala**](#defsworkflowsgalaconfig)|`object`|||

**Additional Properties:** not allowed   
**Example**

```json
{
    "cel": {},
    "gala": {}
}
```

   
<a name="defsworkflowscelconfig"></a>
### $defs/workflows\.CELConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**timeout**|`integer`|||
|**costlimit**|`integer`|||
|**interruptcheckfrequency**|`integer`|||
|**parserrecursionlimit**|`integer`|||
|**parserexpressionsizelimit**|`integer`|||
|**comprehensionnestinglimit**|`integer`|||
|**extendedvalidations**|`boolean`|||
|**optionaltypes**|`boolean`|||
|**identifierescapesyntax**|`boolean`|||
|**crosstypenumericcomparisons**|`boolean`|||
|**macrocalltracking**|`boolean`|||
|**evaloptimize**|`boolean`|||
|**trackstate**|`boolean`|||

**Additional Properties:** not allowed   
   
<a name="defsworkflowsgalaconfig"></a>
### $defs/workflows\.GalaConfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**workercount**|`integer`|||
|**maxretries**|`integer`|||
|**failonenqueueerror**|`boolean`|||
|**queuename**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defshandlerscloudflareconfig"></a>
## $defs/handlers\.CloudflareConfig: object

CloudflareConfig contains configuration for Cloudflare integration.


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled toggles the Cloudflare snapshot handler<br/>||
|**apitoken**|`string`|APIToken is the API token used for Cloudflare client initialization<br/>||
|**accountid**|`string`|AccountID is the Cloudflare account ID to use for snapshot operations<br/>||
|**clientid**|`string`|ClientID is the Cloudflare Access client ID for shortlink API requests<br/>||
|**clientsecret**|`string`|ClientSecret is the Cloudflare Access client secret for shortlink API requests<br/>||

**Additional Properties:** not allowed   
   
<a name="defsshortlinksconfig"></a>
## $defs/shortlinks\.Config: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**endpointurl**|`string`|||

**Additional Properties:** not allowed   
   
<a name="defsconfigbackfill"></a>
## $defs/config\.Backfill: object

Backfill configures one-time startup data backfill routines that populate fields introduced by recent
migrations for organizations and memberships that pre-date them


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|Enabled runs the backfill routines on server startup, this only setups up the main job, all configured backfills will run based on their settings<br/>||

**Additional Properties:** not allowed   

