export type RepoType = 'gitstore' | 'git' | 'none';
/**
 * Detects the type of repository (GitStore .gitclone or standard .git)
 */
export declare function detectRepoType(repoPath: string): RepoType;
/**
 * Validates that a given path is a git repository (either GitStore or standard git)
 * Checks for the existence of .gitclone or .git directory
 */
export declare function validateGitRepo(repoPath: string): {
    valid: boolean;
    error?: string;
    type?: RepoType;
};
/**
 * Gets the absolute path of a repository
 */
export declare function getRepoPath(repoPath: string): string;
//# sourceMappingURL=repo-validator.d.ts.map