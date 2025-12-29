import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import { GitMerge, ArrowLeft, ArrowRight, CheckCircle2, XCircle, GitBranch } from 'lucide-react';
import { MergeResult } from '../types/git';
import { routes } from '../routes';

const MergePage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, mergeBranches, loadRepositories, loading } = useGit();
  const navigate = useNavigate();
  const [isRefetching, setIsRefetching] = useState(false);
  const repo = getRepository(repoId || '');

  const [fromBranch, setFromBranch] = useState('');
  const [toBranch, setToBranch] = useState('');
  const [mergeResult, setMergeResult] = useState<MergeResult | null>(null);
  const [isMerging, setIsMerging] = useState(false);

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

  if (!repoId) {
    return null; // Will redirect
  }

  if (!repo && (loading || isRefetching)) {
    return (
      <div className="min-h-screen bg-background">
        <main className="container mx-auto px-4 py-8">
          <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
            <p className="text-muted-foreground">Loading repository...</p>
          </div>
        </main>
      </div>
    );
  }

  if (!repo) {
    return (
      <div className="min-h-screen bg-background">
        <main className="container mx-auto px-4 py-8">
          <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
            <p className="text-muted-foreground">Repository not found. Please select a repository from the dashboard.</p>
            <Link to={routes.dashboard} className="text-accent hover:text-accent/80 mt-4 inline-block">
              ← Back to dashboard
            </Link>
          </div>
        </main>
      </div>
    );
  }

  const handleMerge = async () => {
    if (fromBranch && toBranch && fromBranch !== toBranch) {
      setIsMerging(true);
      setMergeResult(null); // Clear previous result
      try {
        const result = await mergeBranches(repo.id, fromBranch, toBranch);
        setMergeResult(result);
        
        // If merge was successful, refresh the repo data to show updated branches/commits
        if (result.success) {
          // The mergeBranches function already calls loadRepositories, but we can also
          // explicitly refresh the current repo view
          await loadRepositories();
        }
      } catch (err) {
        // This should not happen as mergeBranches catches errors, but just in case
        setMergeResult({
          success: false,
          message: err instanceof Error ? err.message : 'An unexpected error occurred'
        });
      } finally {
        setIsMerging(false);
      }
    }
  };

  const canMerge = fromBranch && toBranch && fromBranch !== toBranch;

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <button
          onClick={() => navigate(routes.dashboardRepo(repo.id))}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to {repo.name}
        </button>

        <div className="mb-6">
          <h1 className="text-2xl font-bold text-foreground flex items-center gap-3">
            <GitMerge className="h-6 w-6 text-muted-foreground" />
            Merge Branches
          </h1>
          <p className="text-muted-foreground mt-2">
            Merge branches in <span className="font-mono text-primary">{repo.name}</span>
          </p>
        </div>

        {mergeResult && (
          <div
            className={`github-card p-4 mb-6 flex items-start gap-3 animate-fade-in ${
              mergeResult.success ? 'bg-success/10 border-success/30' : 'bg-destructive/10 border-destructive/30'
            }`}
          >
            {mergeResult.success ? (
              <CheckCircle2 className="h-5 w-5 text-success flex-shrink-0 mt-0.5" />
            ) : (
              <XCircle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5" />
            )}
            <div>
              <p className={mergeResult.success ? 'text-success font-medium' : 'text-destructive font-medium'}>
                {mergeResult.success ? 'Merge successful!' : 'Merge failed'}
              </p>
              <p className={`text-sm mt-1 ${mergeResult.success ? 'text-success/80' : 'text-destructive/80'}`}>
                {mergeResult.message}
              </p>
              {mergeResult.success && mergeResult.type === 'fast-forward' && (
                <div className="mt-3 p-3 bg-card rounded-md border border-border">
                  <p className="text-xs text-muted-foreground">Merge type</p>
                  <p className="font-mono text-sm text-foreground mt-1">Fast-forward merge completed</p>
                </div>
              )}
              {!mergeResult.success && (
                <div className="mt-3 p-3 bg-card rounded-md border border-border">
                  <p className="text-xs text-muted-foreground">Note</p>
                  <p className="text-sm text-foreground mt-1">
                    No changes were made to the repository. You can try a different merge or navigate away.
                  </p>
                </div>
              )}
            </div>
          </div>
        )}

        <div className="github-card">
          <div className="p-4 border-b border-border">
            <h2 className="font-semibold text-foreground">Compare & Merge</h2>
            <p className="text-sm text-muted-foreground mt-1">
              Select branches to merge
            </p>
          </div>

          <div className="p-6">
            <div className="flex flex-col md:flex-row items-center gap-4">
              <div className="flex-1 w-full">
                <label className="block text-sm font-medium text-muted-foreground mb-2">
                  From branch (source)
                </label>
                <select
                  value={fromBranch}
                  onChange={(e) => setFromBranch(e.target.value)}
                  className="github-input w-full font-mono"
                >
                  <option value="">Select branch...</option>
                  {repo.branches.map((branch) => (
                    <option key={branch.name} value={branch.name}>
                      {branch.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex items-center justify-center w-12 h-12 rounded-full bg-secondary border border-border">
                <ArrowRight className="h-5 w-5 text-muted-foreground" />
              </div>

              <div className="flex-1 w-full">
                <label className="block text-sm font-medium text-muted-foreground mb-2">
                  Into branch (target)
                </label>
                <select
                  value={toBranch}
                  onChange={(e) => setToBranch(e.target.value)}
                  className="github-input w-full font-mono"
                >
                  <option value="">Select branch...</option>
                  {repo.branches.map((branch) => (
                    <option key={branch.name} value={branch.name}>
                      {branch.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {fromBranch && toBranch && fromBranch === toBranch && (
              <p className="text-sm text-destructive mt-4">
                Cannot merge a branch into itself
              </p>
            )}

            <div className="mt-6 pt-6 border-t border-border">
              <button
                onClick={handleMerge}
                disabled={!canMerge || isMerging}
                className={`flex items-center gap-2 ${
                  canMerge && !isMerging ? 'github-btn-primary' : 'bg-muted text-muted-foreground cursor-not-allowed px-4 py-2 rounded-md'
                }`}
              >
                <GitMerge className="h-4 w-4" />
                {isMerging ? 'Merging...' : 'Merge branches'}
              </button>
            </div>
          </div>
        </div>

        <div className="github-card mt-6">
          <div className="p-4 border-b border-border">
            <h2 className="font-semibold text-foreground">Available Branches</h2>
          </div>
          <div className="divide-y divide-border">
            {repo.branches.map((branch, index) => (
              <div key={`${repoId}-${branch.name}-${index}`} className="p-4 flex items-center gap-3">
                <GitBranch className={`h-4 w-4 ${branch.name === repo.currentBranch ? 'text-success' : 'text-muted-foreground'}`} />
                <span className="font-mono text-sm text-foreground">{branch.name}</span>
                {branch.name === repo.currentBranch && (
                  <span className="bg-success/20 text-success text-xs px-2 py-0.5 rounded-full">
                    default
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
};

export default MergePage;
