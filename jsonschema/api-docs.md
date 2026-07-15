# object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**domain**|`string`|||
|**refreshinterval**|`integer`|||
|[**server**](#server)|`object`||yes|
|[**entconfig**](#entconfig)|`object`|||
|[**auth**](#auth)|`object`||yes|
|[**authz**](#authz)|`object`||yes|
|[**db**](#db)|`object`||yes|
|[**jobqueue**](#jobqueue)|`object`|||
|[**redis**](#redis)|`object`|||
|[**sessions**](#sessions)|`object`|||
|[**totp**](#totp)|`object`|||
|[**ratelimit**](#ratelimit)|`object`|||
|[**ratelimitunmatched**](#ratelimitunmatched)|`object`|||
|[**objectstorage**](#objectstorage)|`object`|||
|[**subscription**](#subscription)|`object`|||
|[**keywatcher**](#keywatcher)|`object`|||
|[**integrations**](#integrations)|`object`|||
|[**workflows**](#workflows)|`object`|||
|[**cloudflare**](#cloudflare)|`object`|||
|[**shortlinks**](#shortlinks)|`object`|||
|[**backfill**](#backfill)|`object`|||

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

<a name="server"></a>
## server: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**dev**|`boolean`||no|
|**listen**|`string`||yes|
|**metricsport**|`string`||no|
|**shutdowngraceperiod**|`integer`||no|
|**readtimeout**|`integer`||no|
|**writetimeout**|`integer`||no|
|**idletimeout**|`integer`||no|
|**readheadertimeout**|`integer`||no|
|[**tls**](#servertls)|`object`||no|
|[**cors**](#servercors)|`object`||no|
|[**secure**](#serversecure)|`object`||no|
|[**cachecontrol**](#servercachecontrol)|`object`||no|
|[**mime**](#servermime)|`object`||no|
|[**graphpool**](#servergraphpool)|`object`||no|
|**enablegraphextensions**|`boolean`||no|
|**enablegraphsubscriptions**|`boolean`||no|
|**complexitylimit**|`integer`||no|
|**maxresultlimit**|`integer`||no|
|[**csrfprotection**](#servercsrfprotection)|`object`||no|
|**secretmanager**|`string`||no|
|**defaulttrustcenterdomain**|`string`||no|
|**trustcentercnametarget**|`string`||no|
|**trustcenterpreviewzoneid**|`string`||no|
|**notificationlookbackdays**|`integer`||no|

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

<a name="servertls"></a>
### server\.tls: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**certfile**|`string`|||
|**certkey**|`string`|||
|**autocert**|`boolean`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**maxworkers**|`integer`|||

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
|**questionnaireproducturl**|`string`|||

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
|**devmode**|`boolean`|||

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`||no|
|[**token**](#authtoken)|`object`||yes|
|[**supportedproviders**](#authsupportedproviders)|`string[]`||no|
|[**providers**](#authproviders)|`object`||no|
|[**supportaccess**](#authsupportaccess)|`object`||no|

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**redirecturl**|`string`|||
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
<a name="authsupportaccess"></a>
### auth\.supportaccess: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**email**|`string`|||
|**displayname**|`string`|||
|**subjectid**|`string`|||
|**password**|`string`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**issuerurl**|`string`|||
|**discoveryendpoint**|`string`|||
|**redirecturl**|`string`|||
|**alloweddomain**|`string`|||

**Additional Properties:** not allowed  
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
|**modulefile**|`string`|path to the fga module file<br/>|no|
|[**credentials**](#authzcredentials)|`object`||no|
|**maxbatchwritesize**|`integer`|maximum number of writes per batch in a transaction<br/>|no|
|**enableparentcontext**|`boolean`|disables the automatic addition of parent context tuples<br/>|no|
|[**parentcontextskipkinds**](#authzparentcontextskipkinds)|`string[]`||no|
|[**parentcontextconditions**](#authzparentcontextconditions)|`array`||no|

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
<a name="authzparentcontextskipkinds"></a>
### authz\.parentcontextskipkinds: array

**Items**

**Item Type:** `string`  
<a name="authzparentcontextconditions"></a>
### authz\.parentcontextconditions: array

**Items**

**Example**

```json
[
    {
        "context": {}
    }
]
```

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
|[**ReindexerIndexNames**](#jobqueueriverconfreindexerindexnames)|`string[]`|||
|**ReindexerTimeout**|`integer`|||
|**RescueStuckJobsAfter**|`integer`|||
|**RetryPolicy**||||
|**Schema**|`string`|||
|**SoftStopTimeout**|`integer`|||
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
<a name="jobqueueriverconfreindexerindexnames"></a>
#### jobqueue\.riverconf\.ReindexerIndexNames: array

**Items**

**Item Type:** `string`  
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
<a name="ratelimitunmatched"></a>
## ratelimitunmatched: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**options**](#ratelimitunmatchedoptions)|`array`|||
|[**headers**](#ratelimitunmatchedheaders)|`string[]`|||
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

<a name="ratelimitunmatchedoptions"></a>
### ratelimitunmatched\.options: array

**Items**

**Example**

```json
[
    {}
]
```

<a name="ratelimitunmatchedheaders"></a>
### ratelimitunmatched\.headers: array

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
|[**s3**](#objectstorageproviderss3)|`object`|||
|[**r2**](#objectstorageprovidersr2)|`object`|||
|[**disk**](#objectstorageprovidersdisk)|`object`|||
|[**database**](#objectstorageprovidersdatabase)|`object`|||

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
<a name="objectstorageprovidersr2"></a>
#### objectstorage\.providers\.r2: object

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
|[**credentials**](#objectstorageprovidersr2credentials)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "credentials": {}
}
```

<a name="objectstorageprovidersr2credentials"></a>
##### objectstorage\.providers\.r2\.credentials: object

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

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**keydir**|`string`|||

**Additional Properties:** not allowed  
<a name="integrations"></a>
## integrations: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**consoleintegrationpath**|`string`|||
|[**awssecurityhub**](#integrationsawssecurityhub)|`object`|||
|[**cloudflareruntime**](#integrationscloudflareruntime)|`object`|||
|[**githubapp**](#integrationsgithubapp)|`object`|||
|[**slack**](#integrationsslack)|`object`|||
|[**slackruntime**](#integrationsslackruntime)|`object`|||
|[**googledrive**](#integrationsgoogledrive)|`object`|||
|[**googleworkspace**](#integrationsgoogleworkspace)|`object`|||
|[**azureentraid**](#integrationsazureentraid)|`object`|||
|[**microsoftteams**](#integrationsmicrosoftteams)|`object`|||
|[**onedrive**](#integrationsonedrive)|`object`|||
|[**oidclocal**](#integrationsoidclocal)|`object`|||
|[**email**](#integrationsemail)|`object`||yes|
|[**paymentreminder**](#integrationspaymentreminder)|`object`|||
|[**organizationdelete**](#integrationsorganizationdelete)|`object`|||

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
    }
}
```

<a name="integrationsawssecurityhub"></a>
### integrations\.awssecurityhub: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**accesskeyid**|`string`|||
|**secretaccesskey**|`string`|||
|**arn**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationscloudflareruntime"></a>
### integrations\.cloudflareruntime: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**apitoken**|`string`|Cloudflare API token for the operator-owned account<br/>||
|**accountid**|`string`|Cloudflare account ID for the operator-owned account<br/>||
|[**domainscan**](#integrationscloudflareruntimedomainscan)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "domainscan": {}
}
```

<a name="integrationscloudflareruntimedomainscan"></a>
#### integrations\.cloudflareruntime\.domainscan: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**nonvendorcategories**](#integrationscloudflareruntimedomainscannonvendorcategories)|`string[]`|||
|[**deniedvendornames**](#integrationscloudflareruntimedomainscandeniedvendornames)|`string[]`|||
|**scanttl**|`integer`|||

**Additional Properties:** not allowed  
<a name="integrationscloudflareruntimedomainscannonvendorcategories"></a>
##### integrations\.cloudflareruntime\.domainscan\.nonvendorcategories: array

**Items**

**Item Type:** `string`  
<a name="integrationscloudflareruntimedomainscandeniedvendornames"></a>
##### integrations\.cloudflareruntime\.domainscan\.deniedvendornames: array

**Items**

**Item Type:** `string`  
<a name="integrationsgithubapp"></a>
### integrations\.githubapp: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**appid**|`string`|||
|**privatekey**|`string`|||
|**webhooksecret**|`string`|||
|**appslug**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsslack"></a>
### integrations\.slack: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**appid**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsslackruntime"></a>
### integrations\.slackruntime: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**webhookURL**|`string`|Slack incoming webhook URL for fire-and-forget system notifications<br/>||
|**botToken**|`string`|Bot User OAuth Token for full Web API access to the platform workspace<br/>||
|**defaultChannel**|`string`|Default channel id for system messages when no explicit channel is provided<br/>||

**Additional Properties:** not allowed  
<a name="integrationsgoogledrive"></a>
### integrations\.googledrive: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsgoogleworkspace"></a>
### integrations\.googleworkspace: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsazureentraid"></a>
### integrations\.azureentraid: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**defaulttenant**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsmicrosoftteams"></a>
### integrations\.microsoftteams: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsonedrive"></a>
### integrations\.onedrive: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**redirecturl**|`string`|||
|**contentmode**|`string`|||
|**applicationid**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsoidclocal"></a>
### integrations\.oidclocal: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**discoveryurl**|`string`|||
|**redirecturl**|`string`|||

**Additional Properties:** not allowed  
<a name="integrationsemail"></a>
### integrations\.email: object

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
<a name="integrationspaymentreminder"></a>
### integrations\.paymentreminder: object

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

<a name="integrationsorganizationdelete"></a>
### integrations\.organizationdelete: object

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

<a name="workflows"></a>
## workflows: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|[**cel**](#workflowscel)|`object`|||
|[**gala**](#workflowsgala)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "cel": {},
    "gala": {}
}
```

<a name="workflowscel"></a>
### workflows\.cel: object

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
<a name="workflowsgala"></a>
### workflows\.gala: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**workercount**|`integer`|||
|**maxretries**|`integer`|||
|**failonenqueueerror**|`boolean`|||
|**queuename**|`string`|||

**Additional Properties:** not allowed  
<a name="cloudflare"></a>
## cloudflare: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**apitoken**|`string`|||
|**accountid**|`string`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||

**Additional Properties:** not allowed  
<a name="shortlinks"></a>
## shortlinks: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||
|**clientid**|`string`|||
|**clientsecret**|`string`|||
|**endpointurl**|`string`|||

**Additional Properties:** not allowed  
<a name="backfill"></a>
## backfill: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**enabled**|`boolean`|||

**Additional Properties:** not allowed  

