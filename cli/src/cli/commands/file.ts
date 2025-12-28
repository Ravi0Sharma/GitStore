import { Command } from 'commander';
import { validateGitRepo, getRepoPath, detectRepoType } from '../utils/repo-validator.js';
import { writeFile, appendFile, mkdir } from 'fs/promises';
import { existsSync } from 'fs';
import { join, dirname } from 'path';
import inquirer from 'inquirer';

const fileCommand = new Command('file')
  .description('File operations (create, write, append)');

// Create file command
fileCommand
  .command('create')
  .description('Create a new file in the repository')
  .option('-r, --repo <path>', 'Repository path')
  .option('-p, --path <path>', 'File path relative to repository')
  .option('-c, --content <content>', 'File content')
  .option('-i, --interactive', 'Use interactive mode')
  .action(async (options) => {
    try {
      let repoPath = options.repo;
      let filePath = options.path;
      let content = options.content;

      // Interactive mode
      if (options.interactive || !repoPath || !filePath) {
        const answers = await inquirer.prompt([
          {
            type: 'input',
            name: 'repo',
            message: 'Repository path:',
            default: repoPath || process.cwd(),
            when: !repoPath,
          },
          {
            type: 'input',
            name: 'path',
            message: 'File path (relative to repo):',
            when: !filePath,
            validate: (input) => input.length > 0 || 'File path is required',
          },
          {
            type: 'editor',
            name: 'content',
            message: 'File content (opens editor):',
            when: !content,
            default: '',
          },
        ]);

        repoPath = repoPath || answers.repo;
        filePath = filePath || answers.path;
        content = content || answers.content || '';
      }

      // Validate repository
      const absoluteRepoPath = getRepoPath(repoPath);
      const validation = validateGitRepo(absoluteRepoPath);
      if (!validation.valid) {
        console.error(`✗ ${validation.error}`);
        process.exit(1);
      }

      // File operations work with both git and GitStore repos
      // (they just modify the filesystem, no git operations needed)

      // Create file
      const absoluteFilePath = join(absoluteRepoPath, filePath);
      const fileDir = dirname(absoluteFilePath);

      // Create directory if it doesn't exist
      if (!existsSync(fileDir)) {
        await mkdir(fileDir, { recursive: true });
      }

      await writeFile(absoluteFilePath, content || '', 'utf8');
      console.log(`✓ Created file: ${filePath}`);
      console.log(`  Location: ${absoluteFilePath}`);
    } catch (error) {
      console.error(`✗ Error creating file:`, error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

// Write file command (overwrites existing)
fileCommand
  .command('write')
  .description('Write content to a file (overwrites existing)')
  .option('-r, --repo <path>', 'Repository path')
  .option('-p, --path <path>', 'File path relative to repository')
  .option('-c, --content <content>', 'File content')
  .option('-i, --interactive', 'Use interactive mode')
  .action(async (options) => {
    try {
      let repoPath = options.repo;
      let filePath = options.path;
      let content = options.content;

      // Interactive mode
      if (options.interactive || !repoPath || !filePath) {
        const answers = await inquirer.prompt([
          {
            type: 'input',
            name: 'repo',
            message: 'Repository path:',
            default: repoPath || process.cwd(),
            when: !repoPath,
          },
          {
            type: 'input',
            name: 'path',
            message: 'File path (relative to repo):',
            when: !filePath,
            validate: (input) => input.length > 0 || 'File path is required',
          },
          {
            type: 'editor',
            name: 'content',
            message: 'File content (opens editor):',
            when: !content,
            default: '',
          },
        ]);

        repoPath = repoPath || answers.repo;
        filePath = filePath || answers.path;
        content = content || answers.content || '';
      }

      // Validate repository
      const absoluteRepoPath = getRepoPath(repoPath);
      const validation = validateGitRepo(absoluteRepoPath);
      if (!validation.valid) {
        console.error(`✗ ${validation.error}`);
        process.exit(1);
      }

      // Write file
      const absoluteFilePath = join(absoluteRepoPath, filePath);
      const fileDir = dirname(absoluteFilePath);

      // Create directory if it doesn't exist
      if (!existsSync(fileDir)) {
        await mkdir(fileDir, { recursive: true });
      }

      await writeFile(absoluteFilePath, content || '', 'utf8');
      console.log(`✓ Wrote to file: ${filePath}`);
      console.log(`  Location: ${absoluteFilePath}`);
    } catch (error) {
      console.error(`✗ Error writing file:`, error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

// Append file command
fileCommand
  .command('append')
  .description('Append content to a file')
  .option('-r, --repo <path>', 'Repository path')
  .option('-p, --path <path>', 'File path relative to repository')
  .option('-c, --content <content>', 'Content to append')
  .option('-i, --interactive', 'Use interactive mode')
  .action(async (options) => {
    try {
      let repoPath = options.repo;
      let filePath = options.path;
      let content = options.content;

      // Interactive mode
      if (options.interactive || !repoPath || !filePath) {
        const answers = await inquirer.prompt([
          {
            type: 'input',
            name: 'repo',
            message: 'Repository path:',
            default: repoPath || process.cwd(),
            when: !repoPath,
          },
          {
            type: 'input',
            name: 'path',
            message: 'File path (relative to repo):',
            when: !filePath,
            validate: (input) => input.length > 0 || 'File path is required',
          },
          {
            type: 'input',
            name: 'content',
            message: 'Content to append:',
            when: !content,
            default: '',
          },
        ]);

        repoPath = repoPath || answers.repo;
        filePath = filePath || answers.path;
        content = content || answers.content || '';
      }

      // Validate repository
      const absoluteRepoPath = getRepoPath(repoPath);
      const validation = validateGitRepo(absoluteRepoPath);
      if (!validation.valid) {
        console.error(`✗ ${validation.error}`);
        process.exit(1);
      }

      // Append to file
      const absoluteFilePath = join(absoluteRepoPath, filePath);
      await appendFile(absoluteFilePath, content || '', 'utf8');
      console.log(`✓ Appended to file: ${filePath}`);
      console.log(`  Location: ${absoluteFilePath}`);
    } catch (error) {
      console.error(`✗ Error appending to file:`, error instanceof Error ? error.message : error);
      process.exit(1);
    }
  });

export { fileCommand };

