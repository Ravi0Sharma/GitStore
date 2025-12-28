/**
 * Validates that a given path is a git repository
 * Checks for the existence of .git directory
 */
export declare function validateGitRepo(repoPath: string): {
    valid: boolean;
    error?: string;
};
/**
 * Gets the absolute path of a repository
 */
export declare function getRepoPath(repoPath: string): string;
//# sourceMappingURL=repo-validator.d.ts.map