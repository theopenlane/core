package notifications

import (
    "encoding/json"
    "fmt"
    "reflect"

    "github.com/theopenlane/core/internal/ent/generated"
    "github.com/theopenlane/core/internal/ent/generated/privacy"
    "github.com/theopenlane/core/pkg/enums"
    "github.com/theopenlane/core/pkg/events/soiree"
    "github.com/theopenlane/core/pkg/logx"
    "github.com/theopenlane/iam/auth"
)

type notificationData struct {
    title            string
    body             string
    channels         []string
    userIDs          []string
    data             string
    ownerID          string
    notificationType string
}

// HandleTaskMutation processes task mutations and creates notifications when assignee changes
func HandleTaskMutation(ctx *soiree.EventContext) error {
    event := ctx.Event()
    if event == nil {
        return nil
    }

    props := ctx.Properties()
    if props == nil {
        return nil
    }

    // Get assignee_id from properties
    assigneeID := props.GetKey("assignee_id")
    if assigneeID == nil {
        return nil
    }

    newAssigneeID := fmt.Sprintf("%v", assigneeID)
    if !taskAssigneeChanged(ctx, newAssigneeID) {
        return nil
    }

    title := props.GetKey("title")
    entityID := props.GetKey("ID")
    ownerID := props.GetKey("owner_id")

    // Get assigner info (the user who made the change) from context
    assignerID := "system"
    if au, err := auth.GetAuthenticatedUserFromContext(ctx.Context()); err == nil && au != nil {
        assignerID = au.SubjectID
    }

    if err := addTaskAssigneeNotification(ctx, newAssigneeID, fmt.Sprintf("%v", title), fmt.Sprintf("%v", entityID), fmt.Sprintf("%v", ownerID), assignerID); err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add task assignee notification")
        return err
    }

    return nil
}

func taskAssigneeChanged(ctx *soiree.EventContext, newAssigneeID string) bool {
    oldID, ok := previousTaskAssigneeID(ctx)
    if !ok {
        return true
    }

    return oldID != newAssigneeID
}

func previousTaskAssigneeID(ctx *soiree.EventContext) (string, bool) {
    taskMutation := taskMutationFromEvent(ctx)
    if taskMutation == nil {
        return "", false
    }

    oldID, err := taskMutation.OldAssigneeID(ctx.Context())
    if err != nil {
        return "", false
    }

    return oldID, true
}

func taskMutationFromEvent(ctx *soiree.EventContext) *generated.TaskMutation {
    if ctx == nil {
        return nil
    }

    event := ctx.Event()
    if event == nil {
        return nil
    }

    payload := event.Payload()
    if payload == nil {
        return nil
    }

    value := reflect.ValueOf(payload)
    for value.Kind() == reflect.Pointer {
        if value.IsNil() {
            return nil
        }
        value = value.Elem()
    }

    if value.Kind() != reflect.Struct {
        return nil
    }

    field := value.FieldByName("Mutation")
    if !field.IsValid() || !field.CanInterface() {
        return nil
    }

    iface := field.Interface()
    if iface == nil {
        return nil
    }

    taskMutation, _ := iface.(*generated.TaskMutation)
    return taskMutation
}

// HandleInternalPolicyMutation processes internal policy mutations and creates notifications when status = NEEDS_APPROVAL
func HandleInternalPolicyMutation(ctx *soiree.EventContext) error {
    event := ctx.Event()
    if event == nil {
        return nil
    }

    props := ctx.Properties()
    if props == nil {
        return nil
    }

    // Get status from properties
    status := props.GetKey("status")
    if status == nil {
        return nil
    }

    // Check if status is NEEDS_APPROVAL
    if fmt.Sprintf("%v", status) != "NEEDS_APPROVAL" {
        return nil
    }

    // TODO: Check if status was changed (compare with old value)

    // Get approver_id from properties
    approverID := props.GetKey("approver_id")
    if approverID == nil {
        logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for internal policy with NEEDS_APPROVAL status")
        return nil
    }

    name := props.GetKey("name")
    entityID := props.GetKey("ID")
    ownerID := props.GetKey("owner_id")

    if err := addInternalPolicyNotification(ctx, fmt.Sprintf("%v", approverID), fmt.Sprintf("%v", name), fmt.Sprintf("%v", entityID), fmt.Sprintf("%v", ownerID)); err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add internal policy notification")
        return err
    }

    return nil
}

func addTaskAssigneeNotification(ctx *soiree.EventContext, assigneeID, taskTitle, taskID, ownerID, assignerID string) error {
    // create the data map with the URL
    dataMap := map[string]string{
        "url": fmt.Sprintf("https://console.theopenlane.io/tasks?id=%s", taskID),
    }

    dataJSON, err := json.Marshal(dataMap)
    if err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to marshal notification data")
        return err
    }

    data := notificationData{
        notificationType: "USER",
        userIDs:          []string{assigneeID},
        title:            "New task assigned",
        body:             fmt.Sprintf("%s was assigned by %s", taskTitle, assignerID),
        data:             string(dataJSON),
        ownerID:          ownerID,
    }

    return newNotificationCreation(ctx, data, "Task")
}

func addInternalPolicyNotification(ctx *soiree.EventContext, approverID, policyName, policyID, ownerID string) error {
    client, ok := soiree.ClientAs[*generated.Client](ctx)
    if !ok {
        return nil
    }

    // set allow context to query the group
    allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

    // query the approver group to get all users
    group, err := client.Group.Get(allowCtx, approverID)
    if err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", approverID).Msg("failed to get approver group")
        return err
    }

    // get all users in the approver group
    users, err := group.QueryUsers().All(allowCtx)
    if err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", approverID).Msg("failed to query users in approver group")
        return err
    }

    if len(users) == 0 {
        logx.FromContext(ctx.Context()).Warn().Str("group_id", approverID).Msg("no users found in approver group")
        return nil
    }

    // collect user IDs
    userIDs := make([]string, len(users))
    for i, user := range users {
        userIDs[i] = user.ID
    }

    // create the data map with the URL
    dataMap := map[string]string{
        "url": fmt.Sprintf("https://console.theopenlane.io/policies/%s/view", policyID),
    }

    dataJSON, err := json.Marshal(dataMap)
    if err != nil {
        logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to marshal notification data")
        return err
    }

    data := notificationData{
        notificationType: "ORGANIZATION",
        userIDs:          userIDs,
        title:            "Policy approval required",
        body:             fmt.Sprintf("%s needs approval, internalPolicy", policyName),
        data:             string(dataJSON),
        ownerID:          ownerID,
    }

    return newNotificationCreation(ctx, data, "InternalPolicy")
}

func newNotificationCreation(ctx *soiree.EventContext, data notificationData, objectType string) error {
    client, ok := soiree.ClientAs[*generated.Client](ctx)
    if !ok {
        return nil
    }

    // set allow context
    allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

    // create notification per user that it should be sent to
    for _, userID := range data.userIDs {
        mut := client.Notification.Create()

        // Set owner ID
        if data.ownerID != "" {
            mut.SetOwnerID(data.ownerID)
        }

        mut.SetBody(data.body)
        mut.SetTitle(data.title)
        mut.SetObjectType(objectType)

        // set notification type
        if data.notificationType != "" {
            notifType := enums.ToNotificationType(data.notificationType)
            if notifType != nil && *notifType != enums.NotificationTypeInvalid {
                mut.SetNotificationType(*notifType)
            }
        }

        // set data if provided
        if data.data != "" {
            var dataMap map[string]interface{}
            if err := json.Unmarshal([]byte(data.data), &dataMap); err != nil {
                logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to unmarshal notification data")
            } else {
                mut.SetData(dataMap)
            }
        }

        // set channels if provided
        if len(data.channels) > 0 {
            // convert string channels to enums.Channel and set them
            channels := make([]enums.Channel, 0, len(data.channels))
            for _, ch := range data.channels {
                if c := enums.ToChannel(ch); c != nil && *c != enums.ChannelInvalid {
                    channels = append(channels, *c)
                }
            }
            if len(channels) > 0 {
                mut.SetChannels(channels)
            }
        }

        mut.SetUserID(userID)

        if _, err := mut.Save(allowCtx); err != nil {
            logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to create notification")
            return err
        }
    }

    return nil
}

// RegisterListeners registers notification listeners with the given eventer
// This is called from hooks package to register the listeners
func RegisterListeners(addListener func(entityType string, handler func(*soiree.EventContext) error)) {
    addListener(generated.TypeTask, HandleTaskMutation)
    addListener(generated.TypeInternalPolicy, HandleInternalPolicyMutation)
}