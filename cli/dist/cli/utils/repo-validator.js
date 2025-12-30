import { existsSync, statSync } from 'fs';
import { join } from 'path';
/**
 * Detects the type of repository (GitStore .gitclone or standard .git)
 */
export function detectRepoType(repoPath) {
    if (!repoPath) {
        return 'none';
    }
    const absolutePath = repoPath.startsWith('/')
        ? repoPath
        : join(process.cwd(), repoPath);
    if (!existsSync(absolutePath)) {
        return 'none';
    }
    const stats = statSync(absolutePath);
    if (!stats.isDirectory()) {
        return 'none';
    }
    // Check for GitStore repository (.gitclone)
    const gitcloneDir = join(absolutePath, '.gitclone');
    if (existsSync(gitcloneDir)) {
        return 'gitstore';
    }
    // Check for standard git repository (.git)
    const gitDir = join(absolutePath, '.git');
    if (existsSync(gitDir)) {
        return 'git';
    }
    return 'none';
}
/**
 * Validates that a given path is a git repository (either GitStore or standard git)
 * Checks for the existence of .gitclone or .git directory
 */
export function validateGitRepo(repoPath) {
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
    // Detect repository type
    const repoType = detectRepoType(absolutePath);
    if (repoType === 'none') {
        return {
            valid: false,
            error: `Not a git repository (missing .git or .gitclone directory): ${absolutePath}`,
            type: 'none'
        };
    }
    return { valid: true, type: repoType };
}
/**
 * Gets the absolute path of a repository
 */
export function getRepoPath(repoPath) {
    return repoPath.startsWith('/')
        ? repoPath
        : join(process.cwd(), repoPath);
}
//# sourceMappingURL=repo-validator.js.map