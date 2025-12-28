import { createContext, useContext, ReactNode, useState, useEffect } from 'react';
import { Repository, Issue, Priority, Label, MergeResult } from '../types/git';
import { api, Repository as APIRepository } from '../lib/api';
import { convertAPIRepo } from '../lib/repoConverters';
import { normalizeRepos, normalizeRepo, mergeById } from '../lib/normalize';

interface GitContextType {
  repositories: Repository[];
  getRepository: (id: string) => Repository | undefined;
  getIssue: (repoId: string, issueId: string) => Issue | undefined;
  createRepository: (name: string, description?: string) => Promise<void>;
  createIssue: (repoId: string, title: string, body: string, priority: Priority, labels: Label[]) => void;
  updateIssueBody: (repoId: string, issueId: string, body: string) => void;
  toggleIssueStatus: (repoId: string, issueId: string) => void;
  createBranch: (repoId: string, branchName: string) => Promise<void>;
  switchBranch: (repoId: string, branchName: string) => Promise<void>;
  mergeBranches: (repoId: string, fromBranch: string, toBranch: string) => Promise<MergeResult>;
  loadRepositories: (skipEmptyGuard?: boolean) => Promise<void>;
  loading: boolean;
  error: string | null;
  apiStatus: 'unknown' | 'connected' | 'disconnected' | 'error';
  apiError: string | null;
}

const GitContext = createContext<GitContextType | undefined>(undefined);

// Stable placeholder for useGit when context is not available
// Defined as const outside component to ensure Fast Refresh compatibility
const PLACEHOLDER_GIT_CONTEXT: GitContextType = {
  repositories: [],
  getRepository: () => undefined,
  getIssue: () => undefined,
  createRepository: async () => {},
  createIssue: () => {},
  updateIssueBody: () => {},
  toggleIssueStatus: () => {},
  createBranch: async () => {},
  switchBranch: async () => {},
  mergeBranches: async () => Promise.resolve({ success: false, message: 'Git functionality not available' }),
  loadRepositories: async () => {},
  loading: false,
  error: null,
  apiStatus: 'unknown',
  apiError: null,
};

export function useGit(): GitContextType {
  const context = useContext(GitContext);
  if (!context) {
    // Return stable placeholder implementation when context is not available
    return PLACEHOLDER_GIT_CONTEXT;
  }
  return context;
}

interface GitProviderProps {
  children: ReactNode;
}

export function GitProvider({ children }: GitProviderProps) {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [pendingOptimisticRepo, setPendingOptimisticRepo] = useState<string | null>(null);
  const [apiStatus, setApiStatus] = useState<'unknown' | 'connected' | 'disconnected' | 'error'>('unknown');
  const [apiError, setApiError] = useState<string | null>(null);

  useEffect(() => {
    loadRepositories();
  }, []);

  // Debug: Log when repositories change
  useEffect(() => {
    console.log('GitContext: Repositories changed', repositories.length, repositories.map(r => r.id));
  }, [repositories]);

  const loadRepositories = async (skipEmptyGuard = false) => {
    try {
      setLoading(true);
      setError(null);
      setApiStatus('unknown');
      setApiError(null);
      console.log('GitContext: loadRepositories called, skipEmptyGuard:', skipEmptyGuard);
      console.log('GitContext: Current repositories in state:', repositories.length);
      
      const repoList = await api.listRepos();
      console.log('loadRepositories repos:', repoList);
      console.log('GitContext: api.listRepos returned', repoList.length, 'repos');
      
      // Verify repoList is actually an array
      if (!Array.isArray(repoList)) {
        console.error('GitContext: repoList is not an array!', typeof repoList, repoList);
        throw new Error('API returned invalid data: expected array but got ' + typeof repoList);
      }
      
      // API is working
      setApiStatus('connected');
      setApiError(null);
      
      // Guard: Don't overwrite state with empty array if we have optimistic updates pending
      if (!skipEmptyGuard && repoList.length === 0 && repositories.length > 0) {
        console.warn('GitContext: listRepos returned empty array but we have', repositories.length, 'repos in state. Skipping update to preserve optimistic state.');
        if (pendingOptimisticRepo) {
          console.warn('GitContext: Pending optimistic repo:', pendingOptimisticRepo);
        }
        setLoading(false);
        return;
      }
      
      // Normalize repos from list (no need to load full details - list has all we need)
      const repos = normalizeRepos(repoList);
      
      console.log('GitContext: Setting repositories', repos.length, 'repos (isInitialLoad:', isInitialLoad, ')');
      console.log('GitContext: Repo IDs to set:', repos.map(r => r.id));
      setRepositories(repos);
      console.log('GitContext: Repositories state updated');
      setIsInitialLoad(false);
      setPendingOptimisticRepo(null);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load repositories';
      setError(errorMessage);
      setApiStatus('error');
      setApiError(errorMessage);
      console.error('GitContext: Failed to load repositories:', err);
      
      // Check if it's a network/API error
      if (errorMessage.includes('Failed to reach API') || errorMessage.includes('HTML instead of JSON')) {
        setApiStatus('disconnected');
      }
    } finally {
      setLoading(false);
    }
  };

  const getRepository = (id: string): Repository | undefined => {
    return repositories.find(r => r.id === id);
  };

  const getIssue = (repoId: string, issueId: string): Issue | undefined => {
    const repo = getRepository(repoId);
    return repo?.issues.find(i => i.id === issueId);
  };

  const createRepository = async (name: string, description?: string) => {
    let created: any = null;
    try {
      console.log('GitContext: Creating repository', name, description);
      
      // Create via API
      created = await api.createRepo(name, description);
      console.log('GitContext: Repository created via API', created);
      setApiStatus('connected');
      setApiError(null);
      
      // Normalize and merge into state
      const normalized = normalizeRepo(created);
      setRepositories(prev => mergeById(prev, normalized));
      
      // Optionally reload to sync (but don't overwrite with empty)
      try {
        await new Promise(resolve => setTimeout(resolve, 300));
        const repoList = await api.listRepos();
        if (repoList.length > 0) {
          const repos = normalizeRepos(repoList);
          setRepositories(repos);
        }
      } catch (reloadErr) {
        console.warn('GitContext: Failed to reload after create, keeping optimistic state:', reloadErr);
        // Keep the created repo in state
      }
      
      console.log('GitContext: Repository creation completed');
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      console.error('GitContext: Failed to create repository:', err);
      
      // Check if it's a connection error
      if (errorMsg.includes('Failed to reach API') || errorMsg.includes('HTML instead of JSON')) {
        setApiStatus('disconnected');
        setApiError(errorMsg);
      } else {
        setApiStatus('error');
        setApiError(errorMsg);
      }
      
      throw err;
    }
  };

  const createIssue = (repoId: string, title: string, body: string, priority: Priority, labels: Label[]) => {
    // TODO: Implement issue creation
    console.log('Create issue not implemented yet');
  };

  const updateIssueBody = (repoId: string, issueId: string, body: string) => {
    // TODO: Implement issue update
    console.log('Update issue not implemented yet');
  };

  const toggleIssueStatus = (repoId: string, issueId: string) => {
    // TODO: Implement issue status toggle
    console.log('Toggle issue status not implemented yet');
  };

  const createBranch = async (repoId: string, branchName: string) => {
    try {
      console.log('GitContext: Creating branch', repoId, branchName);
      
      // Call API to create branch
      await api.createBranch(repoId, branchName);
      setApiStatus('connected');
      setApiError(null);
      
      // Reload branches from API
      try {
        const branches = await api.getBranches(repoId);
        const branchDates = branches.map(b => ({
          name: b.name,
          createdAt: new Date(b.createdAt),
        }));
        
        // Update repo with new branches
        setRepositories(prev => prev.map(repo => {
          if (repo.id === repoId) {
            return {
              ...repo,
              branches: branchDates,
            };
          }
          return repo;
        }));
      } catch (reloadErr) {
        console.warn('GitContext: Failed to reload branches after create, using optimistic update:', reloadErr);
        // Optimistic update as fallback
        setRepositories(prev => prev.map(repo => {
          if (repo.id === repoId) {
            if (repo.branches.some(b => b.name === branchName)) {
              return repo;
            }
            return {
              ...repo,
              branches: [...repo.branches, { name: branchName, createdAt: new Date() }],
            };
          }
          return repo;
        }));
      }
      
      console.log('GitContext: Branch creation completed');
    } catch (err) {
      console.error('Failed to create branch:', err);
      const errorMsg = err instanceof Error ? err.message : String(err);
      setApiStatus('error');
      setApiError(errorMsg);
      throw err;
    }
  };

  const switchBranch = async (repoId: string, branchName: string) => {
    try {
      console.log('GitContext: Switching branch', repoId, branchName);
      
      // Try API first
      let apiWorked = false;
      try {
        await api.checkout(repoId, branchName);
        apiWorked = true;
        setApiStatus('connected');
        setApiError(null);
      } catch (apiErr) {
        const errorMsg = apiErr instanceof Error ? apiErr.message : String(apiErr);
        if (errorMsg.includes('Failed to reach API') || errorMsg.includes('HTML instead of JSON')) {
          console.warn('GitContext: API not available, switching branch locally only');
          setApiStatus('disconnected');
          setApiError(errorMsg);
          // Continue with local-only update
        } else {
          throw apiErr;
        }
      }
      
      // Optimistic update: Update currentBranch immediately
      setRepositories(prev => prev.map(repo => {
        if (repo.id === repoId) {
          console.log('GitContext: Updating currentBranch optimistically', branchName);
          return {
            ...repo,
            currentBranch: branchName,
          };
        }
        return repo;
      }));
      
      // Only reload if API worked
      if (apiWorked) {
        // Wait a bit for server to be ready
        await new Promise(resolve => setTimeout(resolve, 300));
        
        try {
          // Then reload to get accurate data (with guard)
          const repoList = await api.listRepos();
          console.log('GitContext: After switchBranch, listRepos returned', repoList.length, 'repos');
          
          if (repoList.length > 0) {
            await loadRepositories(true); // skipEmptyGuard = true
          } else {
            console.warn('GitContext: listRepos returned empty after switchBranch, keeping optimistic state');
          }
        } catch (reloadErr) {
          console.warn('GitContext: Failed to reload after switchBranch, keeping optimistic state');
        }
      }
      console.log('GitContext: Branch switch completed');
    } catch (err) {
      console.error('Failed to switch branch:', err);
      // Note: We don't revert currentBranch on error as it's harder to track previous state
      throw err;
    }
  };

  const mergeBranches = async (repoId: string, fromBranch: string, toBranch: string): Promise<MergeResult> => {
    try {
      // First checkout to target branch
      await api.checkout(repoId, toBranch);
      // Then merge fromBranch into toBranch
      await api.merge(repoId, fromBranch);
      await loadRepositories(); // Reload to get updated state
      return { success: true, message: `Successfully merged ${fromBranch} into ${toBranch}` };
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Merge failed';
      return { success: false, message };
    }
  };

  const value: GitContextType = {
    repositories,
    getRepository,
    getIssue,
    createRepository,
    createIssue,
    updateIssueBody,
    toggleIssueStatus,
    createBranch,
    switchBranch,
    mergeBranches,
    loadRepositories,
    loading,
    error,
    apiStatus,
    apiError,
  };

  return <GitContext.Provider value={value}>{children}</GitContext.Provider>;
}

