## GitStore

GitStore is a Git-inspired system built as a learning and architecture project: a Go backend and “gitclone” engine backed by a custom append-only KV store, a polished React (Vite) UI and a Node.js CLI. The goal is to explore repository modeling (refs, commits, merges), durability trade-offs 

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

![Landing Page](docs/images/landing-1.png)
![Landing Page](docs/images/landing-2.png)

### Authentication (Firebase)

The client uses **Firebase Authentication** for a smooth UX and identity display.

- **Email & password** 
- **Google sign-in** 

- Auth is **frontend-only**: the backend does not validate Firebase ID tokens.

### Dashboard

The dashboard is the primary workspace: it lists repositories and provides navigation into repo features (branches, commits/merge, issues, CLI-like interactions).

![Dashboard](docs/images/dashboard.png)

### Repository Features

#### Create Repository

Create repositories directly from the UI; repositories persist on the server via a metadata registry.

#### Commits & Merge

GitStore models a “local vs pushed” distinction:

- **Commits are created locally** branch refs move.
- **Commits become visible in the UI after push**, because commit listing reads from `refs/remotes/origin/<branch>` (the “pushed view”).
- Merge flow supports fast-forward-style constraints; the UI typically pushes after merge so history becomes visible.

![Commits / Merge](docs/images/commits-merge.png)

#### Issues

Basic issue tracking per repository (creation, listing, and status updates).

![Issues](docs/images/issues.png)

### CLI Tool

The Node-based CLI (`cli/`) is built to support:

- **File operations** in local repositories (create/write/append).
- **Standard Git operations** for regular `.git` repositories using `simple-git`.
- A clear boundary for GitStore repos:
  - GitStore repos require the backend API; “git-style” operations in the CLI are currently not fully implemented for `.gitclone` repos.

#### CLI Commands

Help output (placeholder screenshot):

![CLI help](docs/images/cli-help.png)

Example push command output (placeholder screenshot):

![CLI push](docs/images/cli-push.png)

### Architecture & Engineering Focus

This project is primarily an engineering sandbox to practice system design thinking:

- **Go backend**: HTTP API + service layer (`gitClone/internal/transport/http`, `gitClone/internal/app`)
- **Custom append-only KV store**: `gitDb/` with durability tests
- **Repo storage**: per-repo store under `.gitclone/` and a global metadata store for repo registry
- Clear separation of concerns (transport / app / storage)

### Disclaimer

This repository is built for **learning and exploration**.

### Known Limitations

- **No backend authentication/authorization** (Firebase auth is frontend-only)
- **Concurrency risks** in backend flows (documented in the production review)
- **No CI/CD ** currently in the repo
- **Storage engine lacks compaction** (append-only log grows over time)

