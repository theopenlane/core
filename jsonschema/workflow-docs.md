# Workflow Definition

Schema for Openlane workflow definitions


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**name**|`string`|||
|**description**|`string`|||
|**schemaType**|`string`|||
|**workflowKind**|`string`|Enum: `"APPROVAL"`, `"LIFECYCLE"`, `"NOTIFICATION"`<br/>||
|**approvalSubmissionMode**|`string`|Enum: `"MANUAL_SUBMIT"`, `"AUTO_SUBMIT"`<br/>Defaults to `AUTO_SUBMIT` when omitted.||
|**version**|`string`|||
|[**targets**](#targets)|`object`|||
|[**triggers**](#triggers)|`array`|||
|[**conditions**](#conditions)|`array`|||
|[**actions**](#actions)|`array`|||
|[**metadata**](#metadata)|`object`|||

**Additional Properties:** not allowed  
**Example**

```json
{
    "targets": {},
    "triggers": [
        {
            "selector": {}
        }
    ],
    "conditions": [
        {}
    ],
    "actions": [
        {}
    ],
    "metadata": {}
}
```

<a name="targets"></a>
## targets: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**tagIds**](#targetstagids)|`string[]`|||
|[**groupIds**](#targetsgroupids)|`string[]`|||
|[**objectTypes**](#targetsobjecttypes)|`string[]`|||

**Additional Properties:** not allowed  
<a name="targetstagids"></a>
### targets\.tagIds\[\]: array

**Items**

**Item Type:** `string`  
<a name="targetsgroupids"></a>
### targets\.groupIds\[\]: array

**Items**

**Item Type:** `string`  
<a name="targetsobjecttypes"></a>
### targets\.objectTypes\[\]: array

**Items**


The object type the workflow applies to

**Item Type:** `string`  
**Item Enum:** `"ActionPlan"`, `"Campaign"`, `"CampaignTarget"`, `"Control"`, `"Evidence"`, `"IdentityHolder"`, `"InternalPolicy"`, `"Platform"`, `"Procedure"`, `"Subcontrol"`  
<a name="triggers"></a>
## triggers\[\]: array

**Items**

**Example**

```json
[
    {
        "selector": {}
    }
]
```

<a name="conditions"></a>
## conditions\[\]: array

**Items**

**Example**

```json
[
    {}
]
```

<a name="actions"></a>
## actions\[\]: array

**Items**

**Example**

```json
[
    {}
]
```

<a name="metadata"></a>
## metadata: object

**No properties.**

