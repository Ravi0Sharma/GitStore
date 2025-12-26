import { useState } from 'react';
import { Link } from 'react-router-dom';
import { BookMarked, GitBranch, Plus, Star } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

const RepoList = () => {
  const [showNewRepoForm, setShowNewRepoForm] = useState(false);
  const [newRepoName, setNewRepoName] = useState('');
  const [newRepoDesc, setNewRepoDesc] = useState('');

const repositories = [
  { id: "3", name: "repo-3", description: "", branch: "sfs", createdAt: "dcda", branches: "dsfsdf" },
];
  

  const handleCreateRepo = (e: React.FormEvent) => {
    e.preventDefault();
    if (newRepoName.trim()) {
      //createRepository(newRepoName.trim(), newRepoDesc.trim());
      setNewRepoName('');
      setNewRepoDesc('');
      setShowNewRepoForm(false);
    }
  };

  return (
    <div className="github-card">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <h2 className="text-lg font-semibold text-foreground flex items-center gap-2">
          <BookMarked className="h-5 w-5" />
          Repositories
        </h2>
        <button
          onClick={() => setShowNewRepoForm(!showNewRepoForm)}
          className="github-btn-primary flex items-center gap-2 text-sm"
        >
          <Plus className="h-4 w-4" />
          New
        </button>
      </div>

      {showNewRepoForm && (
        <form onSubmit={handleCreateRepo} className="p-4 border-b border-border bg-secondary/50 animate-fade-in">
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Repository name
              </label>
              <input
                type="text"
                value={newRepoName}
                onChange={(e) => setNewRepoName(e.target.value)}
                placeholder="my-awesome-project"
                className="github-input w-full"
                autoFocus
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Description <span className="text-muted-foreground">(optional)</span>
              </label>
              <input
                type="text"
                value={newRepoDesc}
                onChange={(e) => setNewRepoDesc(e.target.value)}
                placeholder="A short description of your repository"
                className="github-input w-full"
              />
            </div>
            <div className="flex gap-2">
              <button type="submit" className="github-btn-primary text-sm">
                Create repository
              </button>
              <button
                type="button"
                onClick={() => setShowNewRepoForm(false)}
                className="github-btn-secondary text-sm"
              >
                Cancel
              </button>
            </div>
          </div>
        </form>
      )}

      <div className="divide-y divide-border">
        {repositories.map((repo) => (
          <Link
            key={repo.id}
            to={`/repo/${repo.id}`}
            className="block p-4 hover:bg-secondary/50 transition-colors animate-slide-in"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <h3 className="text-base font-semibold github-link flex items-center gap-2">
                  <BookMarked className="h-4 w-4 text-muted-foreground" />
                  {repo.name}
                </h3>
                {repo.description && (
                  <p className="text-sm text-muted-foreground mt-1">{repo.description}</p>
                )}
                <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <GitBranch className="h-3.5 w-3.5" />
                    {repo.branches.length} branches
                  </span>
                  <span>
                    Updated {formatDistanceToNow(repo.createdAt, { addSuffix: true })}
                  </span>
                </div>
              </div>
              <button
                onClick={(e) => {
                  e.preventDefault();
                  // Star functionality placeholder
                }}
                className="text-muted-foreground hover:text-foreground transition-colors p-1"
              >
                <Star className="h-4 w-4" />
              </button>
            </div>
          </Link>
        ))}
        {repositories.length === 0 && (
          <div className="p-8 text-center text-muted-foreground">
            <BookMarked className="h-12 w-12 mx-auto mb-3 opacity-50" />
            <p>No repositories yet. Create your first one!</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default RepoList;
