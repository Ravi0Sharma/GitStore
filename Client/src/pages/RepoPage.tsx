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
  CircleDot
} from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';
import { routes } from '../routes';


const RepoPage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, switchBranch, loadRepositories, loading } = useGit();
  const navigate = useNavigate();
  const [isRefetching, setIsRefetching] = useState(false);
  
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

  const latestCommit = repo.commits && repo.commits.length > 0 ? repo.commits[repo.commits.length - 1] : null;

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
          <h1 className="text-2xl font-bold text-foreground flex items-center gap-3">
            <BookMarked className="h-6 w-6 text-muted-foreground" />
            {repo.name}
          </h1>
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
            {repo.branches.map((branch) => (
              <div
                key={branch.name}
                className={`p-4 flex items-center justify-between hover:bg-secondary/60 transition-colors cursor-pointer ${
                  branch.name === repo.currentBranch ? 'bg-secondary/70' : ''
                }`}
                onClick={() => switchBranch(repo.id, branch.name)}
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
            {latestCommit ? (
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
            ) : (
              <p className="text-muted-foreground">No commits yet</p>
            )}
          </div>
        </div>
      </main>
  );
};

export default RepoPage;
