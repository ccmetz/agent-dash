# Agent Usage Dashboard

A local dashboard for understanding how AI coding agents consume models, tokens, cost, and tools across usage over time.

## Language

**Coding Agent**:
A local AI coding application used to perform Agent Sessions.
_Avoid_: Tool, client, app

**Agent Session**:
A continuous period of work performed through one Coding Agent.
_Avoid_: Run, chat, conversation

**Model Call**:
A single request from an agent tool to a model during an Agent Session.
_Avoid_: Completion, API call, inference

**Usage Source**:
A local agent tool data store that provides historical usage records for import.
_Avoid_: Connector, integration, provider

**Usage Sync**:
An ongoing read from a Usage Source that inserts new Usage Metadata and updates changed Usage Metadata in the Analytics Store.
_Avoid_: One-time import, scrape, refresh

**Tool Call**:
An action performed by a Coding Agent inside an Agent Session, such as editing a file or running a command.
_Avoid_: Agent tool, coding tool, app

**Usage Metadata**:
Non-content usage facts such as timestamps, models, token counts, costs, tool names, and session/project labels.
_Avoid_: Transcript, raw content, prompt data

**Usage Overview**:
A dashboard view that summarizes usage across Agent Sessions and Model Calls.
_Avoid_: Home page, report, summary screen

**Analytics Store**:
The dashboard-owned local store of normalized Usage Metadata.
_Avoid_: Source database, cache, warehouse

**Source Identity**:
The original Usage Source and source record IDs used to recognize the same Agent Session, Model Call, or Tool Call across syncs.
_Avoid_: Fingerprint, local ID, external key

**Source Timestamp**:
The original Usage Source timestamp used to detect when synced Usage Metadata may have changed.
_Avoid_: Dashboard timestamp, sync time, display time

**In-Progress Session**:
An Agent Session inferred to still be active because it was updated recently.
_Avoid_: Live session, open session, running session

**Project**:
A local codebase where Agent Sessions occur.
_Avoid_: Repository, workspace, folder

**Actual Cost**:
The cost value reported by a Usage Source for a Model Call.
_Avoid_: Estimated cost, market cost, inferred spend

**Billing Caveat**:
A notice that Actual Cost may be zero or incomplete because usage came through subscription-based or free-usage billing.
_Avoid_: Cost estimate, pricing warning, billing error

**Billing Profile**:
The known billing behavior for a provider or model, such as actual-cost reported, subscription-based, or free-tier.
_Avoid_: Caveated model, pricing mode, cost type

## Relationships

- A **Coding Agent** performs **Agent Sessions**
- An **Agent Session** contains one or more **Model Calls**
- A **Model Call** belongs to exactly one **Agent Session**
- **Model Calls** are grouped by Coding Agent, provider, and model
- A **Usage Source** provides **Agent Sessions** and **Model Calls**
- A **Usage Sync** reads from a **Usage Source** into the **Analytics Store**
- A **Usage Sync** runs when the dashboard opens, when manually requested, and periodically while the dashboard is open
- An **Agent Session** can contain **Tool Calls**
- **Tool Calls** are tracked by name and count, not by input or output content
- **Usage Metadata** describes **Agent Sessions** and **Model Calls** without storing their raw content
- A **Usage Overview** aggregates **Usage Metadata** across many **Agent Sessions**
- The **Analytics Store** stores normalized **Usage Metadata** imported from Usage Sources
- **Source Identity** lets **Usage Sync** upsert existing records instead of duplicating them
- **Source Timestamps** let **Usage Sync** detect changed source records without storing raw content
- An **In-Progress Session** is updated by later **Usage Syncs** as more Usage Metadata appears
- A **Project** can have many **Agent Sessions**
- A **Project** is stored with a display name and local path
- **Actual Cost** belongs to a **Model Call** when the **Usage Source** reports it
- A **Billing Caveat** applies when a **Model Call** uses a **Billing Profile** whose Actual Cost may be incomplete
- A provider-level **Billing Profile** can be overridden by a model-specific **Billing Profile**

## Example dialogue

> **Dev:** "Should the dashboard show usage by complete opencode work session or by every model request?"
> **Domain expert:** "Both: an **Agent Session** is the top-level thing I review, and **Model Calls** explain the token and cost details inside it."

## Flagged ambiguities

- "usage" could mean session-level activity or model-level consumption; resolved: **Agent Session** is the top-level unit, and **Model Call** is the detailed unit.
- "tool" could mean the coding application or an internal action; resolved: **Coding Agent** means OpenCode/Claude Code-style apps, and **Tool Call** means internal actions.
- "import" was ambiguous between a one-time snapshot and ongoing updates; resolved: the dashboard performs **Usage Sync** from OpenCode's local database.
- "metadata only" means storing **Usage Metadata** for analytics while excluding prompts, diffs, tool payloads, and conversation transcripts.
- **Tool Call** analytics include counts by tool name only, not the command, file path, input, or output.
- "cost" means **Actual Cost**, not an estimated market-rate cost.
- A zero-cost **Model Call** is not automatically caveated; **Billing Caveats** depend on manually assigned **Billing Profiles**.
- Account attribution is a desired future enhancement, but skipped for MVP because OpenCode usage records do not currently expose a reliable account identity per Model Call.
