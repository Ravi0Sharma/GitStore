import { Command } from 'commander';
import { validateGitRepo, getRepoPath } from '../utils/repo-validator.js';
import simpleGit, { SimpleGit } from 'simple-git';
import { existsSync } from 'fs';
import { join } from 'path';

export const initCommand = new Command('init')
  .description('Initialize a new git repository')
  .option('-r, --repo <path>', 'Repository path')
  .action(async (options) => {
    try {
      const repoPath = options.repo || process.cwd();
      const absolutePath = getRepoPath(repoPath);

      // Check if already a git repo
      const validation = validateGitRepo(absolutePath);
      if (validation.valid) {
        console.log(`✓ Already a git repository: ${absolutePath}`);
        return;
      }

      // Check if directory exists
      if (!existsSync(absolutePath)) {
        console.error(`✗ Error: Directory does not exist: ${absolutePath}`);
        process.exit(1);
      }

      // Initialize git repository
      const git: SimpleGit = simpleGit(absolutePath);
      await git.init();

      console.log(`✓ Initialized git repository in: ${absolutePath}`);
    } catch (error) {
      console.error(`✗ Error initializing repository:`, error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

