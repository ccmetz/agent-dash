# Vue, Go, and SQLite for Local Dashboard

The dashboard will use a Vite + Vue 3 frontend, a Go backend server, and SQLite for the dashboard-owned Analytics Store. The Go backend owns Usage Sync, OpenCode database reads, Analytics Store writes, polling, and API endpoints; the Vue frontend only renders dashboard data and triggers user actions such as manual refresh.

We chose this split over a single full-stack JavaScript app or desktop shell because the product needs local filesystem/database access, durable local analytics, and testable ingestion logic, while still benefiting from a lightweight browser-based dashboard UI. This creates two development surfaces, but keeps source ingestion and storage concerns out of the frontend.
