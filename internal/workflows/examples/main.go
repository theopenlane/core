// Package main provides a workflow demonstration seed script
//
// This script creates a complete workflow automation example showcasing:
// - Organization and user creation via API
// - Workflow definition with dual approval
// - Control that triggers the workflow
// - Workflow instances and assignments
//
// Run with: go run main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	api "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/workflows"
	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/ulids"
)

type WorkflowDemoSeed struct {
	// OrganizationID is the demo organization id
	OrganizationID string
	// Control is the created control payload
	Control *graphclient.CreateControl_CreateControl_Control
	// WorkflowDef is the created workflow definition
	WorkflowDef *graphclient.CreateWorkflowDefinition_CreateWorkflowDefinition_WorkflowDefinition
	// InstanceID is the created workflow instance id
	InstanceID string
}

type DemoConfig struct {
	// OrgID is the target organization id
	OrgID string
	// UseDefaultOrg selects the user's default organization
	UseDefaultOrg bool
}

type ScenarioRunner func(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, []string, error)

var (
	flagScenario       = flag.String("scenario", "", "Workflow scenario (dual-approval|queue-approval|webhook|slack-webhook|field-update|evidence-review|examples)")
	flagSlackWebhook   = flag.String("slack-webhook-url", "", "Slack webhook URL (required for slack-webhook and evidence-review scenarios)")
	flagOpenlaneAPIURL = flag.String("openlane-api-url", "", "Override Openlane API base URL")
	flagExamplesDir    = flag.String("examples-dir", "internal/workflows/examples", "Workflow definition JSON directory (used by examples scenario)")
	flagEmail          = flag.String("email", "", "Demo user email (env: WORKFLOW_DEMO_EMAIL)")
	flagPassword       = flag.String("password", "", "Demo user password (env: WORKFLOW_DEMO_PASSWORD)")
	flagOrgID          = flag.String("org-id", "", "Existing organization ID (env: WORKFLOW_ORG_ID)")
	flagUseDefaultOrg  = flag.Bool("use-default-org", false, "Use the user's default org instead of creating a new one (env: WORKFLOW_USE_DEFAULT_ORG)")
	flagAPIToken       = flag.String("api-token", "", "API token for authentication, bypasses login (env: OPENLANE_API_TOKEN)")
)

// main runs the workflow demo scenarios
func main() {
	ctx := context.Background()

	fmt.Println("=== Workflow Demo Seed Script ===")
	fmt.Println("This demonstrates the complete workflow automation lifecycle")
	fmt.Println("\nFlags: --scenario (dual-approval|queue-approval|webhook|slack-webhook|field-update|evidence-review|examples), --slack-webhook-url, --examples-dir")
	fmt.Println("       --email, --password, --org-id, --use-default-org")
	fmt.Println("Env fallback: WORKFLOW_SCENARIO, SLACK_WEBHOOK_URL/SLACK_WEBHOOK, OPENLANE_API_URL")
	fmt.Println("              WORKFLOW_DEMO_EMAIL, WORKFLOW_DEMO_PASSWORD, WORKFLOW_ORG_ID, WORKFLOW_USE_DEFAULT_ORG")
	flag.Parse()

	config := openlane.NewDefaultConfig()
	// Default to localhost to prevent accidental production usage
	baseURL := firstNonEmpty(*flagOpenlaneAPIURL, os.Getenv("OPENLANE_API_URL"), "http://localhost:17608")
	parsed, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("Invalid API URL %q: %v", baseURL, err)
	}
	config.BaseURL = parsed
	fmt.Printf("\nUsing API URL: %s\n", parsed.String())

	client, err := newClient(config.BaseURL)
	if err != nil {
		log.Fatalf("Failed to create API client: %v", err)
	}

	fmt.Println("\nSetting up test user...")
	email := firstNonEmpty(*flagEmail, os.Getenv("WORKFLOW_DEMO_EMAIL"), "mitb@theopenlane.io")
	password := firstNonEmpty(*flagPassword, os.Getenv("WORKFLOW_DEMO_PASSWORD"), "mattisthebest1234")

	registerInput := api.RegisterRequest{
		Email:     email,
		FirstName: "Matt",
		LastName:  "Anderson",
		Password:  password,
	}

	registerResp, err := client.Register(ctx, &registerInput)
	if err != nil {
		fmt.Printf("   User may already exist (register failed: %v), attempting login...\n", err)
	} else {
		fmt.Printf("   Registered new user: %s\n", email)
		if registerResp.Token != "" {
			tokenPreview := registerResp.Token
			if len(tokenPreview) > 20 {
				tokenPreview = tokenPreview[:20] + "..."
			}
			fmt.Printf("   Verifying user with token: %s\n", tokenPreview)

			verifyResp, err := client.VerifyEmail(ctx, &api.VerifyRequest{Token: registerResp.Token})
			if err != nil {
				log.Fatalf("Failed to verify user: %v", err)
			}
			fmt.Printf("   User verified: %s\n", verifyResp.Email)
		} else {
			fmt.Println("   No verification token returned (email verification may be disabled)")
		}
	}

	fmt.Println("\nAuthenticating...")
	loginInput := api.LoginRequest{
		Username: email,
		Password: password,
	}

	resp, err := client.Login(ctx, &loginInput)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("   ✓ Login successful")

	session, err := client.GetSessionFromCookieJar()
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}

	authClient, err := newClient(
		config.BaseURL,
		openlane.WithCredentials(openlane.Authorization{
			BearerToken: resp.AccessToken,
			Session:     session,
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create authenticated client: %v", err)
	}

	scenario := firstNonEmpty(*flagScenario, os.Getenv("WORKFLOW_SCENARIO"), "dual-approval")
	slackWebhookURL := firstNonEmpty(*flagSlackWebhook, os.Getenv("SLACK_WEBHOOK_URL"), os.Getenv("SLACK_WEBHOOK"))
	demo := DemoConfig{
		OrgID:         firstNonEmpty(*flagOrgID, os.Getenv("WORKFLOW_ORG_ID")),
		UseDefaultOrg: *flagUseDefaultOrg || envBool("WORKFLOW_USE_DEFAULT_ORG"),
	}

	demos := map[string]ScenarioRunner{
		"dual-approval": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runDualApprovalDemo(c, cfg, cl, d)
		},
		"queue-approval": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runQueueApprovalDemo(c, cfg, cl, d)
		},
		"webhook": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runWebhookDemo(c, cfg, cl, d)
		},
		"field-update": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runFieldUpdateDemo(c, cfg, cl, d)
		},
		"slack-webhook": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runSlackWebhookDemo(c, cfg, cl, d, slackWebhookURL)
		},
		"evidence-review": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runEvidenceReviewDemo(c, cfg, cl, d, slackWebhookURL)
		},
		"examples": func(c context.Context, cfg openlane.Config, cl *openlane.Client, d DemoConfig) (*WorkflowDemoSeed, []string, error) {
			return runExamplesDemo(c, cfg, cl, d, *flagExamplesDir)
		},
	}

	run, ok := demos[scenario]
	if !ok {
		log.Fatalf("Unknown WORKFLOW_SCENARIO %q. Valid options: dual-approval, queue-approval, webhook, slack-webhook, field-update, evidence-review, examples.", scenario)
	}

	fmt.Printf("\nRunning scenario: %s\n", scenario)

	seed, steps, err := run(ctx, config, authClient, demo)
	if err != nil {
		log.Fatalf("Error creating workflow demo: %v", err)
	}

	printSeedDetails(seed, steps)
}

// runDualApprovalDemo provisions a demo workflow with two approval steps
func runDualApprovalDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, []string, error) {
	seed, userResp, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}
	currentUserID := userResp.Self.ID

	fmt.Println("\n3. Creating workflow definition with dual approval...")
	fmt.Println("   Note: Using current user for both approval targets in this demo")

	targets1 := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeUser,
			ID:   currentUserID,
		},
	}

	approvalParams1 := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: targets1,
		},
		Required: ptr(true),
		Label:    "First Level Approval - Compliance Review",
		Fields:   []string{"status"},
	}

	params1Bytes, err := marshalParams("approval params (compliance)", approvalParams1)
	if err != nil {
		return nil, nil, err
	}

	targets2 := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeUser,
			ID:   currentUserID,
		},
	}

	approvalParams2 := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: targets2,
		},
		Required: ptr(true),
		Label:    "Second Level Approval - Category Sign-off",
		Fields:   []string{"category"},
	}

	params2Bytes, err := marshalParams("approval params (category)", approvalParams2)
	if err != nil {
		return nil, nil, err
	}

	notificationParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{
					Type: enums.WorkflowTargetTypeUser,
					ID:   currentUserID,
				},
			},
		},
		Channels: []enums.Channel{enums.ChannelInApp},
		Title:    "Control approval completed",
		Body:     "Workflow {{instance_id}} approved control {{object_id}}.",
	}

	notificationParamsBytes, err := marshalParams("notification params (completion)", notificationParams)
	if err != nil {
		return nil, nil, err
	}

	workflowDefDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeControl,
				Fields:     []string{"status", "category"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{
				Expression: "'status' in changed_fields || 'category' in changed_fields",
			},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeApproval.String(),
				Key:    "compliance_approval",
				Params: params1Bytes,
				When:   "'status' in changed_fields",
			},
			{
				Type:   enums.WorkflowActionTypeApproval.String(),
				Key:    "executive_approval",
				Params: params2Bytes,
				When:   "'category' in changed_fields",
			},
			{
				Type:   enums.WorkflowActionTypeNotification.String(),
				Key:    "completion_notification",
				Params: notificationParamsBytes,
			},
		},
	}

	workflowDefResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Control Status + Category Change - Dual Approval Workflow",
		Description:    ptr("Two-tier approval workflow for control status/category changes with notification on completion"),
		SchemaType:     string(enums.WorkflowObjectTypeControl),
		WorkflowKind:   enums.WorkflowKindApproval,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &workflowDefDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create workflow definition: %w", err)
	}
	seed.WorkflowDef = &workflowDefResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Triggers: When control status or category is updated\n")
	fmt.Printf("   - First approval: %s (%s)\n", userResp.Self.Email, approvalParams1.Label)
	fmt.Printf("   - Second approval: %s (%s)\n", userResp.Self.Email, approvalParams2.Label)
	fmt.Printf("   - On completion: Send notification\n")

	fmt.Println("\n4. Creating control to trigger workflow...")

	control, err := createControl(ctx, client, seed.OrganizationID,
		"Access Control and Authentication",
		"Ensure all users are authenticated and authorized before accessing sensitive data",
		"Access Control",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control

	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)
	fmt.Printf("   Initial status: %s\n", *seed.Control.Status)

	fmt.Println("\n5. Updating control fields to trigger workflow...")
	approvedStatus := enums.ControlStatusApproved
	updatedCategory := "Access Control (Updated)"
	updatedControl, err := client.UpdateControl(ctx, seed.Control.ID, graphclient.UpdateControlInput{
		Status:   &approvedStatus,
		Category: &updatedCategory,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update control fields: %w", err)
	}
	fmt.Printf("   Updated control status: %s -> %s\n", *seed.Control.Status, *updatedControl.UpdateControl.Control.Status)
	fmt.Printf("   Updated control category: %s -> %s\n", *seed.Control.Category, *updatedControl.UpdateControl.Control.Category)
	logControlSnapshot(ctx, client, seed.Control.ID)

	instance, err := waitForWorkflowInstance(ctx, client, seed, 2, 3*time.Second)
	if err != nil {
		return nil, nil, err
	}
	seed.InstanceID = instance.ID
	fmt.Printf("   Workflow instance started: %s (state: %s)\n", instance.ID, instance.State.String())

	fmt.Println("\n6. Checking for workflow assignments...")
	assignments, err := waitForPendingAssignments(ctx, client, seed, 2, 3*time.Second)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("   Found %d pending assignment(s) for the new workflow instance\n", len(assignments))
	for _, a := range assignments {
		label := "Unnamed"
		if a.Label != nil {
			label = *a.Label
		}
		fmt.Printf("   - %s (%s)\n", label, a.ID)
	}

	fmt.Printf("\n7. Approving %d assignments...\n", len(assignments))
	for i, a := range assignments {
		fmt.Printf("   Approving assignment %d/%d (%s)...\n", i+1, len(assignments), a.ID)
		if err := approveWorkflowAssignment(ctx, client, a.ID); err != nil {
			return nil, nil, fmt.Errorf("failed to approve assignment %s: %w", a.ID, err)
		}
		fmt.Printf("   ✓ Assignment %d approved\n", i+1)
		time.Sleep(1 * time.Second)
	}

	// Handle multi-stage workflows (e.g., subsequent approvals) by checking for newly created pending assignments.
	for round := 1; round <= 2; round++ {
		time.Sleep(1 * time.Second)
		more, err := listPendingAssignments(ctx, client, seed)
		if err != nil {
			return nil, nil, err
		}
		if len(more) == 0 {
			break
		}

		fmt.Printf("\n7.%d Found %d additional pending assignment(s); approving...\n", round, len(more))
		for i, a := range more {
			fmt.Printf("   Approving assignment %d/%d (%s)...\n", i+1, len(more), a.ID)
			if err := approveWorkflowAssignment(ctx, client, a.ID); err != nil {
				return nil, nil, fmt.Errorf("failed to approve assignment %s: %w", a.ID, err)
			}
			fmt.Printf("   ✓ Assignment %d approved\n", i+1)
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("\n8. Verifying workflow completion...")
	finalState, err := waitForInstanceState(ctx, client, seed.InstanceID, enums.WorkflowInstanceStateCompleted, 3, time.Second)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("   Workflow instance state: %s\n", finalState.String())

	steps := []string{
		"Initialized organization context",
		"Created dual-approval workflow (control status/category update trigger)",
		"Created control object",
		"Triggered workflow by updating control status and category",
		"Approved assignments programmatically",
		"Confirmed workflow completion",
	}

	return seed, steps, nil
}

// runQueueApprovalDemo creates approvals and leaves them pending for UI review.
func runQueueApprovalDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, []string, error) {
	seed, userResp, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}
	currentUserID := userResp.Self.ID

	fmt.Println("\n3. Creating workflow definition for queued approvals...")

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{
					Type: enums.WorkflowTargetTypeUser,
					ID:   currentUserID,
				},
			},
		},
		Required: ptr(true),
		Label:    "Control category approval",
		Fields:   []string{"category"},
	}

	paramsBytes, err := marshalParams("approval params (category)", approvalParams)
	if err != nil {
		return nil, nil, err
	}

	workflowDefDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeControl,
				Fields:     []string{"category"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{
				Expression: "'category' in changed_fields",
			},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeApproval.String(),
				Key:    "category_approval",
				Params: paramsBytes,
			},
		},
	}

	workflowDefResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Control Category Change - Approval Queue",
		Description:    ptr("Queue approvals when control category changes"),
		SchemaType:     string(enums.WorkflowObjectTypeControl),
		WorkflowKind:   enums.WorkflowKindApproval,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &workflowDefDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create workflow definition: %w", err)
	}
	seed.WorkflowDef = &workflowDefResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Trigger: category update\n")
	fmt.Printf("   - Approval target: %s\n", userResp.Self.Email)

	fmt.Println("\n4. Creating control to trigger workflow...")
	control, err := createControl(ctx, client, seed.OrganizationID,
		"Approval Queue Demo Control",
		"Control used to demonstrate queued approvals",
		"Queue Demo",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control
	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)

	fmt.Println("\n5. Updating control category to trigger workflow...")
	newCategory := "Approval Queue Demo"
	updatedControl, err := client.UpdateControl(ctx, seed.Control.ID, graphclient.UpdateControlInput{
		Category: &newCategory,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update control category: %w", err)
	}
	fmt.Printf("   Updated control category: %s -> %s\n", *seed.Control.Category, *updatedControl.UpdateControl.Control.Category)
	logControlSnapshot(ctx, client, seed.Control.ID)

	instance, err := waitForWorkflowInstance(ctx, client, seed, 2, 3*time.Second)
	if err != nil {
		return nil, nil, err
	}
	seed.InstanceID = instance.ID
	fmt.Printf("   Workflow instance started: %s (state: %s)\n", instance.ID, instance.State.String())

	fmt.Println("\n6. Checking for workflow assignments...")
	assignments, err := waitForPendingAssignments(ctx, client, seed, 2, 3*time.Second)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("   Found %d pending assignment(s)\n", len(assignments))
	for _, a := range assignments {
		label := "Unnamed"
		if a.Label != nil {
			label = *a.Label
		}
		fmt.Printf("   - %s (%s)\n", label, a.ID)
	}

	fmt.Println("\nAssignments are left pending for UI review.")
	fmt.Println("   - Visit https://console.theopenlane.io/workflows/assignments to approve.")

	steps := []string{
		"Created approval workflow (category update trigger)",
		"Created control object",
		"Triggered workflow by updating control category",
		"Left approval assignments pending",
	}

	return seed, steps, nil
}

// firstNonEmpty returns the first non-empty string in values
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// envBool parses a boolean value from an environment variable
func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

// printSeedDetails prints the seed identifiers and step checklist
func printSeedDetails(seed *WorkflowDemoSeed, steps []string) {
	fmt.Println("\n========================================")
	fmt.Println("   WORKFLOW AUTOMATION TEST COMPLETE")
	fmt.Println("========================================")
	fmt.Println("\nSuccessfully tested workflow automation:")
	for _, step := range steps {
		fmt.Printf("  ✓ %s\n", step)
	}

	fmt.Println("\n--- Created Entities ---")
	fmt.Printf("Organization ID:         %s\n", seed.OrganizationID)
	if seed.WorkflowDef != nil {
		fmt.Printf("Workflow Definition ID:  %s\n", seed.WorkflowDef.ID)
	}
	if seed.InstanceID != "" {
		fmt.Printf("Workflow Instance ID:    %s\n", seed.InstanceID)
	}
	if seed.Control != nil {
		fmt.Printf("Control ID:              %s\n", seed.Control.ID)
		fmt.Printf("Control RefCode:         %s\n", seed.Control.RefCode)
	}

	fmt.Println("\n--- View Results in UI ---")
	fmt.Println("Navigate to the control in ObjectWorkflowPanel to see:")
	fmt.Println("  - Completed workflow history")
	fmt.Println("  - Approved assignments with timestamps")
	fmt.Println("  - Workflow audit trail")
	fmt.Println("========================================")
	fmt.Println()
}

// newClient creates an Openlane API client for the provided base URL
func newClient(baseURL *url.URL, opts ...openlane.ClientOption) (*openlane.Client, error) {
	if baseURL != nil {
		opts = append([]openlane.ClientOption{openlane.WithBaseURL(baseURL.String())}, opts...)
	}

	return openlane.New(opts...)
}

// marshalParams encodes action params for demo setup and returns JSON bytes
func marshalParams(label string, params any) (json.RawMessage, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", label, err)
	}

	return data, nil
}

// createControl standardizes control creation for workflow demos.
func createControl(ctx context.Context, apiClient *openlane.Client, orgID, title, description, category string) (*graphclient.CreateControl_CreateControl_Control, error) {
	if strings.TrimSpace(orgID) == "" {
		return nil, fmt.Errorf("organization ID is required to create a control")
	}

	input := graphclient.CreateControlInput{
		RefCode:     fmt.Sprintf("CTL-%s", ulids.New().String()),
		Title:       ptr(title),
		Description: ptr(description),
		Status:      ptr(enums.ControlStatusNotImplemented),
		Source:      ptr(enums.ControlSourceUserDefined),
		Category:    ptr(category),
		OwnerID:     &orgID,
	}

	resp, err := apiClient.CreateControl(ctx, input)
	if err != nil {
		return nil, err
	}

	return &resp.CreateControl.Control, nil
}

const approveWorkflowAssignmentDocument = `mutation ApproveWorkflowAssignment($id: ID!) {
  approveWorkflowAssignment(id: $id) {
    workflowAssignment {
      id
      status
    }
  }
}`

type approveWorkflowAssignmentResponse struct {
	// ApproveWorkflowAssignment stores the mutation payload
	ApproveWorkflowAssignment struct {
		// WorkflowAssignment holds the approved assignment data
		WorkflowAssignment struct {
			// ID is the assignment identifier
			ID string `json:"id"`
			// Status is the assignment status after approval
			Status enums.WorkflowAssignmentStatus `json:"status"`
		} `json:"workflowAssignment"`
	} `json:"approveWorkflowAssignment"`
}

// approveWorkflowAssignment issues the approval mutation for an assignment
func approveWorkflowAssignment(ctx context.Context, apiClient *openlane.Client, assignmentID string) error {
	graphClient, ok := apiClient.GraphClient.(*graphclient.Client)
	if !ok || graphClient == nil {
		return fmt.Errorf("graph client does not support raw workflow assignment mutations")
	}

	var resp approveWorkflowAssignmentResponse
	vars := map[string]any{"id": assignmentID}

	if err := graphClient.Client.Post(ctx, "ApproveWorkflowAssignment", approveWorkflowAssignmentDocument, &resp, vars); err != nil {
		return err
	}

	if resp.ApproveWorkflowAssignment.WorkflowAssignment.ID == "" {
		return fmt.Errorf("approve workflow assignment returned empty result")
	}

	return nil
}

// runWebhookDemo demonstrates a webhook + notification workflow without approvals.
func runWebhookDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, []string, error) {
	seed, userResp, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}
	currentUserID := userResp.Self.ID

	fmt.Println("\n3. Creating webhook workflow...")

	webhookParams := workflows.WebhookActionParams{
		URL:    "https://example.com/webhook",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Payload: map[string]any{
			"event": "control_description_updated",
		},
		TimeoutMS: 5000,
	}
	webhookBytes, err := marshalParams("webhook params (control description)", webhookParams)
	if err != nil {
		return nil, nil, err
	}

	notificationParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{
					Type: enums.WorkflowTargetTypeUser,
					ID:   currentUserID,
				},
			},
		},
		Channels: []enums.Channel{enums.ChannelInApp},
		Title:    "Control description updated",
		Body:     "Workflow {{instance_id}} updated control {{object_id}} description.",
	}
	notificationBytes, err := marshalParams("notification params (control description)", notificationParams)
	if err != nil {
		return nil, nil, err
	}

	defDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeControl,
				Fields:     []string{"description"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "'description' in changed_fields"},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeWebhook.String(),
				Key:    "send_webhook",
				Params: webhookBytes,
			},
			{
				Type:   enums.WorkflowActionTypeNotification.String(),
				Key:    "notify_ops",
				Params: notificationBytes,
			},
		},
	}

	defResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Control Description Webhook Workflow",
		Description:    ptr("Sends a webhook + notification when control description changes"),
		SchemaType:     string(enums.WorkflowObjectTypeControl),
		WorkflowKind:   enums.WorkflowKindNotification,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &defDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create webhook workflow definition: %w", err)
	}
	seed.WorkflowDef = &defResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Trigger: description update\n")
	fmt.Printf("   - Action: webhook then notification\n")

	fmt.Println("\n4. Creating control to trigger workflow...")
	control, err := createControl(ctx, client, seed.OrganizationID,
		"Webhook Demo Control",
		"Initial description",
		"Automation",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control
	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)

	fmt.Println("\n5. Updating control description to trigger workflow...")
	newDescription := "Updated by webhook demo"
	updatedControl, err := client.UpdateControl(ctx, seed.Control.ID, graphclient.UpdateControlInput{
		Description: &newDescription,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update control description: %w", err)
	}
	fmt.Printf("   Updated control description to: %s\n", *updatedControl.UpdateControl.Control.Description)
	logControlSnapshot(ctx, client, seed.Control.ID)

	instance, err := waitForWorkflowInstance(ctx, client, seed, 5, time.Second)
	if err != nil {
		return nil, nil, err
	}
	seed.InstanceID = instance.ID
	fmt.Printf("   Workflow instance started: %s (state: %s)\n", instance.ID, instance.State.String())

	fmt.Println("\n6. Verifying webhook workflow completion (no approvals expected)...")
	finalState, err := waitForInstanceState(ctx, client, seed.InstanceID, enums.WorkflowInstanceStateCompleted, 10, time.Second)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("   Workflow instance state: %s\n", finalState.String())

	steps := []string{
		"Initialized organization context",
		"Created webhook+notification workflow (description update trigger)",
		"Created control object",
		"Triggered workflow by updating control description",
		"Confirmed webhook workflow completed (no approvals)",
	}

	return seed, steps, nil
}

// runSlackWebhookDemo demonstrates posting to a Slack incoming webhook when control status hits APPROVED.
func runSlackWebhookDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig, slackURL string) (*WorkflowDemoSeed, []string, error) {
	if slackURL == "" {
		return nil, nil, fmt.Errorf("slack webhook URL required: pass --slack-webhook-url or set SLACK_WEBHOOK_URL/SLACK_WEBHOOK")
	}
	fmt.Println("   Using Slack webhook from flag/env (length:", len(slackURL), ")")

	seed, _, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("\n3. Creating Slack webhook workflow...")

	webhookParams := workflows.WebhookActionParams{
		URL:    slackURL,
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Payload: map[string]any{
			"blocks": []any{
				map[string]any{
					"type": "header",
					"text": map[string]any{
						"type":  "plain_text",
						"text":  "✅ Control Approved",
						"emoji": true,
					},
				},
				map[string]any{
					"type": "section",
					"text": map[string]any{
						"type": "mrkdwn",
						"text": "*Control:* <https://console.theopenlane.io/controls/{{object_id}}|{{ref_code}} - {{title}}>\n*Status:* :white_check_mark: {{status}}",
					},
				},
				map[string]any{
					"type": "context",
					"elements": []any{
						map[string]any{
							"type": "mrkdwn",
							"text": "Triggered by *{{initiator}}* · {{approved_at}}",
						},
					},
				},
				map[string]any{
					"type": "section",
					"fields": []any{
						map[string]any{
							"type": "mrkdwn",
							"text": "*Control*\n{{ref_code}}",
						},
						map[string]any{
							"type": "mrkdwn",
							"text": "*Title*\n{{title}}",
						},
						map[string]any{
							"type": "mrkdwn",
							"text": "*Triggered By*\n{{initiator}}",
						},
						map[string]any{
							"type": "mrkdwn",
							"text": "*Timestamp*\n{{approved_at}}",
						},
					},
				},
				map[string]any{
					"type": "divider",
				},
				map[string]any{
					"type": "actions",
					"elements": []any{
						map[string]any{
							"type": "button",
							"text": map[string]any{
								"type":  "plain_text",
								"text":  "View Control",
								"emoji": true,
							},
							"style": "primary",
							"url":   "https://console.theopenlane.io/controls/{{object_id}}",
						},
					},
				},
			},
		},
		TimeoutMS: 5000,
	}
	webhookBytes, err := marshalParams("webhook params (Slack)", webhookParams)
	if err != nil {
		return nil, nil, err
	}

	defDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeControl,
				Fields:     []string{"status"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "'status' in changed_fields && object.status == \"APPROVED\""},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeWebhook.String(),
				Key:    "slack_webhook",
				Params: webhookBytes,
			},
		},
	}

	defResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Slack Alert on Control Approval",
		Description:    ptr("Posts to Slack when a control is approved"),
		SchemaType:     string(enums.WorkflowObjectTypeControl),
		WorkflowKind:   enums.WorkflowKindNotification,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &defDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Slack webhook workflow definition: %w", err)
	}
	seed.WorkflowDef = &defResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Trigger: status -> APPROVED\n")
	fmt.Printf("   - Action: send Slack webhook\n")

	fmt.Println("\n4. Creating control to trigger workflow...")
	control, err := createControl(ctx, client, seed.OrganizationID,
		"Slack Webhook Demo Control",
		"Control that will be approved to fire Slack webhook",
		"Automation",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control
	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)

	fmt.Println("\n5. Updating control status to APPROVED to trigger workflow...")
	approved := enums.ControlStatusApproved
	updatedControl, err := client.UpdateControl(ctx, seed.Control.ID, graphclient.UpdateControlInput{
		Status: &approved,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to approve control: %w", err)
	}
	fmt.Printf("   Updated control status to: %s\n", *updatedControl.UpdateControl.Control.Status)
	logControlSnapshot(ctx, client, seed.Control.ID)

	instance, err := waitForWorkflowInstance(ctx, client, seed, 10, 2*time.Second)
	if err != nil {
		return nil, nil, err
	}
	seed.InstanceID = instance.ID
	fmt.Printf("   Workflow instance started: %s (state: %s)\n", instance.ID, instance.State.String())

	fmt.Println("\n6. Waiting for completion (webhook-only, no approvals)...")
	finalState, err := waitForInstanceState(ctx, client, seed.InstanceID, enums.WorkflowInstanceStateCompleted, 10, time.Second)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("   Workflow instance state: %s\n", finalState.String())

	steps := []string{
		"Initialized organization context",
		"Created Slack webhook workflow (status APPROVED trigger)",
		"Created control object",
		"Approved control to trigger webhook",
		"Confirmed webhook workflow completed (no approvals)",
	}

	return seed, steps, nil
}

// runFieldUpdateDemo demonstrates a field-update + notification workflow without approvals.
func runFieldUpdateDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, []string, error) {
	seed, userResp, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}
	currentUserID := userResp.Self.ID

	fmt.Println("\n3. Creating field-update workflow...")

	fieldUpdateParams := workflows.FieldUpdateActionParams{
		Updates: map[string]any{
			"category":            "Reviewed - Automated",
			"reference_framework": "SOC 2",
		},
	}
	fieldUpdateBytes, err := marshalParams("field update params", fieldUpdateParams)
	if err != nil {
		return nil, nil, err
	}

	notificationParams := workflows.NotificationActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{
					Type: enums.WorkflowTargetTypeUser,
					ID:   currentUserID,
				},
			},
		},
		Channels: []enums.Channel{enums.ChannelInApp},
		Title:    "Control auto-enriched",
		Body:     "Workflow {{instance_id}} updated control {{object_id}} fields.",
	}
	notificationBytes, err := marshalParams("notification params (auto-enrich)", notificationParams)
	if err != nil {
		return nil, nil, err
	}

	defDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeControl,
				Fields:     []string{"status"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{Expression: "'status' in changed_fields && object.status == \"APPROVED\""},
		},
		Actions: []models.WorkflowAction{
			{
				Type:   enums.WorkflowActionTypeFieldUpdate.String(),
				Key:    "auto_enrich_control",
				Params: fieldUpdateBytes,
			},
			{
				Type:   enums.WorkflowActionTypeNotification.String(),
				Key:    "notify_gov",
				Params: notificationBytes,
			},
		},
	}

	defResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Control Auto-Enrichment Workflow",
		Description:    ptr("Automatically updates control metadata and notifies stakeholders"),
		SchemaType:     string(enums.WorkflowObjectTypeControl),
		WorkflowKind:   enums.WorkflowKindLifecycle,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &defDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create field-update workflow definition: %w", err)
	}
	seed.WorkflowDef = &defResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Trigger: status update to APPROVED\n")
	fmt.Printf("   - Action: update category/reference framework then notify\n")

	fmt.Println("\n4. Creating control to trigger workflow...")
	control, err := createControl(ctx, client, seed.OrganizationID,
		"Field Update Demo Control",
		"Demonstrates field update workflow",
		"Automation",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control
	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)

	fmt.Println("\n5. Updating control status to APPROVED to trigger workflow...")
	approved := enums.ControlStatusApproved
	updatedControl, err := client.UpdateControl(ctx, seed.Control.ID, graphclient.UpdateControlInput{
		Status: &approved,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update control status: %w", err)
	}
	fmt.Printf("   Updated control status: %s -> %s\n", *seed.Control.Status, *updatedControl.UpdateControl.Control.Status)
	logControlSnapshot(ctx, client, seed.Control.ID)

	instance, err := waitForWorkflowInstance(ctx, client, seed, 5, time.Second)
	if err != nil {
		return nil, nil, err
	}
	seed.InstanceID = instance.ID
	fmt.Printf("   Workflow instance started: %s (state: %s)\n", instance.ID, instance.State.String())

	fmt.Println("\n6. Verifying field-update workflow completion (no approvals expected)...")
	finalState, err := waitForInstanceState(ctx, client, seed.InstanceID, enums.WorkflowInstanceStateCompleted, 10, time.Second)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("   Workflow instance state: %s\n", finalState.String())

	steps := []string{
		"Initialized organization context",
		"Created field-update workflow (status APPROVED trigger)",
		"Created control object",
		"Triggered workflow by updating control status to APPROVED",
		"Confirmed field-update workflow completed (no approvals)",
	}

	return seed, steps, nil
}

// runEvidenceReviewDemo demonstrates an evidence review workflow triggered by edge changes.
func runEvidenceReviewDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig, slackURL string) (*WorkflowDemoSeed, []string, error) {
	seed, userResp, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("\n=== Evidence Review Workflow Demo ===")
	fmt.Println("This demo creates an approval workflow triggered when evidence is linked to a control")

	fmt.Println("\n2. Creating workflow definition for evidence review...")
	fmt.Println("   Note: Using current user for the review assignment in this demo")

	currentUserID := userResp.Self.ID

	approvalParams := workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{
				{
					Type: enums.WorkflowTargetTypeUser,
					ID:   currentUserID,
				},
			},
		},
		Required: ptr(true),
		Label:    "Evidence Review",
		Fields:   []string{"workflow_eligible_marker"},
	}

	approvalParamsBytes, err := marshalParams("approval params (evidence review)", approvalParams)
	if err != nil {
		return nil, nil, err
	}

	var workflowActions []models.WorkflowAction

	if slackURL != "" {
		reviewWebhook := workflows.WebhookActionParams{
			URL:    slackURL,
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Payload: map[string]any{
				"text": "Evidence review required for {{object_id}} (instance {{instance_id}}).",
			},
			TimeoutMS: 5000,
		}
		reviewWebhookBytes, err := marshalParams("webhook params (review required)", reviewWebhook)
		if err != nil {
			return nil, nil, err
		}
		workflowActions = append(workflowActions, models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "notify_review_required",
			Params: reviewWebhookBytes,
		})
	} else {
		fmt.Printf("   Skipping Slack webhooks (no webhook URL provided)\n")
	}

	workflowActions = append(workflowActions, models.WorkflowAction{
		Type:   enums.WorkflowActionTypeApproval.String(),
		Key:    "evidence_review",
		Params: approvalParamsBytes,
	})

	if slackURL != "" {
		completionWebhook := workflows.WebhookActionParams{
			URL:    slackURL,
			Method: "POST",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Payload: map[string]any{
				"text": "Evidence review approved for {{object_id}}.",
			},
			TimeoutMS: 5000,
		}
		completionWebhookBytes, err := marshalParams("webhook params (review complete)", completionWebhook)
		if err != nil {
			return nil, nil, err
		}
		workflowActions = append(workflowActions, models.WorkflowAction{
			Type:   enums.WorkflowActionTypeWebhook.String(),
			Key:    "notify_review_complete",
			Params: completionWebhookBytes,
		})
	}

	workflowDefDoc := models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{
			{
				Operation:  "UPDATE",
				ObjectType: enums.WorkflowObjectTypeEvidence,
				Edges:      []string{"controls"},
			},
		},
		Conditions: []models.WorkflowCondition{
			{
				Expression: "'controls' in changed_edges && size(added_ids['controls']) > 0",
			},
		},
		Actions: workflowActions,
	}

	workflowDefResp, err := client.CreateWorkflowDefinition(ctx, graphclient.CreateWorkflowDefinitionInput{
		Name:           "Evidence Review Workflow",
		Description:    ptr("Request approval when evidence is linked to a control"),
		SchemaType:     string(enums.WorkflowObjectTypeEvidence),
		WorkflowKind:   enums.WorkflowKindApproval,
		Active:         ptr(true),
		Draft:          ptr(false),
		OwnerID:        &seed.OrganizationID,
		DefinitionJSON: &workflowDefDoc,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create workflow definition: %w", err)
	}
	seed.WorkflowDef = &workflowDefResp.CreateWorkflowDefinition.WorkflowDefinition
	syncOrgFromDefinition(seed)

	fmt.Printf("   Created workflow: %s\n", seed.WorkflowDef.Name)
	fmt.Printf("   - Schema type: Evidence\n")
	fmt.Printf("   - Triggers: When evidence is linked to controls\n")
	fmt.Printf("   - Review assignment: %s\n", userResp.Self.Email)

	fmt.Println("\n3. Creating control...")
	control, err := createControl(ctx, client, seed.OrganizationID,
		"Evidence Collection and Review",
		"Ensure evidence is reviewed when linked to controls",
		"Audit & Compliance",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create control: %w", err)
	}
	seed.Control = control
	fmt.Printf("   Created control: %s (%s)\n", *seed.Control.Title, seed.Control.RefCode)

	fmt.Println("\n4. Creating evidence in SUBMITTED status...")
	evidenceResp, err := client.CreateEvidence(ctx, graphclient.CreateEvidenceInput{
		Name:        "Q4 Access Control Logs",
		Description: ptr("Access control audit logs for Q4 2024 showing user authentication and authorization events"),
		Status:      ptr(enums.EvidenceStatusSubmitted),
		OwnerID:     &seed.OrganizationID,
	}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create evidence: %w", err)
	}
	evidenceID := evidenceResp.CreateEvidence.Evidence.ID
	fmt.Printf("   Created evidence: %s (ID: %s, status: %s)\n",
		evidenceResp.CreateEvidence.Evidence.Name, evidenceID, *evidenceResp.CreateEvidence.Evidence.Status)

	fmt.Println("\n=== Testing Instructions ===")
	fmt.Println("\n1. Attach evidence to control to trigger workflow:")
	fmt.Printf("   - Evidence ID: %s\n", evidenceID)
	fmt.Printf("   - Control ID:  %s\n", seed.Control.ID)
	fmt.Println("   - Use the UI or API to link the evidence to the control")
	fmt.Printf("\n2. Approve the assignment as %s:\n", userResp.Self.Email)
	fmt.Println("   - Navigate to https://console.theopenlane.io/workflows/assignments")
	fmt.Println("   - Approve the evidence review assignment")
	fmt.Println("\n3. View workflow history:")
	fmt.Println("   - Check the evidence detail page for the workflow panel")
	fmt.Println("   - Review assignments for the approval record")
	fmt.Println("\n=========================")

	fmt.Println("\n=== Entity IDs for API Testing ===")
	fmt.Printf("Organization ID: %s\n", seed.OrganizationID)
	fmt.Printf("Control ID:      %s\n", seed.Control.ID)
	fmt.Printf("Evidence ID:     %s\n", evidenceID)
	fmt.Println("=========================")

	steps := []string{
		"Initialized organization context",
		"Created workflow definition for evidence review",
		"Created control object",
		"Created evidence in SUBMITTED status",
		"Provided testing guide for the evidence review workflow",
	}

	return seed, steps, nil
}

// runExamplesDemo loads workflow definitions from JSON files.
func runExamplesDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig, examplesDir string) (*WorkflowDemoSeed, []string, error) {
	seed, _, client, err := bootstrapDemo(ctx, config, apiClient, demo)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("\n3. Loading workflow definitions from %s...\n", examplesDir)

	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read examples dir %s: %w", examplesDir, err)
	}

	created := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(examplesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read %s: %w", path, err)
		}

		var doc models.WorkflowDefinitionDocument
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}

		input := graphclient.CreateWorkflowDefinitionInput{
			Name:           doc.Name,
			SchemaType:     doc.SchemaType,
			WorkflowKind:   doc.WorkflowKind,
			Active:         ptr(true),
			Draft:          ptr(false),
			OwnerID:        &seed.OrganizationID,
			DefinitionJSON: &doc,
		}
		if doc.Description != "" {
			input.Description = ptr(doc.Description)
		}

		resp, err := client.CreateWorkflowDefinition(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create workflow definition from %s: %w", path, err)
		}

		created++
		seed.WorkflowDef = &resp.CreateWorkflowDefinition.WorkflowDefinition
		syncOrgFromDefinition(seed)
		fmt.Printf("   Created workflow: %s (schema: %s)\n", seed.WorkflowDef.Name, seed.WorkflowDef.SchemaType)
	}

	steps := []string{
		fmt.Sprintf("Created %d workflow definition(s) from %s", created, examplesDir),
	}

	return seed, steps, nil
}

// bootstrapDemo returns a client scoped to either an existing org or a new demo org.
func bootstrapDemo(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, *graphclient.GetSelf, *openlane.Client, error) {
	if demo.OrgID != "" || demo.UseDefaultOrg {
		return bootstrapExistingOrg(ctx, config, apiClient, demo)
	}

	return bootstrapNewOrg(ctx, config, apiClient)
}

// bootstrapExistingOrg seeds demo data using an existing organization
func bootstrapExistingOrg(ctx context.Context, config openlane.Config, apiClient *openlane.Client, demo DemoConfig) (*WorkflowDemoSeed, *graphclient.GetSelf, *openlane.Client, error) {
	seed := &WorkflowDemoSeed{}

	fmt.Println("\n1. Using existing organization...")
	userResp, err := apiClient.GetSelf(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get current user: %w", err)
	}
	fmt.Printf("   Current user: %s (%s)\n", userResp.Self.Email, userResp.Self.ID)

	orgID := demo.OrgID
	if orgID == "" {
		// Find a non-personal org the user belongs to
		nonPersonalOrg, err := findNonPersonalOrg(ctx, apiClient)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to find non-personal org: %w", err)
		}
		if nonPersonalOrg != nil {
			orgID = nonPersonalOrg.ID
			fmt.Printf("   Found non-personal organization: %s (%s)\n", nonPersonalOrg.Name, orgID)
		} else if userResp.Self.Setting.DefaultOrg != nil {
			// Fall back to default org if no non-personal org found
			orgID = userResp.Self.Setting.DefaultOrg.ID
			fmt.Println("   No non-personal org found, using default org")
		}
	}
	if orgID == "" {
		return nil, nil, nil, fmt.Errorf("no organization ID provided and no non-personal or default org available")
	}
	seed.OrganizationID = orgID
	fmt.Printf("   Using organization ID: %s\n", seed.OrganizationID)

	return bootstrapOrgClient(ctx, config, apiClient, seed, userResp)
}

// findNonPersonalOrg queries the user's organizations and returns the first non-personal org
func findNonPersonalOrg(ctx context.Context, apiClient *openlane.Client) (*graphclient.GetOrganizations_Organizations_Edges_Node, error) {
	first := int64(50)
	personalOrgFalse := false
	where := &graphclient.OrganizationWhereInput{
		PersonalOrg: &personalOrgFalse,
	}

	resp, err := apiClient.GetOrganizations(ctx, &first, nil, nil, nil, where, nil)
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.Organizations.Edges {
		if edge.Node != nil {
			return edge.Node, nil
		}
	}

	return nil, nil
}

// bootstrapNewOrg creates a new organization and seeds demo data
func bootstrapNewOrg(ctx context.Context, config openlane.Config, apiClient *openlane.Client) (*WorkflowDemoSeed, *graphclient.GetSelf, *openlane.Client, error) {
	seed := &WorkflowDemoSeed{}

	fmt.Println("\n1. Creating organization...")
	orgName := fmt.Sprintf("workflow-demo-%s", ulids.New().String())
	orgResp, err := apiClient.CreateOrganization(ctx, graphclient.CreateOrganizationInput{
		Name:        orgName,
		DisplayName: ptr(fmt.Sprintf("Workflow Demo Organization %s", time.Now().Format("2006-01-02"))),
		Description: ptr("Organization for demonstrating workflow automation"),
	}, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create organization: %w", err)
	}
	seed.OrganizationID = orgResp.CreateOrganization.Organization.ID
	fmt.Printf("   Created organization: %s (ID: %s)\n", orgName, seed.OrganizationID)

	userResp, err := apiClient.GetSelf(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get current user: %w", err)
	}
	fmt.Printf("   Current user: %s (%s)\n", userResp.Self.Email, userResp.Self.ID)

	return bootstrapOrgClient(ctx, config, apiClient, seed, userResp)
}

// bootstrapOrgClient builds an org-scoped client for demo operations
func bootstrapOrgClient(ctx context.Context, config openlane.Config, apiClient *openlane.Client, seed *WorkflowDemoSeed, userResp *graphclient.GetSelf) (*WorkflowDemoSeed, *graphclient.GetSelf, *openlane.Client, error) {
	tokenName := fmt.Sprintf("workflow-demo-token-%s", ulids.New().String())
	tokenResp, err := apiClient.CreateAPIToken(ctx, graphclient.CreateAPITokenInput{
		Name:    tokenName,
		OwnerID: &seed.OrganizationID,
		Scopes: []string{
			"read",
			"write",
		},
		IsActive: func() *bool {
			b := true
			return &b
		}(),
	}, openlane.WithOrganizationHeader(seed.OrganizationID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create organization API token: %w", err)
	}

	orgToken := tokenResp.CreateAPIToken.APIToken.Token
	if len(orgToken) >= 6 {
		fmt.Printf("   Created org API token for demo (ends with ...%s)\n", orgToken[len(orgToken)-6:])
	} else {
		fmt.Println("   Created org API token for demo")
	}

	orgClient, err := newClient(
		config.BaseURL,
		openlane.WithCredentials(openlane.Authorization{
			BearerToken: orgToken,
		}),
		openlane.WithInterceptors(openlane.WithOrganizationHeader(seed.OrganizationID)),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create org token client: %w", err)
	}

	return seed, userResp, orgClient, nil
}

// waitForWorkflowInstance polls until a workflow instance tied to the new control/definition exists.
func waitForWorkflowInstance(ctx context.Context, apiClient *openlane.Client, seed *WorkflowDemoSeed, attempts int, interval time.Duration) (*graphclient.GetWorkflowInstances_WorkflowInstances_Edges_Node, error) {
	first := int64(5)
	defID := seed.WorkflowDef.ID
	controlID := seed.Control.ID

	where := &graphclient.WorkflowInstanceWhereInput{
		OwnerID:              &seed.OrganizationID,
		WorkflowDefinitionID: &defID,
		HasWorkflowObjectRefsWith: []*graphclient.WorkflowObjectRefWhereInput{
			{ControlID: &controlID},
		},
	}

	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(interval)
		}
		// First, try the strict filter: definition + control ref
		resp, err := apiClient.GetWorkflowInstances(ctx, &first, nil, nil, nil, where, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get workflow instances: %w", err)
		}

		for _, edge := range resp.WorkflowInstances.Edges {
			if edge.Node != nil {
				return edge.Node, nil
			}
		}

		// If none found, look for any instance for this definition and verify its object refs manually.
		defOnlyWhere := &graphclient.WorkflowInstanceWhereInput{
			OwnerID:              &seed.OrganizationID,
			WorkflowDefinitionID: &defID,
		}
		defOnlyResp, err := apiClient.GetWorkflowInstances(ctx, &first, nil, nil, nil, defOnlyWhere, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get workflow instances for definition: %w", err)
		}

		for _, edge := range defOnlyResp.WorkflowInstances.Edges {
			if edge.Node == nil {
				continue
			}

			instID := edge.Node.ID
			refWhere := &graphclient.WorkflowObjectRefWhereInput{
				WorkflowInstanceID: &instID,
				ControlID:          &controlID,
			}
			refResp, err := apiClient.GetWorkflowObjectRefs(ctx, &first, nil, nil, nil, refWhere, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to load object refs for instance %s: %w", instID, err)
			}

			if refResp.WorkflowObjectRefs.TotalCount > 0 {
				return edge.Node, nil
			}

			fmt.Printf("   (debug) Instance %s found for def %s but no control ref yet; retrying...\n", instID, defID)
		}
	}

	debugInstancesForDefinition(ctx, apiClient, seed.OrganizationID, seed.WorkflowDef.ID)
	return nil, fmt.Errorf("workflow instance not found after %d attempts", attempts)
}

// waitForPendingAssignments polls until pending assignments for the new instance are available.
func waitForPendingAssignments(ctx context.Context, apiClient *openlane.Client, seed *WorkflowDemoSeed, attempts int, interval time.Duration) ([]*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges_Node, error) {
	first := int64(50)
	status := enums.WorkflowAssignmentStatusPending
	instanceID := seed.InstanceID

	where := &graphclient.WorkflowAssignmentWhereInput{
		OwnerID: &seed.OrganizationID,
		Status:  &status,
		HasWorkflowInstanceWith: []*graphclient.WorkflowInstanceWhereInput{
			{ID: &instanceID},
		},
	}

	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(interval)
		}
		resp, err := apiClient.GetWorkflowAssignments(ctx, &first, nil, nil, nil, where, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get workflow assignments: %w", err)
		}

		assignments := make([]*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges_Node, 0, len(resp.WorkflowAssignments.Edges))
		for _, edge := range resp.WorkflowAssignments.Edges {
			if edge.Node != nil {
				assignments = append(assignments, edge.Node)
			}
		}

		if len(assignments) > 0 {
			return assignments, nil
		}

	}

	fmt.Println("   (debug) No pending assignments found; dumping workflow instances for context")
	debugInstancesForDefinition(ctx, apiClient, seed.OrganizationID, seed.WorkflowDef.ID)
	return nil, fmt.Errorf("pending assignments not found after %d attempts", attempts)
}

// listPendingAssignments fetches pending assignments for the instance without retries.
func listPendingAssignments(ctx context.Context, apiClient *openlane.Client, seed *WorkflowDemoSeed) ([]*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges_Node, error) {
	first := int64(50)
	status := enums.WorkflowAssignmentStatusPending
	instanceID := seed.InstanceID

	where := &graphclient.WorkflowAssignmentWhereInput{
		OwnerID: &seed.OrganizationID,
		Status:  &status,
		HasWorkflowInstanceWith: []*graphclient.WorkflowInstanceWhereInput{
			{ID: &instanceID},
		},
	}

	resp, err := apiClient.GetWorkflowAssignments(ctx, &first, nil, nil, nil, where, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow assignments: %w", err)
	}

	assignments := make([]*graphclient.GetWorkflowAssignments_WorkflowAssignments_Edges_Node, 0, len(resp.WorkflowAssignments.Edges))
	for _, edge := range resp.WorkflowAssignments.Edges {
		if edge.Node != nil {
			assignments = append(assignments, edge.Node)
		}
	}

	return assignments, nil
}

// waitForInstanceState waits until the workflow instance reaches the desired state.
func waitForInstanceState(ctx context.Context, apiClient *openlane.Client, instanceID string, desired enums.WorkflowInstanceState, attempts int, interval time.Duration) (enums.WorkflowInstanceState, error) {
	lastState := enums.WorkflowInstanceStateRunning

	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(interval)
		}
		resp, err := apiClient.GetWorkflowInstanceByID(ctx, instanceID)
		if err != nil {
			return lastState, fmt.Errorf("failed to load workflow instance %s: %w", instanceID, err)
		}

		lastState = resp.WorkflowInstance.State
		if lastState == desired {
			return lastState, nil
		}
	}

	return lastState, fmt.Errorf("workflow instance %s did not reach state %s after %d attempts (last seen: %s)", instanceID, desired.String(), attempts, lastState.String())
}

// logControlSnapshot fetches and prints the current control status/metadata for debugging.
func logControlSnapshot(ctx context.Context, apiClient *openlane.Client, controlID string) {
	resp, err := apiClient.GetControlByID(ctx, controlID)
	if err != nil {
		fmt.Printf("   (debug) Unable to load control %s: %v\n", controlID, err)
		return
	}

	ctrl := resp.Control
	status := ""
	if ctrl.Status != nil {
		status = string(*ctrl.Status)
	}
	fmt.Printf("   (debug) Control snapshot id=%s status=%s updatedAt=%v\n", ctrl.ID, status, ctrl.UpdatedAt)
}

// debugInstancesForDefinition lists any instances for the given org/definition.
func debugInstancesForDefinition(ctx context.Context, apiClient *openlane.Client, orgID, defID string) {
	first := int64(10)
	where := &graphclient.WorkflowInstanceWhereInput{
		OwnerID:              &orgID,
		WorkflowDefinitionID: &defID,
	}
	resp, err := apiClient.GetWorkflowInstances(ctx, &first, nil, nil, nil, where, nil)
	if err != nil {
		fmt.Printf("   (debug) Unable to list instances for def %s: %v\n", defID, err)
		return
	}

	if resp.WorkflowInstances.TotalCount == 0 {
		fmt.Printf("   (debug) No instances found for def %s\n", defID)
		return
	}

	fmt.Printf("   (debug) Found %d instance(s) for def %s\n", resp.WorkflowInstances.TotalCount, defID)
	for _, edge := range resp.WorkflowInstances.Edges {
		if edge.Node == nil {
			continue
		}
		printDebugInstance(ctx, apiClient, edge.Node.ID)
	}
}

// printDebugInstance dumps lightweight diagnostics for a workflow instance.
func printDebugInstance(ctx context.Context, apiClient *openlane.Client, instanceID string) {
	resp, err := apiClient.GetWorkflowInstanceByID(ctx, instanceID)
	if err != nil {
		fmt.Printf("   (debug) Unable to load instance %s: %v\n", instanceID, err)
		return
	}

	inst := resp.WorkflowInstance
	fmt.Printf("   (debug) Instance %s state=%s def=%s owner=%s\n",
		inst.ID, inst.State.String(), inst.WorkflowDefinitionID, safeString(inst.OwnerID))

	// Fetch object refs to see if the control link is present
	first := int64(10)
	refWhere := &graphclient.WorkflowObjectRefWhereInput{
		WorkflowInstanceID: &instanceID,
	}
	refs, err := apiClient.GetWorkflowObjectRefs(ctx, &first, nil, nil, nil, refWhere, nil)
	if err != nil {
		fmt.Printf("   (debug) Unable to load object refs for %s: %v\n", instanceID, err)
		return
	}
	if refs.WorkflowObjectRefs.TotalCount == 0 {
		fmt.Printf("   (debug) Instance %s has no object refs\n", instanceID)
		return
	}

	for _, edge := range refs.WorkflowObjectRefs.Edges {
		if edge.Node == nil {
			continue
		}
		fmt.Printf("   (debug) ObjectRef %s ctrl=%s policy=%s evidence=%s\n",
			edge.Node.ID,
			safeString(edge.Node.ControlID),
			safeString(edge.Node.InternalPolicyID),
			"", // evidence not currently exposed on this query
		)
	}
}

// safeString returns an empty string when the pointer is nil
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ptr returns a pointer to the provided value
func ptr[T any](v T) *T {
	return &v
}

// syncOrgFromDefinition updates seed.OrganizationID to match the workflow definition's actual owner.
// This ensures subsequent queries use the correct org that the definition was created in.
func syncOrgFromDefinition(seed *WorkflowDemoSeed) {
	if seed.WorkflowDef != nil && seed.WorkflowDef.OwnerID != nil && *seed.WorkflowDef.OwnerID != "" {
		if seed.OrganizationID != *seed.WorkflowDef.OwnerID {
			fmt.Printf("   (sync) Updating org ID from %s to %s (from workflow definition)\n", seed.OrganizationID, *seed.WorkflowDef.OwnerID)
			seed.OrganizationID = *seed.WorkflowDef.OwnerID
		}
	}
}
