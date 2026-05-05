# Integrations CLI

Today this CLI can be used consistently for 2 main things:

- Sending all of our system emails at once using configured fixtures and the code checked in at that time
- "Quickstart" campaigns which creates 3x different campaigns using various configurations

## Example usage

### Send all system emails

You only need the server to be running to execute this; it sends emails with fixture data for things like links, company infos, etc. We can add actual seeding of information and real URL creation in the future, but the use of this tool is primiarly to see the fully rendered template configuration in your inbox so we can confirm visual branding / aesthetics.

```bash
go run main.go email-test send-all --to manderson@theopenlane.io
Dispatcher                    Status   Error
----------                    ------   -----
BrandedMessageRequest         OK
VerifyEmailRequest            OK
WelcomeRequest                OK
InviteRequest                 OK
InviteJoinedRequest           OK
PasswordResetEmailRequest     OK
PasswordResetSuccessRequest   OK
SubscribeRequest              OK
VerifyBillingRequest          OK
TrustCenterNDASignedEmail     OK
QuestionnaireAuthEmail        OK
BillingEmailChangedEmail      OK
TrustCenterNDARequestEmail    OK
TrustCenterAuthEmail          OK
```

### Quickstart Campaigns

Quick start will create the email templates, the associated campaigns and campaign targets, and then execute the dispatch of the campaign giving you the full end to end flow. If you look in cmd/quickstart you can see the associated files used to create the email templates, also demonstrating the variable interpolation + substitution with email templates, and the branding that can be applied to email configuration.

```bash
go run main.go quickstart
{"level":"info","app":"integrations","campaign_id":"01KQTD9SKMHED7N6DY2WBPHX7S","time":"2026-05-04T16:11:26-05:00","message":"branded campaign launched"}
{"level":"info","app":"integrations","campaign_id":"01KQTD9SP9XGJQ0DMKXT7YYYJD","time":"2026-05-04T16:11:26-05:00","message":"cloudflare branded campaign launched"}
{"level":"info","app":"integrations","campaign_id":"01KQTD9SR1ES10HNCD6EE3W9V9","time":"2026-05-04T16:11:26-05:00","message":"questionnaire campaign launched"}
Type                 CampaignID                   Recipient             Queued   Skipped
----                 ----------                   ---------             ------   -------
branded              01KQTD9SKMHED7N6DY2WBPHX7S   mitb@theopenlane.io   1        0
branded-cloudflare   01KQTD9SP9XGJQ0DMKXT7YYYJD   mitb@theopenlane.io   1        0
questionnaire        01KQTD9SR1ES10HNCD6EE3W9V9   mitb@theopenlane.io   1        0
```
