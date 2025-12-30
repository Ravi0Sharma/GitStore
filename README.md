## GitStore

GitStore is a Git-inspired project built to explore repository modeling and backend architecture.
It consists of a Go backend with a custom append-only key–value store, a React (Vite) frontend,
and a Node.js CLI.

### High-level Overview

GitStore lets you:

- **Create repositories** from the web UI and persist them via the Go server
- **Stage files, commit, create branches, and merge** using Git-like flows
- **Push** to update “remote refs” so commits become visible in the UI
- **Track simple issues** per repository
- Use a **Node CLI** for local file operations and standard Git operations

### UI & User Experience

The web client is intentionally designed to feel “product-grade” (layout, navigation, empty-states, and clear flows), even though the backend is learning-focused.

#### Landing Page

Clean landing experience with onboarding-oriented sections and a consistent design system.

![Landing Page](docs/images/landingPage-1.png)
![Landing Page](docs/images/landingPage-2.png)

### Authentication (Firebase)

The client uses **Firebase Authentication** 

- **Email & password** 
- **Google sign-in** 

### Dashboard

The dashboard is the primary workspace: it lists repositories and provides navigation into repo features (branches, commits/merge, issues, CLI-like interactions).

![Dashboard](docs/images/dashboard.png)

### Repository Features

#### Create Repository

Create repositories directly from the UI; repositories persist on the server via a metadata registry.

#### Commits & Pu

GitStore models a “local vs pushed” distinction:

- **Commits are created locally** branch refs move.
- **Commits become visible in the UI after push**, because commit listing reads from `refs/remotes/origin/<branch>` (the “pushed view”).

![RepoPage](docs/images/repoView.png)

#### Issues

Basic issue tracking per repository (creation, listing and status updates).

![Issues](docs/images/issues.png)

### CLI Tool

The Node-based CLI (`cli/`) is built to support:

- **File operations** (create/write/append) in local repositories.
- **Standard Git operations** for regular `.git`repositories 


#### CLI Commands

![CLI help](docs/images/cli-help.png)

![CLI push](docs/images/cli-push.png)

### Storage Engine

The backend uses a custom append-only key–value storage engine written in Go.

- Used for repository metadata and per-repo state
- Built to explore durability, crash recovery, and storage design tradeoffs

### API

REST API (`/api/repos/*`) for repository operations: create, branches, commits, merge, files and issues.

### Disclaimer

This repository is built for **learning and exploration**.

### Known Limitations

- **No backend authentication/authorization** (Firebase auth is frontend-only)
- **Concurrency risks** in backend flows (documented in the production review)
- **No CI/CD ** currently in the repo
- **Storage engine lacks compaction** (append-only log grows over time)

