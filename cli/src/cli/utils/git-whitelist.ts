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
] as const;

export type AllowedGitCommand = typeof ALLOWED_GIT_COMMANDS[number];

/**
 * Validates if a git command is in the whitelist
 */
export function isAllowedGitCommand(command: string): command is AllowedGitCommand {
  return ALLOWED_GIT_COMMANDS.includes(command as AllowedGitCommand);
}

/**
 * Gets a list of all allowed git commands
 */
export function getAllowedGitCommands(): string[] {
  return [...ALLOWED_GIT_COMMANDS];
}

