# Directory Sync: Remaining Work

Directory sync pulls each customer's user directory (Azure Entra ID, Google
Workspace, etc.) on a schedule so Openlane can show who works at the company,
which systems they appear in, and eventually drive onboarding and offboarding
automation with audit evidence.

## What works today

- New people, groups, and group memberships are detected and stored.
- Changes to a person's profile are detected; unchanged records are recognized
  and skipped cheaply.
- When someone is removed from a group, that removal is detected and recorded
  with a timestamp.
- A new person appearing in the directory creates a person record and emits
  events that workflows can react to, so an onboarding-style trigger exists.

## What is missing for the customer-facing goal

- **A person leaving the company is not detected.** Group membership removals
  are inferred, but a user account disappearing from the directory is not
  recorded anywhere. Nothing can trigger offboarding, and there is no removal
  evidence to show an auditor.
- **No timeline.** The UI shows a person as a flat record with the systems they
  are in today. When they were added to or removed from each system is not
  presented, even though some of the underlying dates exist.
- **Provider noise.** Google Workspace reports login activity inside the user
  profile, so people who logged in recently look "changed" every sync and
  generate unnecessary update work.

## Questions to answer before building

- What counts as "removed"? A provider can delete an account, disable it, or
  report a future leave date — which of these should trigger offboarding, and
  which should only be recorded?
- What should a re-hire look like — does history restart, or continue?
- What exactly must an auditor be able to see or export per person: which
  systems, which dates, observed by which sync?
- How quickly must additions and removals be detected? This sets the sync
  frequency floor and its infrastructure cost.
