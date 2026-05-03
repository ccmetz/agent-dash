# Periodic Usage Sync Into Analytics Store

The dashboard will maintain its own local Analytics Store and keep it updated by periodically syncing Usage Metadata from OpenCode's local database when the dashboard opens, when refresh is requested, and every 60 seconds while open. We chose this over a one-time import because the dashboard should keep reflecting new Coding Agent usage, and over a plugin/live event pipeline because OpenCode already persists the needed metadata and DB sync avoids extra setup and tighter runtime coupling.

Considered alternatives: one-time import, filesystem watching, and an OpenCode plugin posting events to a local server. Periodic DB sync requires idempotent upserts by Source Identity, but it gives enough freshness for a local usage dashboard while preserving a path to add plugin-based ingestion later if source metadata proves insufficient.
