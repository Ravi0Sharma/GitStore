export type IssueStatus = 'open' | 'closed';
export type Priority = 'low' | 'medium' | 'high';

export interface Label {
  id: string;
  name: string;
  color: string;
}

export const DEFAULT_LABELS: Label[] = [
  { id: 'bug', name: 'bug', color: '#d73a4a' },
  { id: 'enhancement', name: 'enhancement', color: '#a2eeef' },
  { id: 'documentation', name: 'documentation', color: '#0075ca' },
  { id: 'question', name: 'question', color: '#d876e3' },
];

export interface Issue {
  id: string;
  title: string;
  body: string;
  status: IssueStatus;
  priority: Priority;
  labels: Label[];
  author: string;
  authorAvatar: string;
  createdAt: string;
  commentCount: number;
}

export interface Commit {
  hash: string;
  message: string;
  author: string;
  date: Date;
}

export interface Branch {
  name: string;
  createdAt: Date;
}

export interface Repository {
  id: string;
  name: string;
  description?: string;
  currentBranch: string;
  branches: Branch[];
  issues: Issue[];
  commits: Commit[];
  // Optional counts from backend (when full branch/commit data not loaded)
  branchCount?: number;
  commitCount?: number;
  // Optional timestamp for "last updated" display
  lastUpdated?: string;
}

export interface MergeResult {
  success: boolean;
  message: string;
  type?: 'fast-forward' | 'merge-commit';
}

