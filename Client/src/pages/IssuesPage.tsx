import { useState, useMemo } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import IssueCard from '../components/IssueCard';
import IssueFilters from '../components/IssueFilters';
import IssueEmptyState from '../components/IssueEmptyState';
import CreateIssueModal from '../components/CreateIssueModal';
import Checklist from '../components/Checklist';
import { ArrowLeft, Plus, CircleDot, CheckCircle2 } from 'lucide-react';
import { IssueStatus, Label, Priority } from '../types/git';
import { routes } from '../routes';

type SortOption = 'priority' | 'newest' | 'oldest';

const priorityOrder = { high: 0, medium: 1, low: 2 };

const IssuesPage = () => {
  const { repoId } = useParams<{ repoId: string }>();
  const { getRepository, createIssue } = useGit();
  const navigate = useNavigate();
  const repo = getRepository(repoId || '');

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [sortBy, setSortBy] = useState<SortOption>('newest');
  const [statusFilter, setStatusFilter] = useState<IssueStatus | 'all'>('all');
  const [labelFilter, setLabelFilter] = useState<string | null>(null);

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

  const filteredAndSortedIssues = useMemo(() => {
    let issues = [...repo.issues];

    // Filter by status
    if (statusFilter !== 'all') {
      issues = issues.filter((issue) => issue.status === statusFilter);
    }

    // Filter by label
    if (labelFilter) {
      issues = issues.filter((issue) =>
        issue.labels.some((label) => label.id === labelFilter)
      );
    }

    // Sort
    switch (sortBy) {
      case 'priority':
        issues.sort((a, b) => priorityOrder[a.priority] - priorityOrder[b.priority]);
        break;
      case 'newest':
        issues.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        break;
      case 'oldest':
        issues.sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());
        break;
    }

    return issues;
  }, [repo.issues, sortBy, statusFilter, labelFilter]);

  const openCount = repo.issues.filter((i) => i.status === 'open').length;
  const closedCount = repo.issues.filter((i) => i.status === 'closed').length;

  const handleCreateIssue = (title: string, body: string, priority: Priority, labels: Label[]) => {
    createIssue(repo.id, title, body, priority, labels);
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

        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-foreground">Issues</h1>
          {repo.issues.length > 0 && (
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="github-btn-warning flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              Create New Issue
            </button>
          )}
        </div>

        {repo.issues.length === 0 ? (
          <div className="github-card">
            <IssueEmptyState onCreateClick={() => setIsCreateModalOpen(true)} />
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2">
              <div className="github-card">
                <div className="p-4 border-b border-border flex items-center gap-4">
                  <button
                    onClick={() => setStatusFilter('open')}
                    className={`flex items-center gap-2 text-sm ${
                      statusFilter === 'open' ? 'text-foreground font-medium' : 'text-muted-foreground'
                    }`}
                  >
                    <CircleDot className="h-4 w-4" />
                    {openCount} Open
                  </button>
                  <button
                    onClick={() => setStatusFilter('closed')}
                    className={`flex items-center gap-2 text-sm ${
                      statusFilter === 'closed' ? 'text-foreground font-medium' : 'text-muted-foreground'
                    }`}
                  >
                    <CheckCircle2 className="h-4 w-4" />
                    {closedCount} Closed
                  </button>
                </div>

                <div className="p-4 border-b border-border">
                  <IssueFilters
                    sortBy={sortBy}
                    onSortChange={setSortBy}
                    statusFilter={statusFilter}
                    onStatusFilterChange={setStatusFilter}
                    labelFilter={labelFilter}
                    onLabelFilterChange={setLabelFilter}
                  />
                </div>

                <div className="divide-y divide-border">
                  {filteredAndSortedIssues.length === 0 ? (
                    <div className="p-8 text-center text-muted-foreground">
                      No issues match your filters.
                    </div>
                  ) : (
                    filteredAndSortedIssues.map((issue) => (
                      <IssueCard key={issue.id} issue={issue} repoId={repo.id} />
                    ))
                  )}
                </div>
              </div>
            </div>

            <div className="lg:col-span-1">
              <Checklist issueId={`repo-${repo.id}`} />
            </div>
          </div>
        )}
      </main>

      <CreateIssueModal
        open={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateIssue}
      />
    </div>
  );
};

export default IssuesPage;