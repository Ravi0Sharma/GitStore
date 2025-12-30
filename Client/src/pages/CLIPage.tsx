import { useParams, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import { useState, useRef, useEffect } from 'react';
import { ArrowLeft, Terminal } from 'lucide-react';
import { routes } from '../routes';
import { api } from '../lib/api';

interface CommandHistory {
  command: string;
  output: string;
  timestamp: Date;
}

const CLIPage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, loadRepositories } = useGit();
  const navigate = useNavigate();
  const [command, setCommand] = useState('');
  const [history, setHistory] = useState<CommandHistory[]>([]);
  const [isExecuting, setIsExecuting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const terminalRef = useRef<HTMLDivElement>(null);

  const repo = getRepository(repoId || '');

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Scroll to bottom when new output is added
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [history]);

  const executeCommand = async (cmd: string) => {
    if (!cmd.trim()) return;

    setIsExecuting(true);
    const commandLine = cmd.trim();
    
    // Add command to history immediately
    const newHistory: CommandHistory = {
      command: commandLine,
      output: '',
      timestamp: new Date(),
    };
    setHistory(prev => [...prev, newHistory]);

    try {
      // Parse command
      const parts = commandLine.split(' ').filter(p => p);
      const [mainCommand, ...args] = parts;

      let output = '';

      // Handle different commands
      switch (mainCommand) {
        case 'help':
          output = `Available commands:
  help                    - Show this help message
  status                  - Show repository status
  branch                  - List branches
  branches                - List branches
  commits                 - List commits
  log                     - Show commit log
  create file <path> [content] - Create a new file
  edit file <path> [content]   - Edit an existing file
  clear                   - Clear terminal
  exit                    - Exit CLI

Git commands (via API):
  git status              - Show repository status
  git branch              - List branches
  git log                 - Show commit log
  git add [path]          - Stage files (default: ".")
  git commit -m "message" - Commit changes (requires staged files)
  git push [remote] [branch] - Push to remote (default: origin, current branch)

Workflow:
  1. create file <path> [content]  - Create a file
  2. git add <path>                - Stage the file
  3. git commit -m "message"       - Commit (local only)
  4. git push                      - Push to remote (then visible in UI)

Examples:
  create file text.txt Hello World
  git add text.txt
  git commit -m "Add test file"
  git push

Repository: ${repo?.name || repoId}
Current branch: ${repo?.currentBranch || 'main'}
`;
          break;

        case 'clear':
          setHistory([]);
          setIsExecuting(false);
          return;

        case 'exit':
          navigate(routes.dashboardRepo(repoId || ''));
          return;

        case 'status':
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          try {
            const commits = await api.getCommits(repoId);
            output = `Repository: ${repo?.name || repoId}\n`;
            output += `Current branch: ${repo?.currentBranch || 'main'}\n`;
            output += `Commits: ${commits.length || 0}\n`;
            output += `Branches: ${repo?.branches.length || 0}\n`;
            if (repo?.issues) {
              output += `Issues: ${repo.issues.length || 0}\n`;
            }
          } catch (error) {
            output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch status'}`;
          }
          break;

        case 'git':
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          if (args[0] === 'status') {
            // Same as status command
            try {
              const commits = await api.getCommits(repoId);
              const branches = await api.getBranches(repoId);
              
              output = `Repository: ${repo?.name || repoId}\n`;
              output += `Current branch: ${repo?.currentBranch || 'main'}\n`;
              output += `Commits: ${commits.length}\n`;
              output += `Branches: ${branches.length}\n`;
            } catch (error) {
              output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch status'}`;
            }
          } else if (args[0] === 'branch' || args[0] === 'branches') {
            try {
              const branches = await api.getBranches(repoId!);
              output = `Branches:\n`;
              branches.forEach(branch => {
                const current = branch.name === repo?.currentBranch ? ' *' : '';
                output += `  ${branch.name}${current}\n`;
              });
            } catch (error) {
              output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch branches'}`;
            }
          } else if (args[0] === 'log') {
            try {
              const commits = await api.getCommits(repoId!);
              const limit = args[1] && !isNaN(parseInt(args[1])) ? parseInt(args[1]) : 10;
              output = `Commit Log:\n`;
              commits.slice(0, limit).forEach((commit, index) => {
                output += `  [${index + 1}] ${commit.hash.substring(0, 7)} - ${commit.message}\n`;
                output += `      ${commit.author} - ${commit.date}\n`;
              });
              if (commits.length > limit) {
                output += `\n  ... and ${commits.length - limit} more commits`;
              }
            } catch (error) {
              output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch commits'}`;
            }
          } else if (args[0] === 'add') {
            if (!repoId) {
              output = 'Error: Repository ID missing';
              break;
            }
            try {
              // Path is optional, default to "." (all changes)
              const path = args[1] || '.';
              
              // Debug: log command parsing
              console.log(`[CLI DEBUG] Parsed command: git add ${path}`);
              console.log(`[CLI DEBUG] Calling api.add(${repoId}, ${path}) -> POST /api/repos/${repoId}/add`);
              
              // Await add to ensure it completes before any subsequent commands
              const result = await api.add(repoId, path);
              
              // Debug: log result
              console.log(`[CLI DEBUG] Add result: stagedCount=${result.stagedCount}, stagedPaths=${result.stagedPaths.length}`);
              
              // Show actual staged info from API response
              if (result.stagedCount === 0) {
                output = `Error: No files were staged. Make sure the path exists and contains files.`;
              } else {
                output = `Staged ${result.stagedCount} file(s)\n`;
                // Show first few staged paths
                const pathsToShow = result.stagedPaths.slice(0, 5);
                pathsToShow.forEach(p => {
                  output += `  ${p}\n`;
                });
                if (result.stagedCount > 5) {
                  output += `  ... and ${result.stagedCount - 5} more\n`;
                }
              }
            } catch (error) {
              const errorMsg = error instanceof Error ? error.message : 'Failed to stage files';
              console.error(`[CLI DEBUG] Add error:`, error);
              output = `Error: ${errorMsg}\n`;
              if (errorMsg.includes('404') || errorMsg.includes('Not Found')) {
                output += `\nPossible issues:\n`;
                output += `- GitStore server might not be running (check http://localhost:8080)\n`;
                output += `- Repository "${repoId}" might not exist\n`;
                output += `- Make sure the backend server has been restarted with the latest changes\n`;
              } else if (errorMsg.includes('git add failed')) {
                output += `\nNote: This command requires a standard Git repository (not GitStore).\n`;
                output += `Make sure the repository is a standard Git repo with a .git directory.`;
              }
            }
          } else if (args[0] === 'commit' && args[1] === '-m' && args[2]) {
            if (!repoId) {
              output = 'Error: Repository ID missing';
              break;
            }
            try {
              const message = args.slice(2).join(' ');
              
              // Debug: log command parsing
              console.log(`[CLI DEBUG] Parsed command: git commit -m "${message}"`);
              console.log(`[CLI DEBUG] Calling api.commit(${repoId}, "${message}") -> POST /api/repos/${repoId}/commit`);
              
              // Await commit - this will only run after previous commands complete
              await api.commit(repoId, message);
              output = `Commit created successfully (local only)!\nMessage: ${message}`;
            } catch (error) {
              const errorMsg = error instanceof Error ? error.message : 'Failed to create commit';
              console.error(`[CLI DEBUG] Commit error:`, error);
              output = `Error: ${errorMsg}`;
              if (errorMsg.includes('Nothing to commit') || errorMsg.includes('Stage changes')) {
                output += `\n\nStage changes first with 'git add <path>' before committing.`;
              }
            }
          } else if (args[0] === 'push') {
            if (!repoId) {
              output = 'Error: Repository ID missing';
              break;
            }
            try {
              const remote = args[1] || 'origin';
              const branch = args[2] || repo?.currentBranch || 'main';
              await api.push(repoId, remote, branch);
              output = `Pushed to ${remote}/${branch} successfully!\nCommits are now visible in the repository.`;
              // Refresh repository data after push (commits will be updated)
              try {
                await loadRepositories();
              } catch (err) {
                // Ignore errors - commits will refresh when user navigates back to repo page
              }
            } catch (error) {
              output = `Error: ${error instanceof Error ? error.message : 'Failed to push'}`;
            }
          } else {
            output = `Unknown git command: ${args.join(' ')}\nUse 'help' for available commands.`;
          }
          break;

        case 'branch':
        case 'branches':
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          try {
            const branches = await api.getBranches(repoId);
            if (branches.length > 0) {
              output = `Branches:\n`;
              branches.forEach(branch => {
                const current = branch.name === repo?.currentBranch ? ' *' : '';
                output += `  ${branch.name}${current}\n`;
              });
            } else {
              output = 'No branches found';
            }
          } catch (error) {
            output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch branches'}`;
          }
          break;

        case 'commits':
        case 'log':
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          try {
            const commits = await api.getCommits(repoId);
            if (commits.length > 0) {
              output = `Commits:\n`;
              commits.slice(0, 10).forEach((commit, index) => {
                output += `  [${index + 1}] ${commit.hash.substring(0, 7)} - ${commit.message}\n`;
                output += `      ${commit.author} - ${commit.date}\n`;
              });
              if (commits.length > 10) {
                output += `\n  ... and ${commits.length - 10} more commits`;
              }
            } else {
              output = 'No commits found';
            }
          } catch (error) {
            output = `Error: ${error instanceof Error ? error.message : 'Failed to fetch commits'}`;
          }
          break;

        case 'create':
          if (args[0] !== 'file') {
            output = `Unknown command: create ${args[0] || '(nothing)'}\nUse 'create file <path> [content]' to create a file.\nExample: create file test.txt Hello World`;
            break;
          }
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          if (args.length < 2) {
            output = 'Error: File path is required\nUsage: create file <path> [content]\nExample: create file test.txt Hello World';
            break;
          }
          try {
            const filePath = args[1];
            // Join all remaining args as content, preserving spaces
            const content = args.slice(2).join(' ') || '';
            
            // Debug: log command parsing
            console.log(`[CLI DEBUG] Parsed command: create file ${filePath}`);
            console.log(`[CLI DEBUG] Calling api.createOrEditFile(${repoId}, ${filePath}, ...) -> POST /api/repos/${repoId}/files`);
            
            // Await file creation to ensure it completes before subsequent commands
            await api.createOrEditFile(repoId, filePath, content);
            output = `File created: ${filePath}\n`;
            if (content) {
              output += `Content: "${content}"\n`;
            } else {
              output += `(empty file)\n`;
            }
          } catch (error) {
            const errorMsg = error instanceof Error ? error.message : 'Failed to create file';
            console.error('Error creating file:', error);
            output = `Error: ${errorMsg}\n`;
            if (errorMsg.includes('404') || errorMsg.includes('Not Found') || errorMsg.includes('Failed to reach API')) {
              output += `\nPossible issues:\n`;
              output += `- GitStore server might not be running (check http://localhost:8080)\n`;
              output += `- Repository "${repoId}" might not exist\n`;
              output += `- Check browser console for more details\n`;
            } else if (errorMsg.includes('Network error') || errorMsg.includes('Failed to reach')) {
              output += `\nNetwork error: Make sure the GitStore server is running on http://localhost:8080\n`;
            } else {
              output += `\nMake sure the GitStore server is running on http://localhost:8080`;
            }
          }
          break;

        case 'edit':
          if (args[0] !== 'file') {
            output = `Unknown command: edit ${args[0] || '(nothing)'}\nUse 'edit file <path> [content]' to edit a file.\nExample: edit file test.txt Updated content`;
            break;
          }
          if (!repoId) {
            output = 'Error: Repository ID missing';
            break;
          }
          if (args.length < 2) {
            output = 'Error: File path is required\nUsage: edit file <path> [content]\nExample: edit file test.txt Updated content';
            break;
          }
          try {
            const filePath = args[1];
            // Join all remaining args as content, preserving spaces
            const content = args.slice(2).join(' ') || '';
            
            if (import.meta.env.DEV) {
              console.log(`Editing file: ${filePath} in repo: ${repoId}`);
            }
            
            await api.createOrEditFile(repoId, filePath, content);
            output = `File updated: ${filePath}\n`;
            if (content) {
              output += `Content: "${content}"`;
            } else {
              output += `(file cleared)`;
            }
          } catch (error) {
            const errorMsg = error instanceof Error ? error.message : 'Failed to edit file';
            console.error('Error editing file:', error);
            output = `Error: ${errorMsg}\n`;
            if (errorMsg.includes('404') || errorMsg.includes('Not Found') || errorMsg.includes('Failed to reach API')) {
              output += `\nPossible issues:\n`;
              output += `- GitStore server might not be running (check http://localhost:8080)\n`;
              output += `- Repository "${repoId}" might not exist\n`;
              output += `- Check browser console for more details\n`;
            } else if (errorMsg.includes('Network error') || errorMsg.includes('Failed to reach')) {
              output += `\nNetwork error: Make sure the GitStore server is running on http://localhost:8080\n`;
            } else {
              output += `\nMake sure the GitStore server is running on http://localhost:8080`;
            }
          }
          break;

        default:
          output = `Command not found: ${mainCommand}\nUse 'help' for available commands.`;
      }

      // Update history with output
      setHistory(prev => {
        const updated = [...prev];
        updated[updated.length - 1].output = output;
        return updated;
      });
    } catch (error) {
      const errorOutput = `Error: ${error instanceof Error ? error.message : 'Unknown error'}`;
      setHistory(prev => {
        const updated = [...prev];
        updated[updated.length - 1].output = errorOutput;
        return updated;
      });
    } finally {
      setIsExecuting(false);
      setCommand('');
      inputRef.current?.focus();
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isExecuting && command.trim()) {
      executeCommand(command);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      // TODO: Implement command history navigation
    }
  };

  if (!repoId) {
    return (
      <main className="container mx-auto px-4 py-8">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
          <p className="text-muted-foreground">Repository ID missing</p>
        </div>
      </main>
    );
  }

  if (!repo) {
    return (
      <main className="container mx-auto px-4 py-8">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
          <p className="text-muted-foreground">Repository not found</p>
        </div>
      </main>
    );
  }

  return (
    <main className="container mx-auto px-4 py-8">
      <button
        onClick={() => navigate(routes.dashboardRepo(repoId))}
        className="flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to repository
      </button>

      <div className="mb-4">
        <h1 className="text-2xl font-bold text-foreground flex items-center gap-3">
          <Terminal className="h-6 w-6 text-orange-500" />
          CLI - {repo.name}
        </h1>
        <p className="text-muted-foreground mt-2">Interactive command line interface</p>
      </div>

      <div className="bg-black rounded-lg border border-border shadow-lg overflow-hidden">
        {/* Terminal header */}
        <div className="bg-gray-800 px-4 py-2 flex items-center gap-2">
          <div className="flex gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500"></div>
            <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
            <div className="w-3 h-3 rounded-full bg-green-500"></div>
          </div>
          <span className="text-gray-400 text-sm ml-2">
            Terminal - {repo.name} | Branch: <span className="text-cyan-400 font-semibold">{repo.currentBranch}</span>
          </span>
        </div>

        {/* Terminal content */}
        <div
          ref={terminalRef}
          className="p-4 h-[600px] overflow-y-auto font-mono text-sm"
          style={{
            backgroundColor: '#1e1e1e',
            color: '#d4d4d4',
          }}
        >
          {/* Welcome message */}
          {history.length === 0 && (
            <div className="mb-4">
              <div className="text-green-400 font-bold">Welcome to GitStore CLI</div>
              <div className="text-gray-400 mt-2">
                Repository: <span className="text-white font-semibold">{repo.name}</span>
              </div>
              <div className="text-gray-400 mt-1">
                Current branch: <span className="text-cyan-400 font-semibold">{repo.currentBranch}</span>
              </div>
              <div className="text-gray-400 mt-3">
                Type 'help' for available commands or 'exit' to return.
              </div>
            </div>
          )}

          {/* Command history */}
          {history.map((item, index) => (
            <div key={index} className="mb-4">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-green-400">$</span>
                <span className="text-cyan-400 text-xs">[{repo.currentBranch}]</span>
                <span className="text-white">{item.command}</span>
              </div>
              {item.output && (
                <div className="ml-6 text-gray-300 whitespace-pre-wrap">{item.output}</div>
              )}
            </div>
          ))}

          {/* Current input line */}
          <form onSubmit={handleSubmit} className="flex items-center gap-2">
            <span className="text-green-400">$</span>
            <span className="text-cyan-400 text-xs">[{repo.currentBranch}]</span>
            <input
              ref={inputRef}
              type="text"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={isExecuting}
              className="flex-1 bg-transparent border-none outline-none text-white"
              placeholder={isExecuting ? 'Executing...' : 'Enter command...'}
              autoComplete="off"
            />
          </form>
        </div>
      </div>

      <div className="mt-4 text-sm text-muted-foreground">
        <p>Type 'help' to see available commands</p>
      </div>
    </main>
  );
};

export default CLIPage;

