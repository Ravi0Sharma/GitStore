/**
 * Whitelist of allowed git subcommands for security
 */
export declare const ALLOWED_GIT_COMMANDS: readonly ["status", "add", "commit", "push", "checkout", "branch", "log", "diff", "pull", "fetch", "merge", "clone", "init"];
export type AllowedGitCommand = typeof ALLOWED_GIT_COMMANDS[number];
/**
 * Validates if a git command is in the whitelist
 */
export declare function isAllowedGitCommand(command: string): command is AllowedGitCommand;
/**
 * Gets a list of all allowed git commands
 */
export declare function getAllowedGitCommands(): string[];
//# sourceMappingURL=git-whitelist.d.ts.map