# Metadata-Only Usage Sync

The dashboard will sync Usage Metadata only: session titles, project names and local paths, timestamps, Coding Agent/provider/model identifiers, token counts, Actual Cost, statuses, Source Identity, Source Timestamps, Billing Profile labels, and Tool Call counts by name. It will not sync prompts, assistant responses, diffs, file contents, tool inputs, tool outputs, OpenCode auth data, or account records in the MVP.

We chose this boundary because the dashboard's purpose is usage analytics, not transcript replay, and OpenCode's local data contains sensitive code and auth-adjacent information. Session titles and project paths are allowed because they make the dashboard usable for local analysis; account attribution remains a future enhancement only if a reliable per-Model Call source identity can be established.
