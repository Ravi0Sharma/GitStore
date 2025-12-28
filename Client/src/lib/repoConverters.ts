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
    issues: [], // Issues not implemented in backend yet
  };
}

