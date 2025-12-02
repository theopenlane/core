# object

Config contains the configuration for the core server


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**domain**|`string`|Domain provides a global domain value for other modules to inherit<br/>||
|**refreshinterval**|`integer`|RefreshInterval determines how often to reload the config<br/>||
|[**server**](#server)|`object`|Server settings for the echo server<br/>|yes|
|[**entconfig**](#entconfig)|`object`|||
|[**auth**](#auth)|`object`|Auth settings including oauth2 providers and token configuration<br/>|yes|
|[**authz**](#authz)|`object`||yes|
|[**db**](#db)|`object`||yes|
|[**jobqueue**](#jobqueue)|`object`|||
|[**redis**](#redis)|`object`|||
|[**tracer**](#tracer)|`object`|||
|[**email**](#email)|`object`|||
|[**sessions**](#sessions)|`object`|||
|[**totp**](#totp)|`object`|||
|[**ratelimit**](#ratelimit)|`object`|||
|[**objectstorage**](#objectstorage)|`object`|||
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
            "cloudflarer2": {
                "credentials": {}
            },
            "gcs": {
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
|[**cors**](#servercors)|`object`||no|
|[**secure**](#serversecure)|`object`||no|
|[**redirects**](#serverredirects)|`object`||no|
|[**cachecontrol**](#servercachecontrol)|`object`||no|
|[**mime**](#servermime)|`object`||no|
|[**graphpool**](#servergraphpool)|`object`|PondPool contains the settings for the goroutine pool<br/>|no|
|**enablegraphextensions**|`boolean`|EnableGraphExtensions enables the graph extensions for the graph resolvers<br/>|no|
|**enablegraphsubscriptions**|`boolean`|EnableGraphSubscriptions enables graphql subscriptions to the server using websockets or sse<br/>|no|
|**complexitylimit**|`integer`|ComplexityLimit sets the maximum complexity allowed for a query<br/>|no|
|**maxresultlimit**|`integer`|MaxResultLimit sets the maximum number of results allowed for a query<br/>|no|
|[**csrfprotection**](#servercsrfprotection)|`object`||no|
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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**prefixes**](#servercorsprefixes)|`object`|||
|[**alloworigins**](#servercorsalloworigins)|`string[]`|||
|**cookieinsecure**|`boolean`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**xssprotection**|`string`|||
|**contenttypenosniff**|`string`|||
|**xframeoptions**|`string`|||
|**hstspreloadenabled**|`boolean`|||
|**hstsmaxage**|`integer`|||
|**contentsecuritypolicy**|`string`|||
|**referrerpolicy**|`string`|||
|**cspreportonly**|`boolean`|||

**Additional Properties:** not allowed  
<a name="serverredirects"></a>
### server\.redirects: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**redirects**](#serverredirectsredirects)|`object`|||
|**code**|`integer`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**mimetypesfile**|`string`|||
|**defaultcontenttype**|`string`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**header**|`string`|||
|**cookie**|`string`|||
|**secure**|`boolean`|||
|**samesite**|`string`|||
|**cookiehttponly**|`boolean`|||
|**cookiedomain**|`string`|||
|**cookiepath**|`string`|||

**Additional Properties:** not allowed  
<a name="serverfieldlevelencryption"></a>
### server\.fieldlevelencryption: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**keyset**|`string`|||

**Additional Properties:** not allowed  
<a name="entconfig"></a>
## entconfig: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**entitytypes**](#entconfigentitytypes)|`string[]`|||
|[**summarizer**](#entconfigsummarizer)|`object`|||
|**maxpoolsize**|`integer`|||
|[**modules**](#entconfigmodules)|`object`|||
|**maxschemaimportsize**|`integer`|||
|[**emailvalidation**](#entconfigemailvalidation)|`object`|||
|[**billing**](#entconfigbilling)|`object`|||
|[**notifications**](#entconfignotifications)|`object`|||

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

<a name="entconfigentitytypes"></a>
### entconfig\.entitytypes: array

**Items**

**Item Type:** `string`  
<a name="entconfigsummarizer"></a>
### entconfig\.summarizer: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**type**|`string`|||
|[**llm**](#entconfigsummarizerllm)|`object`|||
|**maximumsentences**|`integer`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**provider**|`string`|||
|[**anthropic**](#entconfigsummarizerllmanthropic)|`object`|||
|[**cloudflare**](#entconfigsummarizerllmcloudflare)|`object`|||
|[**openai**](#entconfigsummarizerllmopenai)|`object`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**betaheader**|`string`|||
|**legacytextcompletion**|`boolean`|||
|**baseurl**|`string`|||
|**model**|`string`|||
|**apikey**|`string`|||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmcloudflare"></a>
##### entconfig\.summarizer\.llm\.cloudflare: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|||
|**apikey**|`string`|||
|**accountid**|`string`|||
|**serverurl**|`string`|||

**Additional Properties:** not allowed  
<a name="entconfigsummarizerllmopenai"></a>
##### entconfig\.summarizer\.llm\.openai: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**model**|`string`|||
|**apikey**|`string`|||
|**url**|`string`|||
|**organizationid**|`string`|||

**Additional Properties:** not allowed  
<a name="entconfigmodules"></a>
### entconfig\.modules: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**usesandbox**|`boolean`|||

**Additional Properties:** not allowed  
<a name="entconfigemailvalidation"></a>
### entconfig\.emailvalidation: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**enableautoupdatedisposable**|`boolean`|||
|**enablegravatarcheck**|`boolean`|||
|**enablesmtpcheck**|`boolean`|||
|[**allowedemailtypes**](#entconfigemailvalidationallowedemailtypes)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "allowedemailtypes": {}
}
```

<a name="entconfigemailvalidationallowedemailtypes"></a>
#### entconfig\.emailvalidation\.allowedemailtypes: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**disposable**|`boolean`|||
|**free**|`boolean`|||
|**role**|`boolean`|||

**Additional Properties:** not allowed  
<a name="entconfigbilling"></a>
### entconfig\.billing: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**requirepaymentmethod**|`boolean`|||
|[**bypassemaildomains**](#entconfigbillingbypassemaildomains)|`string[]`|||

**Additional Properties:** not allowed  
<a name="entconfigbillingbypassemaildomains"></a>
#### entconfig\.billing\.bypassemaildomains: array

**Items**

**Item Type:** `string`  
<a name="entconfignotifications"></a>
### entconfig\.notifications: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleurl**|`string`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**options**](#ratelimitoptions)|`array`|||
|[**headers**](#ratelimitheaders)|`string[]`|||
|**forwardedindexfrombehind**|`integer`|||
|**includepath**|`boolean`|||
|**includemethod**|`boolean`|||
|**keyprefix**|`string`|||
|**denystatus**|`integer`|||
|**denymessage**|`string`|||
|**sendretryafterheader**|`boolean`|||
|**dryrun**|`boolean`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**keys**](#objectstoragekeys)|`string[]`|||
|**maxsizemb**|`integer`|||
|**maxmemorymb**|`integer`|||
|**devmode**|`boolean`|||
|[**providers**](#objectstorageproviders)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "providers": {
        "s3": {
            "credentials": {}
        },
        "cloudflarer2": {
            "credentials": {}
        },
        "gcs": {
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
|[**s3**](#objectstorageproviderss3)|`object`|||
|[**cloudflarer2**](#objectstorageproviderscloudflarer2)|`object`|||
|[**gcs**](#objectstorageprovidersgcs)|`object`|||
|[**disk**](#objectstorageprovidersdisk)|`object`|||
|[**database**](#objectstorageprovidersdatabase)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "s3": {
        "credentials": {}
    },
    "cloudflarer2": {
        "credentials": {}
    },
    "gcs": {
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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**ensureavailable**|`boolean`|||
|**region**|`string`|||
|**bucket**|`string`|||
|**endpoint**|`string`|||
|**proxypresignenabled**|`boolean`|||
|**baseurl**|`string`|||
|[**credentials**](#objectstorageproviderss3credentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageproviderss3credentials"></a>
##### objectstorage\.providers\.s3\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**projectid**|`string`|||
|**accountid**|`string`|||
|**apitoken**|`string`|||

**Additional Properties:** not allowed  
<a name="objectstorageproviderscloudflarer2"></a>
#### objectstorage\.providers\.cloudflarer2: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**ensureavailable**|`boolean`|||
|**region**|`string`|||
|**bucket**|`string`|||
|**endpoint**|`string`|||
|**proxypresignenabled**|`boolean`|||
|**baseurl**|`string`|||
|[**credentials**](#objectstorageproviderscloudflarer2credentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageproviderscloudflarer2credentials"></a>
##### objectstorage\.providers\.cloudflarer2\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**projectid**|`string`|||
|**accountid**|`string`|||
|**apitoken**|`string`|||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersgcs"></a>
#### objectstorage\.providers\.gcs: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**ensureavailable**|`boolean`|||
|**region**|`string`|||
|**bucket**|`string`|||
|**endpoint**|`string`|||
|**proxypresignenabled**|`boolean`|||
|**baseurl**|`string`|||
|[**credentials**](#objectstorageprovidersgcscredentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersgcscredentials"></a>
##### objectstorage\.providers\.gcs\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**projectid**|`string`|||
|**accountid**|`string`|||
|**apitoken**|`string`|||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersdisk"></a>
#### objectstorage\.providers\.disk: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**ensureavailable**|`boolean`|||
|**region**|`string`|||
|**bucket**|`string`|||
|**endpoint**|`string`|||
|**proxypresignenabled**|`boolean`|||
|**baseurl**|`string`|||
|[**credentials**](#objectstorageprovidersdiskcredentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersdiskcredentials"></a>
##### objectstorage\.providers\.disk\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**projectid**|`string`|||
|**accountid**|`string`|||
|**apitoken**|`string`|||

**Additional Properties:** not allowed  
<a name="objectstorageprovidersdatabase"></a>
#### objectstorage\.providers\.database: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**ensureavailable**|`boolean`|||
|**region**|`string`|||
|**bucket**|`string`|||
|**endpoint**|`string`|||
|**proxypresignenabled**|`boolean`|||
|**baseurl**|`string`|||
|[**credentials**](#objectstorageprovidersdatabasecredentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersdatabasecredentials"></a>
##### objectstorage\.providers\.database\.credentials: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**projectid**|`string`|||
|**accountid**|`string`|||
|**apitoken**|`string`|||

**Additional Properties:** not allowed  
<a name="subscription"></a>
## subscription: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**privatestripekey**|`string`|||
|**stripewebhooksecret**|`string`|||
|[**stripewebhooksecrets**](#subscriptionstripewebhooksecrets)|`object`|||
|**stripewebhookurl**|`string`|||
|**stripebillingportalsuccessurl**|`string`|||
|**stripecancellationreturnurl**|`string`|||
|[**stripewebhookevents**](#subscriptionstripewebhookevents)|`string[]`|||
|**stripewebhookapiversion**|`string`|||
|**stripewebhookdiscardapiversion**|`string`|||

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

