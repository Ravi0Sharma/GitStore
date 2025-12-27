import { createContext, useContext, ReactNode } from 'react';
import { Repository, Issue, Priority, Label, MergeResult } from '../types/git';

interface GitContextType {
  repositories: Repository[];
  getRepository: (id: string) => Repository | undefined;
  getIssue: (repoId: string, issueId: string) => Issue | undefined;
  createRepository: (name: string, description?: string) => void;
  createIssue: (repoId: string, title: string, body: string, priority: Priority, labels: Label[]) => void;
  updateIssueBody: (repoId: string, issueId: string, body: string) => void;
  toggleIssueStatus: (repoId: string, issueId: string) => void;
  createBranch: (repoId: string, branchName: string) => void;
  switchBranch: (repoId: string, branchName: string) => void;
  mergeBranches: (repoId: string, fromBranch: string, toBranch: string) => MergeResult;
  loading: boolean;
  error: string | null;
}

const GitContext = createContext<GitContextType | undefined>(undefined);

export function useGit(): GitContextType {
  const context = useContext(GitContext);
  if (!context) {
    // Return placeholder implementation when context is not available
    return {
      repositories: [],
      getRepository: () => undefined,
      getIssue: () => undefined,
      createRepository: () => {},
      createIssue: () => {},
      updateIssueBody: () => {},
      toggleIssueStatus: () => {},
      createBranch: () => {},
      switchBranch: () => {},
      mergeBranches: () => ({ success: false, message: 'Git functionality not available' }),
      loading: false,
      error: null,
    };
  }
  return context;
}

interface GitProviderProps {
  children: ReactNode;
}

export function GitProvider({ children }: GitProviderProps) {
  // Placeholder implementation - returns empty data
  const value: GitContextType = {
    repositories: [],
    getRepository: () => undefined,
    getIssue: () => undefined,
    createRepository: () => {},
    createIssue: () => {},
    updateIssueBody: () => {},
    toggleIssueStatus: () => {},
    createBranch: () => {},
    switchBranch: () => {},
    mergeBranches: () => ({ success: false, message: 'Git functionality not available' }),
    loading: false,
    error: null,
  };

  return <GitContext.Provider value={value}>{children}</GitContext.Provider>;
}

