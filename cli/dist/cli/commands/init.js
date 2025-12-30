import { Command } from 'commander';
import { getRepoPath, detectRepoType } from '../utils/repo-validator.js';
import { isApiAvailable } from '../utils/gitstore-api.js';
import simpleGit from 'simple-git';
import { existsSync } from 'fs';
export const initCommand = new Command('init')
    .description('Initialize a new git repository (standard git only)')
    .option('-r, --repo <path>', 'Repository path')
    .option('--gitstore', 'Initialize as GitStore repository (requires API server)')
    .action(async (options) => {
    try {
        const repoPath = options.repo || process.cwd();
        const absolutePath = getRepoPath(repoPath);
        // Check if already a git repo
        const repoType = detectRepoType(absolutePath);
        if (repoType === 'git' || repoType === 'gitstore') {
            const typeName = repoType === 'gitstore' ? 'GitStore' : 'git';
            console.log(`✓ Already a ${typeName} repository: ${absolutePath}`);
            return;
        }
        // Check if directory exists
        if (!existsSync(absolutePath)) {
            console.error(`✗ Error: Directory does not exist: ${absolutePath}`);
            process.exit(1);
        }
        // GitStore initialization
        if (options.gitstore) {
            const apiAvailable = await isApiAvailable();
            if (!apiAvailable) {
                console.error('✗ GitStore API server is not available.');
                console.error('  Please start the GitStore server (gitserver) first.');
                console.error('  Or use standard git init without --gitstore flag.');
                process.exit(1);
            }
            console.log('ℹ️  GitStore repositories should be created via the API:');
            console.log('   POST /api/repos with { name, description }');
            console.log('   Or use the web interface at http://localhost:5173');
            console.log('   Standard git repositories can be initialized with: gitstore init');
            process.exit(0);
        }
        // Initialize standard git repository
        const git = simpleGit(absolutePath);
        await git.init();
        console.log(`✓ Initialized git repository in: ${absolutePath}`);
        console.log(`  Type: standard git (.git)`);
    }
    catch (error) {
        console.error(`✗ Error initializing repository:`, error instanceof Error ? error.message : error);
        process.exit(1);
    }
});
//# sourceMappingURL=init.js.map