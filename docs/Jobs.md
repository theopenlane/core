# Compliance Automation Jobs

## Setup Dev

1. Run `task docker:windmill:up` to startup windmill
1. Copy the windmill example config and add the license keys
    ```bash
    cp docker/configs/windmill/.env-example docker/configs/windmill/.env1
    ```
1. Get a [token](http://localhost:8090/workspace_settings#user-settings) from user settings
1. Ensure the following settings are in your `.config.yaml`
    ```yaml
    entConfig:
        windmill:
            enabled: true
            baseURL: "http://localhost:8090"
            workspace: "test"
            token: "" # get a token from windmill after it starts up from user settings and add here
            defaultTimeout: "30s"
            onFailureScript: "u/admin/glimmering_script"
            onSuccessScript: "u/admin/glimmering_script"
            folderName: test
    ```
1. Run the normal `task run-dev` to setup the core api
1. From `core-startup` run the following to seed some test job templates:
    ```bash
    task feeling:lucky
    task startup:create:jobs
    ```

## Details

## Job Templates

Job Templates are the core objects that define the jobs available for compliance automation. Each Job Template stores all the necessary information to run a job, including:

- **Title**: A human-readable name for the job.
- **Platform**: The runtime environment or language for the script (e.g., `golang`, `typescript` etc).
- **Description**: A description of what the job does.
- **Download URL**: A URL pointing to the raw script source, which can be fetched and wrapped into a Windmill flow.
- **Windmill Path**: The path in Windmill where the job's flow is stored and managed. This is not exposed via the graphql API and is only used internally
- **Other Metadata**: Such as cron schedule - this is used as a default when scheduling jobs if none is provided in the scheduled job

When a Job Template is created or updated, the system will automatically create or update a corresponding [Windmill](https://windmill.dev/) flow. This allows jobs to be scheduled, executed, and managed through the Windmill.

### How Job Templates Work

- When you create a new Job Template, you provide the script (or a download URL), name, platform, and other metadata.
- The system validates the information and, if Windmill integration is enabled, creates a Windmill flow for the job and the flow path is set on the template.
- Updates to the Job Template (such as changing the script, platform, or description) will update the corresponding Windmill flow.

## Scheduled Jobs

Scheduled Jobs are instances of Job Templates that are configured to run on a schedule (using cron syntax) or on demand. They represent the actual jobs that will be executed in your environment, based on the logic defined in a Job Template.

### Key Concepts

- **Job Template**: The reusable definition of a job (script, platform, description, etc), describe in the section above
- **Scheduled Job**: An instance of a Job Template, configured with a schedule and any job-specific parameters.

### How Scheduled Jobs Work

- When you create a Scheduled Job, you select a Job Template and specify a schedule (e.g., `"0 0 * * *"` for daily at midnight).
- The system creates a corresponding schedule in Windmill, pointing to the flow generated from the Job Template.
- You can provide arguments/parameters for the job run, which will be passed to the script when it executes.
- Scheduled Jobs can be enabled or disabled, and you can update their schedule or parameters at any time.
- Each time the schedule triggers, Windmill executes the associated flow with the provided parameters.

### Example Workflow

1. **Create a Job Template**: Define the script, platform, and metadata.
2. **Create a Scheduled Job**: Choose the Job Template, set the schedule, and provide any arguments.
3. **Execution**: Windmill runs the job according to the schedule, and results/logs are available in the Windmill UI.

### Fields in a Scheduled Job

- **Job Template**: Reference to the Job Template being scheduled.
- **Schedule**: Cron expression defining when the job runs.
- **Arguments**: Optional parameters to pass to the job.
- **Enabled**: Whether the job is active.
- **Windmill Path**: The Windmill schedule path (internal use).

### Notes

- If you update the Job Template (e.g., change the script), future runs of the Scheduled Job will use the updated logic.
- You can create multiple Scheduled Jobs from the same Job Template, each with different schedules or parameters.

For more information on cron syntax, see [crontab.guru](https://crontab.guru/).


## Future Work
- Results need to be posted back to the openlane API to the job results table
- Allow for other triggers (e.g. webhooks) instead of always running on a schedule
- Openlane agent wrapper
- Add success and failure scripts that can be used in both dev + production
- ^ Fix the API (maybe just switch to the go-client) because the syntax is incorrect