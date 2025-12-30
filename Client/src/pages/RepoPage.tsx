import { useParams, Link, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import { useEffect, useState } from 'react';
import { 
  BookMarked, 
  GitBranch, 
  GitMerge, 
  ChevronDown, 
  Code, 
  Clock,
  ArrowLeft,
  CircleDot,
  Terminal
} from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';
import { routes } from '../routes';
import { api, Commit } from '../lib/api';


const RepoPage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, switchBranch, loadRepositories, loading } = useGit();
  const navigate = useNavigate();
  const [isRefetching, setIsRefetching] = useState(false);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [loadingCommits, setLoadingCommits] = useState(false);
  
  // Redirect if no repoId
  useEffect(() => {
    if (!repoId) {
      navigate(routes.dashboard);
    }
  }, [repoId, navigate]);

  // Try to refetch if repo not found
  useEffect(() => {
    if (repoId && !getRepository(repoId) && !loading && !isRefetching) {
      setIsRefetching(true);
      loadRepositories().finally(() => setIsRefetching(false));
    }
  }, [repoId, loading, isRefetching]);

  const repo = getRepository(repoId || '');

  // Load commits for current branch when repo or branch changes
  useEffect(() => {
    if (repoId && repo?.currentBranch) {
      const requestId = Date.now();
      setLoadingCommits(true);
      console.log(`[RepoPage] requestId=${requestId}, Loading commits for branch: ${repo.currentBranch}`);
      api.getCommits(repoId, repo.currentBranch, 10)
        .then(loadedCommits => {
          console.log(`[RepoPage] requestId=${requestId}, Loaded ${loadedCommits.length} commits for ${repo.currentBranch}:`, 
            loadedCommits.map(c => `${c.hash}: ${c.message}`));
          setCommits(loadedCommits); // REPLACE commits (not append)
        })
        .catch(err => {
          console.error(`[RepoPage] requestId=${requestId}, Failed to load commits:`, err);
          setCommits([]);
        })
        .finally(() => {
          setLoadingCommits(false);
        });
    }
  }, [repoId, repo?.currentBranch]);

  if (!repoId) {
    return null; // Will redirect
  }

  if (!repo && (loading || isRefetching)) {
    return (
      <main className="container mx-auto px-4 py-8">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
          <p className="text-muted-foreground">Loading repository...</p>
        </div>
      </main>
    );
  }

  if (!repo) {
    return (
      <main className="container mx-auto px-4 py-8">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
          <p className="text-muted-foreground">Repository not found. Please select a repository from the dashboard.</p>
          <Link to={routes.dashboard} className="text-accent hover:text-accent/80 mt-4 inline-block">
            ← Back to dashboard
          </Link>
        </div>
      </main>
    );
  }

  // Get the latest commit from loaded commits (already sorted by backend, first is latest)
  const latestCommit = commits && commits.length > 0 ? commits[0] : null;

  return (
    <main className="container mx-auto px-4 py-8">
        <button
          onClick={() => navigate(routes.dashboard)}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to dashboard
        </button>

        <div className="mb-6">
          <div className="flex items-center justify-between mb-2">
            <h1 className="text-2xl font-bold text-foreground flex items-center gap-3">
              <BookMarked className="h-6 w-6 text-muted-foreground" />
              {repo.name}
            </h1>
            <Link
              to={routes.dashboardRepoCLI(repo.id)}
              className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-lg transition-colors font-medium"
            >
              <Terminal className="h-4 w-4" />
              Enter CLI
            </Link>
          </div>
          {repo.description && (
            <p className="text-muted-foreground mt-2">{repo.description}</p>
          )}
        </div>

        <div className="flex flex-wrap gap-3 mb-6">
          <Link
            to={routes.dashboardRepoIssues(repo.id)}
            className="github-btn-secondary flex items-center gap-2"
          >
            <CircleDot className="h-4 w-4" />
            Issues
            <span className="bg-muted px-2 py-0.5 rounded-full text-xs">
              {repo.issues?.length || 0}
            </span>
          </Link>
          <Link
            to={routes.dashboardRepoBranches(repo.id)}
            className="github-btn-secondary flex items-center gap-2"
          >
            <GitBranch className="h-4 w-4" />
            Branches
            <span className="bg-muted px-2 py-0.5 rounded-full text-xs">
              {repo.branches.length}
            </span>
          </Link>
          <Link
            to={routes.dashboardRepoMerge(repo.id)}
            className="github-btn-secondary flex items-center gap-2"
          >
            <GitMerge className="h-4 w-4" />
            Merge
          </Link>
        </div>

        <div className="bg-secondary/50 rounded-lg border border-border shadow-md mb-6">
          <div className="p-4 border-b border-border flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="relative">
                <button className="github-btn-secondary flex items-center gap-2 text-sm">
                  <GitBranch className="h-4 w-4" />
                  {repo.currentBranch}
                  <ChevronDown className="h-4 w-4" />
                </button>
              </div>
              <span className="text-sm text-muted-foreground">
                {repo.branches.length} branches
              </span>
            </div>
          </div>
          
          <div className="divide-y divide-border">
            {repo.branches.map((branch, index) => (
              <div
                key={`${repo.id}-${branch.name}-${index}`}
                className={`p-4 flex items-center justify-between hover:bg-secondary/60 transition-colors cursor-pointer ${
                  branch.name === repo.currentBranch ? 'bg-secondary/70' : ''
                }`}
                onClick={async () => {
                  await switchBranch(repo.id, branch.name);
                  // Always reload commits when clicking a branch (even if same branch)
                  // This ensures we show the latest commits for that branch
                  setLoadingCommits(true);
                  try {
                    const loadedCommits = await api.getCommits(repo.id, branch.name, 10);
                    setCommits(loadedCommits);
                  } catch (err) {
                    console.error('Failed to reload commits:', err);
                    setCommits([]);
                  } finally {
                    setLoadingCommits(false);
                  }
                  // Commits will also be reloaded via useEffect when currentBranch changes
                }}
              >
                <div className="flex items-center gap-3">
                  <GitBranch className={`h-4 w-4 ${branch.name === repo.currentBranch ? 'text-success' : 'text-muted-foreground'}`} />
                  <span className={`font-mono text-sm ${branch.name === repo.currentBranch ? 'text-foreground font-medium' : 'text-foreground'}`}>
                    {branch.name}
                  </span>
                  {branch.name === repo.currentBranch && (
                    <span className="bg-success/20 text-success text-xs px-2 py-0.5 rounded-full">
                      current
                    </span>
                  )}
                </div>
                <span className="text-xs text-muted-foreground">
                  {safeDistanceToNow(branch.createdAt)}
                </span>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-2xl border border-border bg-secondary/50 shadow-md">
          <div className="p-4 border-b border-border/50">
            <h2 className="text-lg font-semibold text-foreground flex items-center gap-2">
              <Clock className="h-5 w-5 text-muted-foreground" />
              Latest Commit
            </h2>
          </div>
          <div className="p-4">
            {loadingCommits ? (
              <p className="text-muted-foreground">Loading commits...</p>
            ) : latestCommit ? (
              <div className="space-y-4">
                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-full bg-secondary flex items-center justify-center">
                    <Code className="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div className="flex-1">
                    <p className="text-foreground font-medium">{latestCommit.message}</p>
                    <div className="flex items-center gap-3 mt-2 text-sm text-muted-foreground">
                      <span className="font-mono text-primary">{latestCommit.hash}</span>
                      <span>•</span>
                      <span>{latestCommit.author}</span>
                      <span>•</span>
                      <span>{safeDistanceToNow(latestCommit.date)}</span>
                    </div>
                  </div>
                </div>
                {commits.length > 1 && (
                  <div className="mt-4 pt-4 border-t border-border">
                    <p className="text-sm text-muted-foreground mb-2">Recent commits:</p>
                    <div className="space-y-2">
                      {commits.slice(1, 6).map((commit) => (
                        <div key={commit.hash} className="flex items-start gap-3 text-sm">
                          <span className="font-mono text-primary text-xs">{commit.hash.substring(0, 7)}</span>
                          <span className="text-foreground flex-1">{commit.message}</span>
                          <span className="text-muted-foreground text-xs">{safeDistanceToNow(commit.date)}</span>
                        </div>
                      ))}
                      {commits.length > 6 && (
                        <p className="text-xs text-muted-foreground">... and {commits.length - 6} more commits</p>
                      )}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground">No pushed commits yet</p>
            )}
          </div>
        </div>
      </main>
  );
};

export default RepoPage;
