import { Command } from 'commander';
import inquirer from 'inquirer';

export const interactiveCommand = new Command('interactive')
  .alias('i')
  .description('Start interactive mode')
  .action(async () => {
    console.log('🚀 GitStore CLI - Interactive Mode\n');

    while (true) {
      const { action } = await inquirer.prompt([
        {
          type: 'list',
          name: 'action',
          message: 'What would you like to do?',
          choices: [
            { name: '📁 File Operations', value: 'file' },
            { name: '🔧 Git Operations', value: 'git' },
            { name: '❌ Exit', value: 'exit' },
          ],
        },
      ]);

      if (action === 'exit') {
        console.log('👋 Goodbye!');
        break;
      }

      if (action === 'file') {
        const { fileAction } = await inquirer.prompt([
          {
            type: 'list',
            name: 'fileAction',
            message: 'File operation:',
            choices: [
              { name: 'Create file', value: 'create' },
              { name: 'Write to file', value: 'write' },
              { name: 'Append to file', value: 'append' },
              { name: '← Back', value: 'back' },
            ],
          },
        ]);

        if (fileAction === 'back') continue;

        const { repo, path, content } = await inquirer.prompt([
          {
            type: 'input',
            name: 'repo',
            message: 'Repository path:',
            default: process.cwd(),
          },
          {
            type: 'input',
            name: 'path',
            message: 'File path (relative to repo):',
            validate: (input) => input.length > 0 || 'File path is required',
          },
          {
            type: fileAction === 'append' ? 'input' : 'editor',
            name: 'content',
            message: fileAction === 'append' ? 'Content to append:' : 'File content:',
            default: '',
          },
        ]);

        // Import and execute file command
        const { fileCommand } = await import('./file.js');
        const subCommand = fileCommand.commands.find(cmd => cmd.name() === fileAction);
        if (subCommand) {
          await subCommand.parseAsync(['--repo', repo, '--path', path, '--content', content]);
        }
      }

      if (action === 'git') {
        const { gitAction } = await inquirer.prompt([
          {
            type: 'list',
            name: 'gitAction',
            message: 'Git operation:',
            choices: [
              { name: 'Status', value: 'status' },
              { name: 'Add files', value: 'add' },
              { name: 'Commit', value: 'commit' },
              { name: 'Push', value: 'push' },
              { name: 'Checkout branch', value: 'checkout' },
              { name: 'List branches', value: 'branch' },
              { name: 'View log', value: 'log' },
              { name: '← Back', value: 'back' },
            ],
          },
        ]);

        if (gitAction === 'back') continue;

        const { repo } = await inquirer.prompt([
          {
            type: 'input',
            name: 'repo',
            message: 'Repository path:',
            default: process.cwd(),
          },
        ]);

        // Import and execute git command
        const { gitCommand } = await import('./git.js');
        const subCommand = gitCommand.commands.find(cmd => cmd.name() === gitAction);
        
        if (subCommand) {
          const args = ['--repo', repo];
          
          if (gitAction === 'add') {
            const { path } = await inquirer.prompt([
              {
                type: 'input',
                name: 'path',
                message: 'File or directory path:',
                default: '.',
              },
            ]);
            args.push('--path', path);
          } else if (gitAction === 'commit') {
            const { message } = await inquirer.prompt([
              {
                type: 'input',
                name: 'message',
                message: 'Commit message:',
                validate: (input) => input.length > 0 || 'Commit message is required',
              },
            ]);
            args.push('--message', message);
          } else if (gitAction === 'checkout') {
            const { branch, create } = await inquirer.prompt([
              {
                type: 'input',
                name: 'branch',
                message: 'Branch name:',
                validate: (input) => input.length > 0 || 'Branch name is required',
              },
              {
                type: 'confirm',
                name: 'create',
                message: 'Create branch if it does not exist?',
                default: false,
              },
            ]);
            args.push('--branch', branch);
            if (create) args.push('--create-branch');
          } else if (gitAction === 'log') {
            const { number } = await inquirer.prompt([
              {
                type: 'input',
                name: 'number',
                message: 'Number of commits to show:',
                default: '10',
              },
            ]);
            args.push('--number', number);
          }
          
          await subCommand.parseAsync(args);
        }
      }
    }
  });

