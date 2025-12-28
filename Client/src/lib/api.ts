/**
 * API client for GitStore backend
 */

// API base URL: use env var if set, otherwise use relative paths (Vite proxy)
const API_BASE = import.meta.env.VITE_API_URL || '';
const API_URL = API_BASE ? `${API_BASE}` : '';

// Detect if response is HTML (likely Vite index.html fallback)
function isHTMLResponse(response: Response): Promise<boolean> {
  return response.clone().text().then(text => {
    return text.trim().startsWith('<!DOCTYPE') || text.trim().startsWith('<html');
  });
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const fullUrl = API_URL ? `${API_URL}${url}` : url;
  const method = options?.method || 'GET';
  
  // Dev logging
  if (import.meta.env.DEV) {
    console.log(`API: ${method} ${fullUrl}`);
  }

  let response: Response;
  try {
    response = await fetch(fullUrl, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Network error');
    if (import.meta.env.DEV) {
      console.error(`API: ${method} ${fullUrl} - Network error:`, error);
    }
    throw new Error(`Failed to reach API: ${error.message}. Is the backend running on ${API_BASE || 'http://localhost:8080'}?`);
  }

  // Check if we got HTML instead of JSON (Vite fallback)
  const isHTML = await isHTMLResponse(response);
  if (isHTML) {
    const errorMsg = `API returned HTML instead of JSON. This usually means the backend is not running or the URL is wrong. Tried: ${fullUrl}`;
    if (import.meta.env.DEV) {
      console.error(`API: ${method} ${fullUrl} - Got HTML response (likely Vite index.html)`);
    }
    throw new Error(errorMsg);
  }

  if (import.meta.env.DEV) {
    console.log(`API: ${method} ${fullUrl} - Status: ${response.status} ${response.statusText}`);
  }

  if (!response.ok) {
    let errorMessage = response.statusText;
    try {
      const error = await response.json();
      errorMessage = error.error || errorMessage;
    } catch {
      // Not JSON, use status text
    }
    const error = new Error(errorMessage || `HTTP ${response.status}`);
    if (import.meta.env.DEV) {
      console.error(`API: ${method} ${fullUrl} - Error:`, error);
    }
    throw error;
  }

  try {
    const data = await response.json();
    if (import.meta.env.DEV) {
      console.log(`API: ${method} ${fullUrl} - Success:`, Array.isArray(data) ? `Array(${data.length})` : 'Object');
    }
    return data;
  } catch (err) {
    if (import.meta.env.DEV) {
      console.error(`API: ${method} ${fullUrl} - JSON parse error:`, err);
    }
    throw new Error('Invalid JSON response from API');
  }
}

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

export interface Repository {
  id: string;
  name: string;
  description?: string;
  currentBranch: string;
  branches: Branch[];
  commits: Commit[];
  issues: any[];
}

export const api = {
  async listRepos(): Promise<RepoListItem[]> {
    const data = await fetchJSON<any>('/api/repos');
    
    // Dev log: show what we got
    console.log('api.listRepos got', data);
    
    // Ensure we return an array (never null or undefined)
    if (!Array.isArray(data)) {
      console.error('API: listRepos - Response is not an array:', typeof data, data);
      // If data is null/undefined, return empty array
      if (data == null) {
        console.warn('API: listRepos - Got null/undefined, returning empty array');
        return [];
      }
      // If data is wrapped in an object (e.g. {repos: [...]}), try to extract it
      if (typeof data === 'object' && 'repos' in data && Array.isArray(data.repos)) {
        console.warn('API: listRepos - Response wrapped in object, extracting repos array');
        return data.repos;
      }
      if (typeof data === 'object' && 'repositories' in data && Array.isArray(data.repositories)) {
        console.warn('API: listRepos - Response wrapped in object, extracting repositories array');
        return data.repositories;
      }
      // Last resort: return empty array
      console.error('API: listRepos - Could not extract array from response, returning empty array');
      return [];
    }
    
    console.log('API: listRepos - Successfully parsed array with', data.length, 'items');
    return data;
  },

  async createRepo(name: string, description?: string): Promise<RepoListItem> {
    return fetchJSON<RepoListItem>('/api/repos', {
      method: 'POST',
      body: JSON.stringify({ name, description }),
    });
  },

  async getRepo(repoId: string): Promise<Repository> {
    return fetchJSON<Repository>(`/api/repos/${repoId}`);
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

  async merge(repoId: string, branch: string): Promise<void> {
    await fetchJSON(`/api/repos/${repoId}/merge`, {
      method: 'POST',
      body: JSON.stringify({ branch }),
    });
  },
};

