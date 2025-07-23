# Compliance Automation Jobs

## Overview

The compliance automation job system is built on top of [Windmill](https://windmill.dev/) and consists of several key components that work together to enable automated compliance checks and tasks. The system is designed to be flexible, scalable, and integrated with the organization's compliance framework.

![Jobs System Architecture](images/jobs.png)

The architecture diagram above illustrates the key components and their relationships:

1. **Job Templates** form the foundation of the system, defining reusable job definitions that include:
   - Script source and platform
   - Default configurations
   - Base schedules
   These templates are stored in both the core database and Windmill as flows.

2. **Scheduled Jobs** are instances created from templates that:
   - Reference a specific Job Template
   - Can override configurations and schedules
   - Can be linked to specific Job Runners
   - Are synchronized with Windmill's scheduler

3. **Job Runners** are execution environments that:
   - Authenticate using Job Runner Tokens
   - Execute jobs based on their assigned schedules
   - Report execution status and results back to the system

4. **Job Results** capture execution outcomes:
   - Link back to the Scheduled Job
   - Store execution metrics (start/end times, exit codes)
   - Maintain references to output files and logs
   - Track success/failure status

The system maintains bidirectional synchronization with Windmill, ensuring that job definitions, schedules, and results are consistent across both platforms. Schema hooks are used to update data into Windmill after a successful transaction into the database, and flow success and failure scripts will be used to get the results back and update the job results table.

## Components

### Job Templates

Job Templates (`JobTemplate`) are the core objects that define the jobs available for compliance automation. Each Job Template stores all the necessary information to run a job:

- **Title**: A human-readable name for the job
- **Description**: A description of what the job does
- **Platform**: The runtime environment or language for the script (e.g., `golang`, `typescript`, etc.). Currently the enums we have setup only support `Golang` (tested) and `Typescript` (not tested).
- **Download URL**: A URL pointing to the raw script source, which can be fetched and wrapped into a Windmill flow
- **Windmill Path**: The internal path in Windmill where the job's flow is stored (not exposed via GraphQL API) and include the organizations name is the path folder (e.g `f/acme_corp/my_awesome_script`)
- **Configuration**: Optional JSON configuration that can be used to template the job
- **Cron Schedule**: Optional default cron schedule (6-field syntax) used when creating scheduled jobs

When a Job Template is created or updated:
1. The system validates the information
2. If Windmill integration is enabled, creates/updates a corresponding Windmill flow
3. The flow path is stored on the template for future reference

### Scheduled Jobs

Scheduled Jobs (`ScheduledJob`) are instances of Job Templates configured to run on a schedule. They represent the actual jobs that will be executed in your environment. Key fields include:

- **Job Template ID**: Reference to the Job Template being scheduled
- **Active**: Whether the job is currently active
- **Configuration**: Optional JSON configuration that overrides the template's configuration
- **Cron Schedule**: Optional cron expression (6-field syntax) that overrides the template's schedule
- **Job Runner ID**: Optional reference to a specific runner that should execute this job

### Job Runners

Job Runners (`JobRunner`) are the execution environments where jobs run. They can be organization-specific or shared runners. Key fields include:

- **Name**: Human-readable name for the runner
- **Status**: Current status (ONLINE/OFFLINE)
- **IP Address**: The runner's IP address (immutable and unique)
- **Owner ID**: The organization that owns this runner (if organization-specific)

### Job Runner Tokens

Job Runner Tokens (`JobRunnerToken`) are used to authenticate and authorize job runners. These tokens are essential for secure communication between runners and the core system. Each token includes:

- **Token**: A unique, immutable string prefixed with "runner_" (automatically generated)
- **Expiration**: Optional expiration time (tokens don't expire by default)
- **Last Used**: Timestamp of the token's last usage
- **Status**:
  - **Is Active**: Whether the token is currently active
  - **Revoked At**: When the token was revoked (if applicable)
  - **Revoked By**: User who revoked the token
  - **Revoked Reason**: Reason for revocation

Key features:
- Tokens are organization-scoped
- Each token is linked to a specific job runner
- Tokens can be revoked at any time
- Activity tracking through last_used_at field
- Support for token expiration (optional)

Security considerations:
- Tokens are immutable once created
- Revoked tokens cannot be reactivated
- Token validation checks for expiration and active status
- Organization-level access control

### Job Results

Job Results (`JobResult`) track the execution outcomes of scheduled jobs. Each result includes:

- **Scheduled Job ID**: Reference to the executed job
- **Status**: Execution status (SUCCESS/FAILED/PENDING/CANCELED)
- **Exit Code**: The script's exit code (if applicable)
- **Started At**: When the job started executing
- **Finished At**: When the job finished executing
- **File ID**: Reference to output/logs file
- **Owner ID**: The organization that owns this result

## Development
1. Start windmill, run `task docker:windmill` to startup windmill, this will start the minimum set of components.
    If you want all the extra components you can run `task docker:windmill:full` to start everything in the compose file
    ```bash
    task docker:windmill
    ```
1. Login to [windmill](http://localhost:8090/) and create a workspace
1. Copy the windmill example config and add the license keys**
    ```bash
    cp docker/configs/windmill/.env-example docker/configs/windmill/.env
    ```
    This doesn't seem to quite work, and you may have to manually enter the enterprise key in [super-admin instance settings](http://localhost:8090/#superadmin-settings). Set both the enterprise key and set as a `non-prod` instance
3. Get a [token](http://localhost:8090/workspace_settings#user-settings) from user settings
4. Ensure the following settings are in your `.config.yaml`
    ```yaml
    entConfig:
        windmill:
            enabled: true
            baseURL: "http://localhost:8090"
            workspace: "test" # set to the workspace you created
            token: "" # get a token from windmill after it starts up from user settings and add here
            defaultTimeout: "30s"
            onFailureScript: "" # leave empty until these scripts are setup
            onSuccessScript: "" # leave empty until these scripts are setup
    ```
5. Run the normal `task run-dev` to setup the core api
6. From `core-startup` run the following to seed some test job templates and scheduled jobs:
    ```bash
    task setup
    task startup:create:jobs
    ```

## Job Execution Flow Design

The below is what is intended to happen, it is not all currently implemented

1. When a Scheduled Job's cron schedule triggers:
   - Windmill initiates the job execution
   - The job is assigned to a runner
   - The job status is set to PENDING

2. During execution:
   - The runner executes the job's script
   - Output and logs are captured
   - Start time is recorded

3. (TODO) After execution:
   - End time is recorded
   - Exit code is captured
   - Status is updated (SUCCESS/FAILED)
   - Results are stored in the job_results table
   - Output/logs are stored and linked via file_id

## Overall TODOs

### MVP
1. **Results** - results need to be posted back to the openlane API to the job results table
    1. Add the create resolver, this should be locked down to only system admin tokens
    1. Add success and failure scripts that can be used in both dev + production
    1. Example success script (this is pseudo code, don't expect it to work)
        ```go
        func main() (interface{}, error) {
            v, err := wmill.GetVariable("u/admin/openlane_system_admin_token")
            if err != nil {
                return nil, err
            }

            // something like this, need to get the typed variables instead
            payload := map[string]any{
                "input": map[string]any{
                    "status":"SUCCESS"
                }
            }

            jsonData, err := json.Marshal(payload)
            if err != nil {
                return nil, err
            }

            // and then we should using either httpling, or the openlane client
            // and adding the authorization header
            resp, err := http.Post(
                "https://api.theopenlane.io/query",
                "application/json",
                bytes.NewBuffer(jsonData),
            )
            if err != nil {
                return nil, err
            }

            defer resp.Body.Close()

            return 1, nil
        }
        ```
    1. Fix our client to use these because the syntax is incorrect in `pkg/windmill/client.go` and they aren't being added correctly - works currently with the script names set to empty in config. I started trying to use the `windmill-go-client` and its decent for actually getting the correct types from their `api` package, but the client itself isn't worth it, the one thing I got to work gave me a success response but was silently failing
1. **Openlane Agent** - Both the agent as well as flow for proxy requests
    1. Wrapped windmill agent allowing for customers to run their jobs on their own infrastructure should include a binary + docker image for both arm + x86 arch
    1. Agent should take a `job_runner` token, and register with windmill via a proxy. On startup, it will send a request to a REST endpoint, to allow the registration of the agent with windmill on behalf of the runner
    1. Start of a repo was [here](https://github.com/theopenlane/agent) although this was prior to choosing windmill
    1. We need to should ensure customer runners can only be used by that customer, look at using [queues/tags in windmill](https://www.windmill.dev/docs/core_concepts/worker_groups#set-tags-to-assign-specific-queues) when an agent is registered
1. **Update of scheduled jobs**
    1. Right now the hook for scheduled jobs creates the schedule in windmill but it is never updated
    1. If active is set to false, it should disable the job
    1. If the cron is changed, the schedule in windmill should change
1. **Deletion from windmill**
    1. If a job template or scheduled job is deleted from openlane, it should be deleted from windmill as well
    1. Add prevention of deletion of system-owned job templates if in-use but a customer scheduled job
1. **Testing**
    1. Test update paths for job templates + schedules in windmill, only creation has been tested
    1. Test deletion path + add hooks to ensure templates + schedules get deleted from windmill if they are deleted from the api
    1. Add tests for scheduled job -> parent access from control
1. **Cli commands** for each of the job related schemas
1. **Review Permissions** - right now most if not all of these objects are orgOwned, meaning any one in the org can view the job templates, scheduled jobs, and results. Scheduled jobs have an additional object owned permissions allowing edit if you have edit access to the parent control/subcontrols

### Nice to Have
1. **Example Scripts** - We should some more [example](https://github.com/theopenlane/jobs-examples) scripts (things we would use for our own audit) to examples to get started
1. **Job runner assignment** - instead of using/requiring runner ID we should fall back to being able to assign to any worker in the organization instead, look at using [queues/tags in windmill](https://www.windmill.dev/docs/core_concepts/worker_groups#set-tags-to-assign-specific-queues). This should mean a customer can create a scheduled job without a runner id, it can use any runner in their queue
1. **Tags** for customer openlane agents - allow them to add their own tags, e.g. `meow-jobs` to their agent configuration, and add a corresponding `agent_tags` field to their scheduled job, allowing them to target an agent by tag and not have to use a runner id
1. Return a **warning** on creation of scheduled job if a customer does not have any job runners in their organization
1. Add **priority field** to scheduled job, allow customers to prioritize certain jobs higher than others
1. Allow for other **triggers** (e.g. webhooks) instead of always running on a schedule
1. **Log retention** policies for jobs (and storage add-on)
1. Add `docker` and `bash` as **supported platform types** for job templates
1. **Openlane workers**, allowing customers to not provide a runner and use our workers
    1. Add field to scheduled job indicating `self_hosted` or `hosted`
    1. Consider restricting to only openlane provided scripts
    1. Add ability to store secrets for customers to use in their jobs, would require storage of secrets in windmill, and return a secret path to use in scripts
1. **Dev Environment Setup Automation**, right now there are a few manual steps you have to do in setting up a dev environment. With the `SUPERADMIN_SECRET` we should be able to automate the creation of a token and a workspace. Also should debug why the `LICENSE_KEY` doesn't seem to always correctly pull from the environment on startup