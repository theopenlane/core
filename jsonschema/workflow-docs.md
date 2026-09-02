# Workflow Definition

Schema for Openlane workflow definitions


**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|**name**|`string`|||
|**description**|`string`|||
|**schemaType**|`string`|||
|**workflowKind**|`string`|Enum: `"APPROVAL"`, `"LIFECYCLE"`, `"NOTIFICATION"`<br/>||
|**approvalSubmissionMode**|`string`|Enum: `"MANUAL_SUBMIT"`, `"AUTO_SUBMIT"`<br/>||
|**approvalTiming**|`string`|Enum: `"PRE_COMMIT"`, `"POST_COMMIT"`<br/>||
|**version**|`string`|||
|[**targets**](#defsworkflowselector)|`object`|||
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

   
<a name="defsworkflowselector"></a>
## $defs/WorkflowSelector: object

**Properties**

|Name|Type|Description|Required|
|----|----|-----------|--------|
|[**tagIds**](#defsworkflowselectortagids)|`string[]`|||
|[**groupIds**](#defsworkflowselectorgroupids)|`string[]`|||
|[**objectTypes**](#defsworkflowselectorobjecttypes)|`string[]`|||

**Additional Properties:** not allowed   
   
<a name="defsworkflowselectortagids"></a>
### $defs/WorkflowSelector\.tagIds\[\]: array

**Items**

**Item Type:** `string`   
   
<a name="defsworkflowselectorgroupids"></a>
### $defs/WorkflowSelector\.groupIds\[\]: array

**Items**

**Item Type:** `string`   
   
<a name="defsworkflowselectorobjecttypes"></a>
### $defs/WorkflowSelector\.objectTypes\[\]: array

**Items**


The object type the workflow applies to

**Item Type:** `string`   
**Item Enum:** `"ActionPlan"`, `"Assessment"`, `"AssessmentResponse"`, `"Campaign"`, `"CampaignTarget"`, `"Control"`, `"Evidence"`, `"Finding"`, `"IdentityHolder"`, `"InternalPolicy"`, `"Platform"`, `"Procedure"`, `"Remediation"`, `"Risk"`, `"Subcontrol"`, `"Task"`, `"Vulnerability"`   

