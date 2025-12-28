# GitStore

A monorepo containing:
- **gitClone**: Go CLI for git-like operations with log-structured storage
- **gitDb**: Go KV-store (log-structured store)
- **Client**: React web application (Vite + React + TypeScript + Tailwind v4)

## Quick Start

### Prerequisites
- Go 1.24.4 or later
- Node.js and npm

### Setup from Clean Clone

**Step 1: Clone the repository**
```bash
git clone <repository-url>
cd GitStore
```

**Step 2: Build and run the Go server (REQUIRED for full functionality)**

⚠️ **IMPORTANT**: The backend server must be running for repositories to persist. Without it, repositories are only stored in browser memory.

```bash
# Build the server
cd gitClone
go build -o gitserver ./cmd/gitserver

# Run the server (defaults to port 8080)
# Repositories are stored in ./data/repos and metadata in ./data/db
./gitserver
```

The server will start on `http://localhost:8080` by default and log:
```
Repository base directory (absolute): /path/to/data/repos
Metadata database path (absolute): /path/to/data/db
Starting GitStore server on port 8080, repo base: /path/to/data/repos
```

You can configure:
- `PORT`: Server port (default: 8080)
- `GITSTORE_REPO_BASE`: Base directory for repositories (default: `./data/repos`)
- `GITSTORE_DB_PATH`: Path to metadata database (default: `./data/db`)

Example with custom paths:
```bash
GITSTORE_DB_PATH=./data/db GITSTORE_REPO_BASE=./data/repos PORT=8080 ./gitserver
```


**Step 3: In a new terminal, start the React client**
```bash
cd Client
npm install
npm run dev
```

**Step 4: Configure Environment Variables**

Create a `.env` file in the `Client` directory:

```bash
cp Client/.env.example Client/.env
```

Then edit `Client/.env`:

**API Configuration:**
- Leave `VITE_API_URL` empty to use Vite proxy (recommended for dev)
- Or set `VITE_API_URL=http://localhost:8080` if backend runs on different port
- Vite proxy automatically forwards `/api/*` to `http://localhost:8080`

**Firebase Configuration (required for authentication):**
```env
VITE_FIREBASE_APIKEY=your-api-key
VITE_FIREBASE_AUTH_DOMAIN=your-auth-domain
VITE_FIREBASE_PROJECT_ID=your-project-id
VITE_FIREBASE_STORAGE_BUCKET=your-storage-bucket
VITE_FIREBASE_MESSAGE_SENDER_ID=your-sender-id
VITE_FIREBASE_APP_ID=your-app-id
VITE_FIREBASE_MEASUREMENT_ID=your-measurement-id
```

**Step 5: Access the application**
- Open `http://localhost:5173` (or the port shown by Vite)
- Sign up or sign in with Firebase authentication

### Complete Setup Commands (from clean clone)

```bash
# 1. Clone
git clone <repository-url>
cd GitStore

# 2. Build and run server (in terminal 1)
cd gitClone
go build -o gitserver ./cmd/gitserver
GITSTORE_DB_PATH=./data/db GITSTORE_REPO_BASE=./data/repos ./gitserver

# 3. Setup and run client (in terminal 2)
cd Client
npm install
cp .env.example .env
# Edit .env with your Firebase credentials
npm run dev
```

## Project Structure

```
GitStore/
├── gitClone/          # Go CLI and core logic
│   ├── cmd/
│   │   ├── gitclone/  # CLI command
│   │   └── gitserver/ # HTTP API server
│   └── internal/      # Internal packages
├── gitDb/             # Log-structured KV store
└── Client/             # React frontend
    └── src/
        ├── pages/      # Page components
        ├── components/ # Reusable components
        ├── context/    # React context (GitContext)
        ├── lib/        # API client
        └── routes.ts   # Route definitions
```

## API Endpoints

The Go server exposes the following REST API:

- `GET /api/repos` - List all repositories
- `POST /api/repos` - Create a new repository
- `GET /api/repos/:id` - Get repository details
- `GET /api/repos/:id/branches` - Get repository branches
- `GET /api/repos/:id/commits` - Get repository commits
- `POST /api/repos/:id/checkout` - Checkout a branch
- `POST /api/repos/:id/commit` - Create a commit
- `POST /api/repos/:id/merge` - Merge branches

### Testing API Endpoints

You can test the API endpoints manually using `curl`:

```bash
# 1. List all repositories (should return [] if empty)
curl http://localhost:8080/api/repos

# 2. Create a new repository
curl -X POST http://localhost:8080/api/repos \
  -H "Content-Type: application/json" \
  -d '{"name":"test-repo","description":"A test repository"}'

# 3. List repositories again (should now include the new repo)
curl http://localhost:8080/api/repos

# 4. Get repository details
curl http://localhost:8080/api/repos/test-repo

# 5. Get repository branches
curl http://localhost:8080/api/repos/test-repo/branches

# 6. Create a new branch (checkout creates branch if it doesn't exist)
curl -X POST http://localhost:8080/api/repos/test-repo/checkout \
  -H "Content-Type: application/json" \
  -d '{"branch":"feature/new-feature"}'

# 7. List branches again (should include new branch)
curl http://localhost:8080/api/repos/test-repo/branches

# 8. Merge branches
curl -X POST http://localhost:8080/api/repos/test-repo/merge \
  -H "Content-Type: application/json" \
  -d '{"branch":"feature/new-feature"}'
```

### Automated API Testing

You can run the automated test script to verify all endpoints:

```bash
# From Client directory
node scripts/test-api.mjs

# Or with custom base URL
node scripts/test-api.mjs http://localhost:8080
```

The script tests:
- GET /api/repos
- POST /api/repos
- GET /api/repos/:id/branches
- POST /api/repos/:id/checkout (creates branch)
- GET /api/repos/:id/commits
- POST /api/repos/:id/merge

**Expected behavior:**
- `GET /api/repos` should always return a JSON array: `[]` (empty) or `[{...}, {...}]` (never `null`)
- `POST /api/repos` should return the created repository with status 201
- After creating a repo, `GET /api/repos` should include it in the list

## Development

### Running Both Server and Client

You can run both in separate terminals, or use a process manager like `concurrently`:

```bash
# Terminal 1
go run ./gitClone/cmd/gitserver

# Terminal 2
cd Client && npm run dev
```

### Building for Production

**Server:**
```bash
go build -o gitserver ./gitClone/cmd/gitserver
```

**Client:**
```bash
cd Client
npm run build
```

## Architecture

### Repository Storage
- **Repository folders**: Stored in `GITSTORE_REPO_BASE` (default: `./data/repos`)
  - Each repo is a directory containing a `.gitclone/` subdirectory
  - Git data (commits, branches, objects) is stored in each repo's `.gitclone/` directory
- **Metadata database**: Stored in `GITSTORE_DB_PATH` (default: `./data/db`)
  - Uses gitDb (log-structured KV store) for repository metadata
  - Stores repo list, names, descriptions, branch/commit counts, timestamps
  - Repositories persist even if repo folders are deleted (marked as `missing: true`)

### API Behavior
- `GET /api/repos`: Returns repository list from metadata database (not file system scan)
- `POST /api/repos`: Creates repo folder AND saves metadata to database
- Branch/commit operations update metadata automatically
- Repositories are stable: they appear in UI even after server restart

## Notes

- Repositories are stored in directories containing a `.gitclone/` subdirectory
- Metadata is stored in gitDb at `GITSTORE_DB_PATH` for fast listing and persistence
- If a repo folder is deleted, metadata is preserved and repo is marked as `missing: true`
- Firebase is used for authentication only
- Issues are not yet implemented in the backend
