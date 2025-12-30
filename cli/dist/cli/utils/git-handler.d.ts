import { SimpleGit } from 'simple-git';
/**
 * Handles git operations for both standard git and GitStore repositories
 */
export declare function handleGitOperation<T>(repoPath: string, operation: (git: SimpleGit) => Promise<T>, operationName: string): Promise<T>;
//# sourceMappingURL=git-handler.d.ts.map