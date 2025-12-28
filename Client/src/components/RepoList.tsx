import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { BookMarked, GitBranch, Plus, Star } from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';
import { routes } from '../routes';
import { useGit } from '../context/GitContext';

const RepoList = () => {
  const [showNewRepoForm, setShowNewRepoForm] = useState(false);
  const [newRepoName, setNewRepoName] = useState('');
  const [newRepoDesc, setNewRepoDesc] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const { repositories, createRepository, loading, apiStatus, apiError } = useGit();

  // Debug: Log when repositories change in RepoList
  useEffect(() => {
    console.log('RepoList: Repositories changed', repositories.length, repositories.map(r => r.id));
  }, [repositories]);

  const handleCreateRepo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newRepoName.trim() && !isCreating) {
      setIsCreating(true);
      try {
        console.log('Creating repository:', newRepoName.trim(), newRepoDesc.trim() || undefined);
        await createRepository(newRepoName.trim(), newRepoDesc.trim() || undefined);
        console.log('Repository created successfully');
        setNewRepoName('');
        setNewRepoDesc('');
        setShowNewRepoForm(false);
      } catch (err) {
        console.error('Failed to create repository:', err);
        const errorMessage = err instanceof Error ? err.message : 'Unknown error';
        alert(`Failed to create repository: ${errorMessage}`);
      } finally {
        setIsCreating(false);
      }
    }
  };

  return (
    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <div className="p-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between bg-white dark:bg-gray-900 rounded-t-lg">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
          <BookMarked className="h-5 w-5" />
          Repositories
        </h2>
        <button
          onClick={() => setShowNewRepoForm(!showNewRepoForm)}
          className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-md text-sm font-medium flex items-center gap-2 transition-colors"
        >
          <Plus className="h-4 w-4" />
          New
        </button>
      </div>

      {showNewRepoForm && (
        <form onSubmit={handleCreateRepo} className="p-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Repository name
              </label>
              <input
                type="text"
                value={newRepoName}
                onChange={(e) => setNewRepoName(e.target.value)}
                placeholder="my-awesome-project"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                autoFocus
                disabled={isCreating}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Description <span className="text-gray-500 dark:text-gray-400">(optional)</span>
              </label>
              <textarea
                value={newRepoDesc}
                onChange={(e) => setNewRepoDesc(e.target.value)}
                placeholder="A short description of your repository"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                rows={2}
                disabled={isCreating}
              />
            </div>
            <div className="flex gap-2">
              <button 
                type="submit" 
                disabled={isCreating || !newRepoName.trim()}
                className="bg-green-600 hover:bg-green-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                {isCreating ? 'Creating...' : 'Create repository'}
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowNewRepoForm(false);
                  setNewRepoName('');
                  setNewRepoDesc('');
                }}
                disabled={isCreating}
                className="bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 px-4 py-2 rounded-md text-sm font-medium transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
            </div>
          </div>
        </form>
      )}

      {/* API Status Banner */}
      {apiStatus === 'disconnected' && (
        <div className="p-4 bg-yellow-50 dark:bg-yellow-900/20 border-b border-yellow-200 dark:border-yellow-800">
          <div className="flex items-start gap-3">
            <div className="flex-1">
              <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                ⚠️ Backend API not available
              </p>
              <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                {apiError || 'Cannot connect to backend server. Repositories are stored locally only.'}
              </p>
              <p className="text-xs text-yellow-600 dark:text-yellow-400 mt-2">
                To enable full functionality, start the backend server:
              </p>
              <code className="text-xs bg-yellow-100 dark:bg-yellow-900/40 px-2 py-1 rounded mt-1 block">
                cd gitClone && go build ./cmd/gitserver && ./gitserver
              </code>
            </div>
          </div>
        </div>
      )}
      {apiStatus === 'error' && apiError && !apiError.includes('Failed to reach API') && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
          <p className="text-sm font-medium text-red-800 dark:text-red-200">
            API Error: {apiError}
          </p>
        </div>
      )}

      <div className="divide-y divide-gray-200 dark:divide-gray-700">
        {loading && (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400">
            <p>Loading repositories...</p>
          </div>
        )}
        {!loading && repositories.length > 0 && repositories.map((repo) => {
          // Use branchCount from backend if available, otherwise use branches.length
          const branchCount = repo.branchCount ?? repo.branches?.length ?? 0;
          // Use lastUpdated from backend for "updated" display
          const lastUpdated = repo.lastUpdated;
          
          return (
            <Link
              key={repo.id}
              to={routes.dashboardRepo(repo.id)}
              className="block p-4 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors bg-white dark:bg-gray-900"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="text-base font-semibold text-blue-600 dark:text-blue-400 hover:underline">
                    {repo.name || repo.id}
                  </h3>
                  {repo.description && (
                    <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{repo.description}</p>
                  )}
                  <div className="flex items-center gap-4 mt-3 text-xs text-gray-500 dark:text-gray-400">
                    <span className="flex items-center gap-1">
                      <GitBranch className="h-3.5 w-3.5" />
                      {branchCount} {branchCount === 1 ? 'branch' : 'branches'}
                    </span>
                    {lastUpdated && (
                      <span>
                        Updated {safeDistanceToNow(lastUpdated)}
                      </span>
                    )}
                  </div>
                </div>
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    // Star functionality placeholder
                  }}
                  className="text-gray-400 hover:text-yellow-500 dark:hover:text-yellow-400 transition-colors p-1"
                >
                  <Star className="h-4 w-4" />
                </button>
              </div>
            </Link>
          );
        })}
        {!loading && repositories.length === 0 && (
          <div className="p-8 text-center text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-900">
            <BookMarked className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>No repositories yet. Create your first one!</p>
            {apiStatus === 'disconnected' && (
              <p className="text-xs mt-2 text-yellow-600 dark:text-yellow-400">
                Note: Repositories created without backend will only be stored locally.
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default RepoList;
