/**
 * Normalize backend repository data to frontend Repository format
 * Handles field mapping and ensures all required fields are present
 */

import { Repository, Branch, Commit } from '../types/git';
import { RepoListItem } from './api';

/**
 * Normalize a RepoListItem from API to Repository format
 */
export function normalizeRepo(repo: RepoListItem): Repository {
  // Extract date - prefer lastUpdated, fallback to updatedAt, then createdAt
  const lastUpdated = repo.lastUpdated || (repo as any).updatedAt || (repo as any).createdAt || undefined;
  
  // Ensure branches and commits are arrays (even if empty)
  // Backend listRepos doesn't include full branch/commit data, so we use counts
  const branches: Branch[] = [];
  const commits: Commit[] = [];
  
  return {
    id: repo.id,
    name: repo.name,
    description: repo.description,
    currentBranch: repo.currentBranch || 'master',
    branches,
    commits,
    issues: [], // Issues not implemented in backend yet
    branchCount: repo.branchCount,
    commitCount: repo.commitCount,
    lastUpdated,
  };
}

/**
 * Normalize an array of RepoListItems
 */
export function normalizeRepos(repos: RepoListItem[]): Repository[] {
  return repos.map(normalizeRepo);
}

/**
 * Merge repositories by ID, keeping existing data when possible
 */
export function mergeById(existing: Repository[], newRepo: Repository): Repository[] {
  const existingMap = new Map(existing.map(r => [r.id, r]));
  existingMap.set(newRepo.id, newRepo);
  return Array.from(existingMap.values());
}

