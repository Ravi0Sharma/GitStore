/**
 * Whitelist of allowed git subcommands for security
 */
export const ALLOWED_GIT_COMMANDS = [
    'status',
    'add',
    'commit',
    'push',
    'checkout',
    'branch',
    'log',
    'diff',
    'pull',
    'fetch',
    'merge',
    'clone',
    'init',
];
/**
 * Validates if a git command is in the whitelist
 */
export function isAllowedGitCommand(command) {
    return ALLOWED_GIT_COMMANDS.includes(command);
}
/**
 * Gets a list of all allowed git commands
 */
export function getAllowedGitCommands() {
    return [...ALLOWED_GIT_COMMANDS];
}
//# sourceMappingURL=git-whitelist.js.map