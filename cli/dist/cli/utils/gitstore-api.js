/**
 * GitStore API client for CLI
 * Reuses the API structure from Client/src/lib/api.ts
 */
const DEFAULT_API_URL = 'http://localhost:8080';
/**
 * Get API base URL from environment or use default
 */
function getApiBaseUrl() {
    return process.env.GITSTORE_API_URL || DEFAULT_API_URL;
}
/**
 * Fetch JSON from API
 */
async function fetchJSON(endpoint, options) {
    const apiUrl = getApiBaseUrl();
    const url = `${apiUrl}${endpoint}`;
    const response = await fetch(url, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers,
        },
    });
    if (!response.ok) {
        let errorMessage = response.statusText;
        try {
            const error = await response.json();
            errorMessage = error.error || errorMessage;
        }
        catch {
            // Not JSON, use status text
        }
        throw new Error(errorMessage || `HTTP ${response.status}`);
    }
    return response.json();
}
/**
 * Check if GitStore API server is available
 */
export async function isApiAvailable() {
    try {
        const apiUrl = getApiBaseUrl();
        const response = await fetch(`${apiUrl}/api/repos`, {
            method: 'GET',
            signal: AbortSignal.timeout(2000), // 2 second timeout
        });
        return response.ok || response.status === 404; // 404 is ok, means server is running
    }
    catch {
        return false;
    }
}
/**
 * Find repository ID by path
 * This searches through all repos to find one matching the path
 */
export async function findRepoByPath(repoPath) {
    try {
        const repos = await listRepos();
        // Try to match by name or path
        // For now, we'll need to check if the repo name matches the directory name
        const repoName = repoPath.split('/').pop() || repoPath;
        const repo = repos.find(r => r.id === repoName || r.name === repoName);
        return repo ? repo.id : null;
    }
    catch {
        return null;
    }
}
export const gitstoreApi = {
    async listRepos() {
        return fetchJSON('/api/repos');
    },
    async createRepo(name, description) {
        return fetchJSON('/api/repos', {
            method: 'POST',
            body: JSON.stringify({ name, description }),
        });
    },
    async getBranches(repoId) {
        return fetchJSON(`/api/repos/${repoId}/branches`);
    },
    async getCommits(repoId) {
        return fetchJSON(`/api/repos/${repoId}/commits`);
    },
    async checkout(repoId, branch) {
        await fetchJSON(`/api/repos/${repoId}/checkout`, {
            method: 'POST',
            body: JSON.stringify({ branch }),
        });
    },
    async commit(repoId, message) {
        await fetchJSON(`/api/repos/${repoId}/commit`, {
            method: 'POST',
            body: JSON.stringify({ message }),
        });
    },
    async merge(repoId, branch) {
        return fetchJSON(`/api/repos/${repoId}/merge`, {
            method: 'POST',
            body: JSON.stringify({ branch }),
        });
    },
};
// Export for convenience
export const { listRepos, createRepo, getBranches, getCommits, checkout, commit, merge } = gitstoreApi;
//# sourceMappingURL=gitstore-api.js.map