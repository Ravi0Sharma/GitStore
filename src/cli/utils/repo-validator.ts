import { existsSync, statSync } from 'fs';
import { join } from 'path';

/**
 * Validates that a given path is a git repository
 * Checks for the existence of .git directory
 */
export function validateGitRepo(repoPath: string): { valid: boolean; error?: string } {
  if (!repoPath) {
    return { valid: false, error: 'Repository path is required' };
  }

  // Resolve absolute path
  const absolutePath = repoPath.startsWith('/') 
    ? repoPath 
    : join(process.cwd(), repoPath);

  // Check if path exists
  if (!existsSync(absolutePath)) {
    return { valid: false, error: `Path does not exist: ${absolutePath}` };
  }

  // Check if it's a directory
  const stats = statSync(absolutePath);
  if (!stats.isDirectory()) {
    return { valid: false, error: `Path is not a directory: ${absolutePath}` };
  }

  // Check for .git directory
  const gitDir = join(absolutePath, '.git');
  if (!existsSync(gitDir)) {
    return { 
      valid: false, 
      error: `Not a git repository (missing .git directory): ${absolutePath}` 
    };
  }

  return { valid: true };
}

/**
 * Gets the absolute path of a repository
 */
export function getRepoPath(repoPath: string): string {
  return repoPath.startsWith('/') 
    ? repoPath 
    : join(process.cwd(), repoPath);
}

