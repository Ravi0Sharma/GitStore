import { detectRepoType, getRepoPath } from './repo-validator.js';
import { isApiAvailable } from './gitstore-api.js';
import simpleGit, { SimpleGit } from 'simple-git';

/**
 * Handles git operations for both standard git and GitStore repositories
 */
export async function handleGitOperation<T>(
  repoPath: string,
  operation: (git: SimpleGit) => Promise<T>,
  operationName: string
): Promise<T> {
  const absoluteRepoPath = getRepoPath(repoPath);
  const repoType = detectRepoType(absoluteRepoPath);

  if (repoType === 'gitstore') {
    // For GitStore repos, we need to use the API
    const apiAvailable = await isApiAvailable();
    if (!apiAvailable) {
      console.error(`✗ GitStore repository detected but API server is not available.`);
      console.error(`  Operation: ${operationName}`);
      console.error(`  Please start the GitStore server (gitserver) to use GitStore repositories.`);
      console.error(`  Or use a standard git repository (.git) instead.`);
      console.error(`\n  To start the server:`);
      console.error(`    cd gitClone && go build -o gitserver ./cmd/gitserver && ./gitserver`);
      throw new Error('GitStore API server not available');
    }
    // For now, inform user that GitStore operations need API
    console.error(`✗ GitStore repositories require the API server for git operations.`);
    console.error(`  Operation: ${operationName}`);
    console.error(`  Please use the web interface or API directly.`);
    console.error(`  For standard git operations, use a .git repository instead.`);
    throw new Error('GitStore git operations require API (not yet fully implemented in CLI)');
  }

  // Use standard git for .git repositories
  const git: SimpleGit = simpleGit(absoluteRepoPath);
  return operation(git);
}

