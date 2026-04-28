# SixPatrol Server Steering

This document orients LLMs and contributors to the sixpatrol-server project.

## Project Tree (Directories Only)

```
.
├── cmd
│   └── server
├── db
├── env
├── handlers
├── middleware
├── migrations
├── queue
├── realtime
├── static
├── templates
└── tests
```

## Directory Guide

- cmd/server: Application entrypoint and server wiring.
- db: Database clients, models, and tenant configuration templates.
- env: Environment variable loading helpers.
- handlers: HTTP handlers for API endpoints.
- middleware: HTTP middleware (auth, request checks).
- migrations: SQL schema migrations.
- queue: In-memory queue and ingestion helpers.
- realtime: Server-sent events and broadcaster logic.
- static: Static assets.
- templates: HTML templates for dashboard views.
- tests: Unit and integration tests.
