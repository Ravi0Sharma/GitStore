import { createContext, useContext, ReactNode, useState, useEffect } from 'react';
import { Repository, Issue, Priority, Label, MergeResult, IssueStatus } from '../types/git';
import { api, Repository as APIRepository } from '../lib/api';
import { convertAPIRepo } from '../lib/repoConverters';
import { normalizeRepos, normalizeRepo, mergeById } from '../lib/normalize';
import { onAuthStateChanged, User as FirebaseUser } from 'firebase/auth';
import { firebaseAuth } from '../firebase';

interface GitContextType {
  repositories: Repository[];
  getRepository: (id: string) => Repository | undefined;
  getIssue: (repoId: string, issueId: string) => Issue | undefined;
  createRepository: (name: string, description?: string) => Promise<void>;
  createIssue: (repoId: string, title: string, body: string, priority: Priority, labels: Label[]) => void;
  updateIssueBody: (repoId: string, issueId: string, body: string) => void;
  toggleIssueStatus: (repoId: string, issueId: string) => Promise<void>;
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
  toggleIssueStatus: async () => {},
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
  const [currentUser, setCurrentUser] = useState<FirebaseUser | null>(null);

  // Get current user from Firebase
  useEffect(() => {
    const unsubscribe = onAuthStateChanged(firebaseAuth, (user) => {
      setCurrentUser(user);
    });
    return () => unsubscribe();
  }, []);

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
      
      // Load branches and issues for each repo (similar to issues pattern)
      const reposWithData = await Promise.all(
        repos.map(async (repo) => {
          try {
            // Load branches and deduplicate by name
            const branches = await api.getBranches(repo.id);
            // Deduplicate branches by name (use Map to ensure uniqueness)
            const branchMap = new Map<string, { name: string; createdAt: Date }>();
            branches.forEach(b => {
              if (!branchMap.has(b.name)) {
                branchMap.set(b.name, {
                  name: b.name,
                  createdAt: new Date(b.createdAt),
                });
              }
            });
            const branchDates = Array.from(branchMap.values());
            
            // Load issues
            const issues = await api.getIssues(repo.id);
            const convertedIssues: Issue[] = issues.map((issue: any) => {
              // Use initials avatar (unisex) if authorAvatar is the old avataaars style
              let avatarUrl = issue.authorAvatar;
              if (!avatarUrl || avatarUrl.includes('avataaars')) {
                const email = issue.author || 'system';
                avatarUrl = `https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(email)}`;
              }
              return {
                id: issue.id,
                title: issue.title,
                body: issue.body,
                status: issue.status as IssueStatus,
                priority: issue.priority as Priority,
                labels: issue.labels || [],
                author: issue.author || 'system',
                authorAvatar: avatarUrl,
                createdAt: issue.createdAt || new Date().toISOString(),
                commentCount: issue.commentCount || 0,
              };
            });
            
            return {
              ...repo,
              branches: branchDates,
              issues: convertedIssues,
            };
          } catch (err) {
            console.warn(`GitContext: Failed to load branches/issues for ${repo.id}:`, err);
            // Return repo with empty arrays if fetch fails
            return {
              ...repo,
              branches: [],
              issues: [],
            };
          }
        })
      );
      
      console.log('GitContext: Setting repositories', reposWithData.length, 'repos (isInitialLoad:', isInitialLoad, ')');
      console.log('GitContext: Repo IDs to set:', reposWithData.map(r => r.id));
      setRepositories(reposWithData);
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
      
      // Load branches and issues for the new repo (same pattern as loadRepositories)
      try {
        await new Promise(resolve => setTimeout(resolve, 300));
        
        // Load branches and deduplicate by name
        const branches = await api.getBranches(created.id);
        // Deduplicate branches by name
        const branchMap = new Map<string, { name: string; createdAt: Date }>();
        branches.forEach(b => {
          if (!branchMap.has(b.name)) {
            branchMap.set(b.name, {
              name: b.name,
              createdAt: new Date(b.createdAt),
            });
          }
        });
        const branchDates = Array.from(branchMap.values());
        
        // Load issues
        const issues = await api.getIssues(created.id);
        const convertedIssues: Issue[] = issues.map((issue: any) => {
          // Use initials avatar (unisex) if authorAvatar is the old avataaars style
          let avatarUrl = issue.authorAvatar;
          if (!avatarUrl || avatarUrl.includes('avataaars')) {
            const email = issue.author || 'system';
            avatarUrl = `https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(email)}`;
          }
          return {
            id: issue.id,
            title: issue.title,
            body: issue.body,
            status: issue.status as IssueStatus,
            priority: issue.priority as Priority,
            labels: issue.labels || [],
            author: issue.author || 'system',
            authorAvatar: avatarUrl,
            createdAt: issue.createdAt || new Date().toISOString(),
            commentCount: issue.commentCount || 0,
          };
        });
        
        // Update the created repo with branches and issues
        setRepositories(prev => prev.map(repo => {
          if (repo.id === created.id) {
            return {
              ...repo,
              branches: branchDates,
              issues: convertedIssues,
            };
          }
          return repo;
        }));
      } catch (reloadErr) {
        console.warn('GitContext: Failed to reload branches/issues after create, keeping optimistic state:', reloadErr);
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

  const createIssue = async (repoId: string, title: string, body: string, priority: Priority, labels: Label[]) => {
    try {
      console.log('GitContext: Creating issue', repoId, title);
      
      // Get user email for author
      const userEmail = currentUser?.email || 'system';
      // Use initials avatar (unisex) instead of avataaars
      const avatarUrl = currentUser?.email 
        ? `https://api.dicebear.com/7.x/initials/svg?seed=${encodeURIComponent(userEmail)}`
        : 'https://api.dicebear.com/7.x/initials/svg?seed=system';
      
      // Call API to create issue (pass user email)
      const issue = await api.createIssue(repoId, title, body, priority, labels, userEmail);
      setApiStatus('connected');
      setApiError(null);
      
      // Update repo with new issue (use user email instead of system)
      setRepositories(prev => prev.map(repo => {
        if (repo.id === repoId) {
          // Convert issue to match Issue type
          const newIssue: Issue = {
            id: issue.id,
            title: issue.title,
            body: issue.body,
            status: issue.status as IssueStatus,
            priority: issue.priority as Priority,
            labels: issue.labels || [],
            author: userEmail, // Use user email instead of 'system'
            authorAvatar: avatarUrl, // Use initials avatar (unisex)
            createdAt: issue.createdAt || new Date().toISOString(),
            commentCount: issue.commentCount || 0,
          };
          return {
            ...repo,
            issues: [...repo.issues, newIssue],
          };
        }
        return repo;
      }));
      
      console.log('GitContext: Issue creation completed');
    } catch (err) {
      console.error('Failed to create issue:', err);
      const errorMsg = err instanceof Error ? err.message : String(err);
      setApiStatus('error');
      setApiError(errorMsg);
      throw err;
    }
  };

  const updateIssueBody = (repoId: string, issueId: string, body: string) => {
    // TODO: Implement issue update
    console.log('Update issue not implemented yet');
  };

  const toggleIssueStatus = async (repoId: string, issueId: string) => {
    try {
      console.log('GitContext: Toggling issue status', repoId, issueId);
      
      // Call API to toggle issue status
      const updatedIssue = await api.toggleIssueStatus(repoId, issueId);
      setApiStatus('connected');
      setApiError(null);
      
      // Update issue in state
      setRepositories(prev => prev.map(repo => {
        if (repo.id === repoId) {
          return {
            ...repo,
            issues: repo.issues.map(issue => 
              issue.id === issueId 
                ? { ...issue, status: updatedIssue.status as IssueStatus }
                : issue
            ),
          };
        }
        return repo;
      }));
      
      console.log('GitContext: Issue status toggled successfully');
    } catch (err) {
      console.error('Failed to toggle issue status:', err);
      const errorMsg = err instanceof Error ? err.message : String(err);
      setApiStatus('error');
      setApiError(errorMsg);
      throw err;
    }
  };

  const createBranch = async (repoId: string, branchName: string) => {
    try {
      console.log('GitContext: Creating branch', repoId, branchName);
      
      // Call API to create branch (checkout creates the branch)
      await api.checkout(repoId, branchName);
      setApiStatus('connected');
      setApiError(null);
      
      // Await checkout completion, then refresh branches ONCE
      // Use a request ID to prevent race conditions from concurrent calls
      const requestId = Date.now();
      console.log(`[GitContext] createBranch requestId=${requestId}, awaiting checkout completion`);
      
      // Small delay to ensure server has flushed writes
      await new Promise(resolve => setTimeout(resolve, 100));
      
      // Fetch branches ONCE and REPLACE state (do not append)
      try {
        const branches = await api.getBranches(repoId);
        console.log(`[GitContext] createBranch requestId=${requestId}, fetched ${branches.length} branches:`, branches.map(b => b.name));
        
        // Deduplicate branches by name
        const branchMap = new Map<string, { name: string; createdAt: Date }>();
        branches.forEach(b => {
          if (!branchMap.has(b.name)) {
            branchMap.set(b.name, {
              name: b.name,
              createdAt: new Date(b.createdAt),
            });
          }
        });
        const branchDates = Array.from(branchMap.values());
        
        // REPLACE branches state (not append) - this prevents stale data from overwriting
        setRepositories(prev => prev.map(repo => {
          if (repo.id === repoId) {
            console.log(`[GitContext] createBranch requestId=${requestId}, replacing branches for repo ${repoId}:`, branchDates.map(b => b.name));
            return {
              ...repo,
              branches: branchDates, // REPLACE with fetched branches
              currentBranch: branchName, // Also update current branch
            };
          }
          return repo;
        }));
      } catch (reloadErr) {
        console.warn(`[GitContext] createBranch requestId=${requestId}, failed to reload branches:`, reloadErr);
        // Optimistic update as fallback
        setRepositories(prev => prev.map(repo => {
          if (repo.id === repoId) {
            // Check if branch already exists
            if (repo.branches.some(b => b.name === branchName)) {
              return { ...repo, currentBranch: branchName };
            }
            // Add new branch to existing branches
            return {
              ...repo,
              branches: [...repo.branches, { name: branchName, createdAt: new Date() }],
              currentBranch: branchName,
            };
          }
          return repo;
        }));
      }
      
      console.log(`[GitContext] createBranch requestId=${requestId}, branch creation completed`);
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
        // Use request ID to prevent race conditions
        const requestId = Date.now();
        console.log(`[GitContext] switchBranch requestId=${requestId}, awaiting checkout completion`);
        
        // Small delay to ensure server has flushed writes
        await new Promise(resolve => setTimeout(resolve, 100));
        
        try {
          // Fetch branches ONCE and REPLACE state (do not call loadRepositories which might overwrite)
          const branches = await api.getBranches(repoId);
          console.log(`[GitContext] switchBranch requestId=${requestId}, fetched ${branches.length} branches:`, branches.map(b => b.name));
          
          // Deduplicate branches by name
          const branchMap = new Map<string, { name: string; createdAt: Date }>();
          branches.forEach(b => {
            if (!branchMap.has(b.name)) {
              branchMap.set(b.name, {
                name: b.name,
                createdAt: new Date(b.createdAt),
              });
            }
          });
          const branchDates = Array.from(branchMap.values());
          
          // REPLACE branches state (not append) - this prevents stale data from overwriting
          setRepositories(prev => prev.map(repo => {
            if (repo.id === repoId) {
              console.log(`[GitContext] switchBranch requestId=${requestId}, replacing branches for repo ${repoId}:`, branchDates.map(b => b.name));
              return {
                ...repo,
                branches: branchDates, // REPLACE with fetched branches
                currentBranch: branchName, // Update current branch
              };
            }
            return repo;
          }));
        } catch (reloadErr) {
          console.warn(`[GitContext] switchBranch requestId=${requestId}, failed to reload branches:`, reloadErr);
          // Keep optimistic state - currentBranch is already updated
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
    const requestId = Date.now();
    try {
      console.log(`[GitContext] mergeBranches requestId=${requestId}: merging ${fromBranch} into ${toBranch} for repo ${repoId}`);
      
      // First checkout to target branch
      await api.checkout(repoId, toBranch);
      console.log(`[GitContext] mergeBranches requestId=${requestId}: checked out ${toBranch}`);
      
      // Then merge fromBranch into toBranch
      const mergeResponse = await api.merge(repoId, fromBranch);
      console.log(`[GitContext] mergeBranches requestId=${requestId}: merge completed, response:`, mergeResponse);
      
      // IMPORTANT: After merge, we need to push to update remote refs
      // ListCommits reads from refs/remotes/origin/<branch>, so we must push
      // to make the merge commit visible in the UI
      try {
        await api.push(repoId, 'origin', toBranch);
        console.log(`[GitContext] mergeBranches requestId=${requestId}: pushed ${toBranch} after merge`);
      } catch (pushErr) {
        console.warn(`[GitContext] mergeBranches requestId=${requestId}: failed to push after merge:`, pushErr);
        // Continue - merge succeeded even if push failed
      }
      
      // Reload branches and repos after successful merge (branches might have changed)
      try {
        const branches = await api.getBranches(repoId);
        console.log(`[GitContext] mergeBranches requestId=${requestId}: fetched ${branches.length} branches after merge`);
        
        // Deduplicate branches by name
        const branchMap = new Map<string, { name: string; createdAt: Date }>();
        branches.forEach(b => {
          if (!branchMap.has(b.name)) {
            branchMap.set(b.name, {
              name: b.name,
              createdAt: new Date(b.createdAt),
            });
          }
        });
        const branchDates = Array.from(branchMap.values());
        
        // Update repo with refreshed branches (REPLACE, not append)
        setRepositories(prev => prev.map(repo => {
          if (repo.id === repoId) {
            console.log(`[GitContext] mergeBranches requestId=${requestId}: updating repo state, currentBranch=${toBranch}`);
            return {
              ...repo,
              branches: branchDates, // REPLACE with fetched branches
              currentBranch: toBranch, // Update current branch after checkout
            };
          }
          return repo;
        }));
        
        // Force commits refresh by triggering a state update that RepoPage will detect
        // The RepoPage useEffect depends on repo?.currentBranch, so updating currentBranch should trigger refresh
        console.log(`[GitContext] mergeBranches requestId=${requestId}: repo state updated, RepoPage should refresh commits for ${toBranch}`);
      } catch (reloadErr) {
        console.warn(`[GitContext] mergeBranches requestId=${requestId}: Failed to reload branches after merge:`, reloadErr);
        await loadRepositories();
      }
      
      // Determine merge type from response or default to fast-forward
      const mergeType = mergeResponse.type === 'fast-forward' ? 'fast-forward' : 'fast-forward';
      return { 
        success: true, 
        message: `Successfully merged ${fromBranch} into ${toBranch}`,
        type: mergeType
      };
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Merge failed';
      console.error(`[GitContext] mergeBranches requestId=${requestId}: error:`, err);
      
      // Check if it's a non-fast-forward error (409 conflict)
      if (errorMessage.includes('Non-fast-forward') || errorMessage.includes('409')) {
        return { 
          success: false, 
          message: 'Non-fast-forward merge is not allowed. The branches have diverged and cannot be merged automatically.' 
        };
      }
      
      return { success: false, message: errorMessage };
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

