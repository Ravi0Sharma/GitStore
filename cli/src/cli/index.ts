#!/usr/bin/env node

import { Command } from 'commander';
import { initCommand } from './commands/init.js';
import { fileCommand } from './commands/file.js';
import { gitCommand } from './commands/git.js';
import { interactiveCommand } from './commands/interactive.js';

const program = new Command();

program
  .name('gitstore')
  .description('CLI tool for GitStore - file operations and git commands')
  .version('1.0.0');

// Add commands
program.addCommand(initCommand);
program.addCommand(fileCommand);
program.addCommand(gitCommand);
program.addCommand(interactiveCommand);

// Parse arguments
program.parse();

