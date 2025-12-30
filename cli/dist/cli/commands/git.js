import { Command } from 'commander';
import { validateGitRepo, getRepoPath } from '../utils/repo-validator.js';
import { handleGitOperation } from '../utils/git-handler.js';
import inquirer from 'inquirer';
const gitCommand = new Command('git')
    .description('Git operations (status, add, commit, push, checkout, branch, log)');
// Git status command
gitCommand
    .command('status')
    .description('Show git status')
    .option('-r, --repo <path>', 'Repository path')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        if (options.interactive || !repoPath) {
            const answers = await inquirer.prompt([
                {
                    type: 'input',
                    name: 'repo',
                    message: 'Repository path:',
                    default: repoPath || process.cwd(),
                    when: !repoPath,
                },
            ]);
            repoPath = repoPath || answers.repo;
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        // Use git handler for both git and GitStore repos
        const status = await handleGitOperation(absoluteRepoPath, async (git) => git.status(), 'status');
        console.log('\n📊 Git Status:');
        console.log(`Current branch: ${status.current}`);
        console.log(`\nChanges to be committed:`);
        if (status.staged.length > 0) {
            status.staged.forEach((file) => console.log(`  ✓ ${file}`));
        }
        else {
            console.log('  (none)');
        }
        console.log(`\nChanges not staged for commit:`);
        if (status.not_added.length > 0) {
            status.not_added.forEach((file) => console.log(`  ✗ ${file}`));
        }
        else {
            console.log('  (none)');
        }
        console.log(`\nUntracked files:`);
        const untrackedFiles = status.files?.filter((f) => f.working_dir === '??').map((f) => f.path) || [];
        if (untrackedFiles.length > 0) {
            untrackedFiles.forEach((file) => console.log(`  ? ${file}`));
        }
        else {
            console.log('  (none)');
        }
    }
    catch (error) {
        console.error(`✗ Error getting git status:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git add command
gitCommand
    .command('add')
    .description('Add files to staging area')
    .option('-r, --repo <path>', 'Repository path')
    .option('-p, --path <path>', 'File or directory path (default: .)', '.')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        let filePath = options.path || '.';
        if (options.interactive || !repoPath) {
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
                    message: 'File or directory path:',
                    default: filePath,
                    when: options.interactive,
                },
            ]);
            repoPath = repoPath || answers.repo;
            filePath = answers.path || filePath;
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        await handleGitOperation(absoluteRepoPath, async (git) => {
            await git.add(filePath);
            console.log(`✓ Added ${filePath} to staging area`);
        }, 'add');
    }
    catch (error) {
        console.error(`✗ Error adding files:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git commit command
gitCommand
    .command('commit')
    .description('Create a commit')
    .option('-r, --repo <path>', 'Repository path')
    .option('-m, --message <message>', 'Commit message')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        let message = options.message;
        if (options.interactive || !repoPath || !message) {
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
                    name: 'message',
                    message: 'Commit message:',
                    when: !message,
                    validate: (input) => input.length > 0 || 'Commit message is required',
                },
            ]);
            repoPath = repoPath || answers.repo;
            message = message || answers.message;
        }
        if (!message) {
            console.error('✗ Commit message is required');
            process.exit(1);
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        const result = await handleGitOperation(absoluteRepoPath, async (git) => git.commit(message), 'commit');
        console.log(`✓ Commit created: ${result.commit}`);
        console.log(`  Message: ${message}`);
    }
    catch (error) {
        console.error(`✗ Error creating commit:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git push command
gitCommand
    .command('push')
    .description('Push commits to remote')
    .option('-r, --repo <path>', 'Repository path')
    .option('--remote <remote>', 'Remote name (default: origin)', 'origin')
    .option('--branch <branch>', 'Branch name (default: current branch)')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        if (options.interactive || !repoPath) {
            const answers = await inquirer.prompt([
                {
                    type: 'input',
                    name: 'repo',
                    message: 'Repository path:',
                    default: repoPath || process.cwd(),
                    when: !repoPath,
                },
            ]);
            repoPath = repoPath || answers.repo;
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        const remote = options.remote || 'origin';
        const branch = options.branch;
        await handleGitOperation(absoluteRepoPath, async (git) => {
            if (branch) {
                await git.push(remote, branch);
                console.log(`✓ Pushed to ${remote}/${branch}`);
            }
            else {
                const status = await git.status();
                const currentBranch = status.current || 'main';
                await git.push(remote, currentBranch);
                console.log(`✓ Pushed to ${remote}/${currentBranch}`);
            }
        }, 'push');
    }
    catch (error) {
        console.error(`✗ Error pushing:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git checkout command
gitCommand
    .command('checkout')
    .description('Checkout a branch or create new branch')
    .option('-r, --repo <path>', 'Repository path')
    .option('-b, --branch <branch>', 'Branch name')
    .option('--create-branch', 'Create new branch if it does not exist')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        let branch = options.branch;
        if (options.interactive || !repoPath || !branch) {
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
                    name: 'branch',
                    message: 'Branch name:',
                    when: !branch,
                    validate: (input) => input.length > 0 || 'Branch name is required',
                },
            ]);
            repoPath = repoPath || answers.repo;
            branch = branch || answers.branch;
        }
        if (!branch) {
            console.error('✗ Branch name is required');
            process.exit(1);
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        await handleGitOperation(absoluteRepoPath, async (git) => {
            if (options.createBranch) {
                await git.checkoutLocalBranch(branch);
                console.log(`✓ Created and checked out branch: ${branch}`);
            }
            else {
                await git.checkout(branch);
                console.log(`✓ Checked out branch: ${branch}`);
            }
        }, 'checkout');
    }
    catch (error) {
        console.error(`✗ Error checking out branch:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git branch command
gitCommand
    .command('branch')
    .description('List or create branches')
    .option('-r, --repo <path>', 'Repository path')
    .option('-c, --create <branch>', 'Create a new branch')
    .option('-d, --delete <branch>', 'Delete a branch')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        if (options.interactive || !repoPath) {
            const answers = await inquirer.prompt([
                {
                    type: 'input',
                    name: 'repo',
                    message: 'Repository path:',
                    default: repoPath || process.cwd(),
                    when: !repoPath,
                },
            ]);
            repoPath = repoPath || answers.repo;
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        await handleGitOperation(absoluteRepoPath, async (git) => {
            if (options.create) {
                await git.checkoutLocalBranch(options.create);
                console.log(`✓ Created branch: ${options.create}`);
            }
            else if (options.delete) {
                await git.deleteLocalBranch(options.delete);
                console.log(`✓ Deleted branch: ${options.delete}`);
            }
            else {
                const branches = await git.branchLocal();
                console.log('\n branches:');
                branches.all.forEach(branch => {
                    const prefix = branch === branches.current ? '* ' : '  ';
                    console.log(`${prefix}${branch}`);
                });
            }
        }, 'branch');
    }
    catch (error) {
        console.error(`✗ Error managing branches:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
// Git log command
gitCommand
    .command('log')
    .description('Show commit log')
    .option('-r, --repo <path>', 'Repository path')
    .option('-n, --number <number>', 'Number of commits to show', '10')
    .option('-i, --interactive', 'Use interactive mode')
    .action(async (options) => {
    try {
        let repoPath = options.repo;
        if (options.interactive || !repoPath) {
            const answers = await inquirer.prompt([
                {
                    type: 'input',
                    name: 'repo',
                    message: 'Repository path:',
                    default: repoPath || process.cwd(),
                    when: !repoPath,
                },
            ]);
            repoPath = repoPath || answers.repo;
        }
        const absoluteRepoPath = getRepoPath(repoPath);
        const validation = validateGitRepo(absoluteRepoPath);
        if (!validation.valid) {
            console.error(`✗ ${validation.error}`);
            process.exit(1);
        }
        const number = parseInt(options.number || '10', 10);
        const log = await handleGitOperation(absoluteRepoPath, async (git) => git.log({ maxCount: number }), 'log');
        console.log(`\n📜 Commit Log (${log.total} commits, showing ${log.all.length}):\n`);
        log.all.forEach((commit, index) => {
            console.log(`[${index + 1}] ${commit.hash.substring(0, 7)} - ${commit.message}`);
            console.log(`    Author: ${commit.author_name} <${commit.author_email}>`);
            console.log(`    Date: ${commit.date}`);
            console.log('');
        });
    }
    catch (error) {
        console.error(`✗ Error getting commit log:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
export { gitCommand };
//# sourceMappingURL=git.js.map