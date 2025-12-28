/**
 * Helper functions for converting API types to UI types
 * Moved from GitContext.tsx to ensure stable exports for Fast Refresh
 */

import { Repository, Branch, Commit } from '../types/git';
import { Repository as APIRepository, Branch as APIBranch, Commit as APICommit } from './api';

/**
 * Convert API repository format to UI repository format
 */
export function convertAPIRepo(apiRepo: APIRepository): Repository {
  return {
    id: apiRepo.id,
    name: apiRepo.name,
    description: apiRepo.description,
    currentBranch: apiRepo.currentBranch,
    branches: apiRepo.branches.map((b: APIBranch) => ({
      name: b.name,
      createdAt: new Date(b.createdAt),
    })),
    commits: apiRepo.commits.map((c: APICommit) => ({
      hash: c.hash,
      message: c.message,
      author: c.author,
      date: new Date(c.date),
    })),
    issues: apiRepo.issues ? apiRepo.issues.map((issue: any) => ({
      id: issue.id,
      title: issue.title,
      body: issue.body,
      status: issue.status,
      priority: issue.priority,
      labels: issue.labels || [],
      author: issue.author || 'system',
      authorAvatar: issue.authorAvatar || 'https://api.dicebear.com/7.x/avataaars/svg?seed=system',
      createdAt: issue.createdAt || new Date().toISOString(),
      commentCount: issue.commentCount || 0,
    })) : [],
  };
}

