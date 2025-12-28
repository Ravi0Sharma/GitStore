/**
 * GitStore API client for CLI
 * Reuses the API structure from Client/src/lib/api.ts
 */

const DEFAULT_API_URL = 'http://localhost:8080';

export interface RepoListItem {
  id: string;
  name: string;
  description?: string;
  currentBranch: string;
  branchCount: number;
  commitCount: number;
  lastUpdated?: string;
}

export interface Branch {
  name: string;
  createdAt: string;
}

export interface Commit {
  hash: string;
  message: string;
  author: string;
  date: string;
}

/**
 * Get API base URL from environment or use default
 */
function getApiBaseUrl(): string {
  return process.env.GITSTORE_API_URL || DEFAULT_API_URL;
}

/**
 * Fetch JSON from API
 */
async function fetchJSON<T>(endpoint: string, options?: RequestInit): Promise<T> {
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
      errorMessage = (error as any).error || errorMessage;
    } catch {
      // Not JSON, use status text
    }
    throw new Error(errorMessage || `HTTP ${response.status}`);
  }

  return response.json() as Promise<T>;
}

/**
 * Check if GitStore API server is available
 */
export async function isApiAvailable(): Promise<boolean> {
  try {
    const apiUrl = getApiBaseUrl();
    const response = await fetch(`${apiUrl}/api/repos`, {
      method: 'GET',
      signal: AbortSignal.timeout(2000), // 2 second timeout
    });
    return response.ok || response.status === 404; // 404 is ok, means server is running
  } catch {
    return false;
  }
}

/**
 * Find repository ID by path
 * This searches through all repos to find one matching the path
 */
export async function findRepoByPath(repoPath: string): Promise<string | null> {
  try {
    const repos = await listRepos();
    // Try to match by name or path
    // For now, we'll need to check if the repo name matches the directory name
    const repoName = repoPath.split('/').pop() || repoPath;
    const repo = repos.find(r => r.id === repoName || r.name === repoName);
    return repo ? repo.id : null;
  } catch {
    return null;
  }
}

export const gitstoreApi = {
  async listRepos(): Promise<RepoListItem[]> {
    return fetchJSON<RepoListItem[]>('/api/repos');
  },

  async createRepo(name: string, description?: string): Promise<RepoListItem> {
    return fetchJSON<RepoListItem>('/api/repos', {
      method: 'POST',
      body: JSON.stringify({ name, description }),
    });
  },

  async getBranches(repoId: string): Promise<Branch[]> {
    return fetchJSON<Branch[]>(`/api/repos/${repoId}/branches`);
  },

  async getCommits(repoId: string): Promise<Commit[]> {
    return fetchJSON<Commit[]>(`/api/repos/${repoId}/commits`);
  },

  async checkout(repoId: string, branch: string): Promise<void> {
    await fetchJSON(`/api/repos/${repoId}/checkout`, {
      method: 'POST',
      body: JSON.stringify({ branch }),
    });
  },

  async commit(repoId: string, message: string): Promise<void> {
    await fetchJSON(`/api/repos/${repoId}/commit`, {
      method: 'POST',
      body: JSON.stringify({ message }),
    });
  },

  async merge(repoId: string, branch: string): Promise<{ message: string; type?: string }> {
    return fetchJSON<{ message: string; type?: string }>(`/api/repos/${repoId}/merge`, {
      method: 'POST',
      body: JSON.stringify({ branch }),
    });
  },
};

// Export for convenience
export const { listRepos, createRepo, getBranches, getCommits, checkout, commit, merge } = gitstoreApi;

