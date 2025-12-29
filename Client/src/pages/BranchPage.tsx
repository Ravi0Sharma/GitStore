import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import { api } from '../lib/api';
import { GitBranch, Plus, ArrowLeft, CheckCircle2, AlertCircle } from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';
import { routes } from '../routes';

const BranchPage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, createBranch, loadRepositories, loading } = useGit();
  const navigate = useNavigate();
  const [isRefetching, setIsRefetching] = useState(false);
  const repo = getRepository(repoId || '');

  const [newBranchName, setNewBranchName] = useState('');
  const [showNewBranchForm, setShowNewBranchForm] = useState(false);
  const [successMessage, setSuccessMessage] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [loadingBranches, setLoadingBranches] = useState(false);
  const [branches, setBranches] = useState<{ name: string; createdAt: Date }[]>([]);

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

  // Load branches from API when component mounts or repo changes
  useEffect(() => {
    if (repoId && repo) {
      loadBranches();
    }
  }, [repoId, repo]);

  const loadBranches = async () => {
    if (!repoId) return;
    setLoadingBranches(true);
    setErrorMessage('');
    try {
      const apiBranches = await api.getBranches(repoId);
      // Deduplicate branches by name
      const branchMap = new Map<string, { name: string; createdAt: Date }>();
      apiBranches.forEach(b => {
        if (!branchMap.has(b.name)) {
          branchMap.set(b.name, {
            name: b.name,
            createdAt: new Date(b.createdAt),
          });
        }
      });
      const branchDates = Array.from(branchMap.values());
      setBranches(branchDates);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to load branches';
      setErrorMessage(errorMsg);
      console.error('Failed to load branches:', err);
      // Fallback to repo.branches if available
      if (repo?.branches) {
        setBranches(repo.branches);
      }
    } finally {
      setLoadingBranches(false);
    }
  };

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

  const handleCreateBranch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newBranchName.trim()) {
      return;
    }

    const branchName = newBranchName.trim();
    const branchExists = repo.branches.some((b) => b.name === branchName);
    if (branchExists) {
      setSuccessMessage('Branch already exists');
      setTimeout(() => setSuccessMessage(''), 3000);
      return;
    }
    
    try {
      await createBranch(repo.id, branchName);
      setSuccessMessage(`Branch '${branchName}' created successfully!`);
      setNewBranchName('');
      setShowNewBranchForm(false);
      // Reload branches from API after a short delay
      setTimeout(async () => {
        await loadBranches();
      }, 500);
      setTimeout(() => setSuccessMessage(''), 3000);
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to create branch';
      setErrorMessage(`Error: ${errorMsg}`);
      setTimeout(() => setErrorMessage(''), 5000);
    }
  };

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
            <GitBranch className="h-6 w-6 text-muted-foreground" />
            Branches
          </h1>
          <p className="text-muted-foreground mt-2">
            Manage branches for <span className="font-mono text-primary">{repo.name}</span>
          </p>
        </div>

        {successMessage && (
          <div className="github-card bg-success/10 border-success/30 p-4 mb-6 flex items-center gap-3 animate-fade-in">
            <CheckCircle2 className="h-5 w-5 text-success" />
            <span className="text-success">{successMessage}</span>
          </div>
        )}

        <div className="github-card mb-6">
          <div className="p-4 border-b border-border flex items-center justify-between">
            <h2 className="font-semibold text-foreground">All Branches</h2>
            <button
              onClick={() => setShowNewBranchForm(!showNewBranchForm)}
              className="github-btn-primary flex items-center gap-2 text-sm"
            >
              <Plus className="h-4 w-4" />
              New branch
            </button>
          </div>

          {showNewBranchForm && (
            <form onSubmit={handleCreateBranch} className="p-4 border-b border-border bg-secondary/50 animate-fade-in">
              <div className="flex gap-3">
                <input
                  type="text"
                  value={newBranchName}
                  onChange={(e) => setNewBranchName(e.target.value)}
                  placeholder="feature/new-feature"
                  className="github-input flex-1 font-mono"
                  autoFocus
                />
                <button type="submit" className="github-btn-primary text-sm">
                  Create branch
                </button>
                <button
                  type="button"
                  onClick={() => setShowNewBranchForm(false)}
                  className="github-btn-secondary text-sm"
                >
                  Cancel
                </button>
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                New branch will be created from <span className="font-mono text-primary">{repo.currentBranch}</span>
              </p>
            </form>
          )}

          <div className="divide-y divide-border">
            {loadingBranches ? (
              <div className="p-8 text-center text-muted-foreground">
                Loading branches...
              </div>
            ) : branches.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">
                No branches found
              </div>
            ) : (
              branches.map((branch, index) => (
                <div
                  key={`${repoId}-${branch.name}-${index}`}
                  className="p-4 flex items-center justify-between hover:bg-secondary/30 transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <GitBranch className={`h-4 w-4 ${branch.name === repo?.currentBranch ? 'text-success' : 'text-muted-foreground'}`} />
                    <span className="font-mono text-sm text-foreground">{branch.name}</span>
                    {branch.name === repo?.currentBranch && (
                      <span className="bg-success/20 text-success text-xs px-2 py-0.5 rounded-full">
                        current
                      </span>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground">
                    Created {safeDistanceToNow(branch.createdAt)}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      </main>
    </div>
  );
};

export default BranchPage;
