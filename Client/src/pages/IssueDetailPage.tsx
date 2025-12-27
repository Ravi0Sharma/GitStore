import { useParams, Link, useNavigate } from 'react-router-dom';
import { useGit } from '../context/GitContext';
import MarkdownRenderer from '../components/MarkdownRenderer';
import CommitHistory from '../components/CommitHistory';
import { ArrowLeft, CircleDot, CheckCircle2, Clock } from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';

const priorityColors = {
  low: 'bg-muted text-muted-foreground',
  medium: 'bg-primary/20 text-primary',
  high: 'bg-destructive/20 text-destructive',
};

const IssueDetailPage = () => {
  const { repoId, issueId } = useParams<{ repoId: string; issueId: string }>();
  const { getRepository, getIssue, updateIssueBody, toggleIssueStatus } = useGit();
  const navigate = useNavigate();

  const repo = getRepository(repoId || '');
  const issue = getIssue(repoId || '', issueId || '');

  if (!repoId || !issueId) {
    return (
      <div className="min-h-screen bg-background">
        <main className="container mx-auto px-4 py-8">
          <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
            <p className="text-muted-foreground">Repository or issue ID missing</p>
            <Link to="/dashboard" className="text-accent hover:text-accent/80 mt-4 inline-block">
              ← Back to dashboard
            </Link>
          </div>
        </main>
      </div>
    );
  }

  if (!repo || !issue) {
    return (
      <div className="min-h-screen bg-background">
        <main className="container mx-auto px-4 py-8">
          <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 text-center">
            <p className="text-muted-foreground">Issue not found. Please select a repository and issue from the dashboard.</p>
            <Link to={repoId ? `/repo/${repoId}/issues` : '/dashboard'} className="text-accent hover:text-accent/80 mt-4 inline-block">
              ← Back to {repoId ? 'issues' : 'dashboard'}
            </Link>
          </div>
        </main>
      </div>
    );
  }

  const handleStatusToggle = () => {
    toggleIssueStatus(repo.id, issue.id);
  };

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <button
          onClick={() => navigate(`/repo/${repo.id}/issues`)}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to issues
        </button>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Issue header */}
            <div className="github-card p-6">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1">
                  <h1 className="text-2xl font-bold text-foreground mb-3">
                    {issue.title}
                  </h1>
                  
                  <div className="flex flex-wrap items-center gap-3">
                    <span
                      className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium ${
                        issue.status === 'open'
                          ? 'bg-success/20 text-success'
                          : 'bg-purple-500/20 text-purple-400'
                      }`}
                    >
                      {issue.status === 'open' ? (
                        <CircleDot className="h-4 w-4" />
                      ) : (
                        <CheckCircle2 className="h-4 w-4" />
                      )}
                      {issue.status === 'open' ? 'Open' : 'Closed'}
                    </span>
                    
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${priorityColors[issue.priority]}`}>
                      {issue.priority} priority
                    </span>
                    
                    {issue.labels.map((label) => (
                      <span
                        key={label.id}
                        className="text-xs px-2 py-0.5 rounded-full font-medium"
                        style={{ 
                          backgroundColor: `${label.color}20`,
                          color: label.color,
                        }}
                      >
                        {label.name}
                      </span>
                    ))}
                  </div>
                </div>
                
                <button
                  onClick={handleStatusToggle}
                  className={issue.status === 'open' ? 'github-btn-secondary' : 'github-btn-primary'}
                >
                  {issue.status === 'open' ? 'Close issue' : 'Reopen issue'}
                </button>
              </div>
              
              <div className="flex items-center gap-4 mt-4 pt-4 border-t border-border text-sm text-muted-foreground">
                <div className="flex items-center gap-2">
                  <img
                    src={issue.authorAvatar}
                    alt={issue.author}
                    className="w-5 h-5 rounded-full"
                  />
                  <span>{issue.author}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Clock className="h-4 w-4" />
                  opened {safeDistanceToNow(issue.createdAt)}
                </div>
              </div>
            </div>

            {/* Issue body */}
            <div className="github-card p-6">
              <MarkdownRenderer content={issue.body} />
            </div>

            {/* Commit history */}
            <CommitHistory commits={[]} />
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1 space-y-6">
            {/* Details card */}
            <div className="github-card">
              <div className="p-4 border-b border-border">
                <h3 className="text-sm font-semibold text-foreground">Details</h3>
              </div>
              <div className="divide-y divide-border">
                <div className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">Assignee</p>
                  <div className="flex items-center gap-2">
                    <img
                      src={issue.authorAvatar}
                      alt={issue.author}
                      className="w-5 h-5 rounded-full"
                    />
                    <span className="text-sm text-foreground">{issue.author}</span>
                  </div>
                </div>
                <div className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">Labels</p>
                  <div className="flex flex-wrap gap-1">
                    {issue.labels.length === 0 ? (
                      <span className="text-sm text-muted-foreground">None</span>
                    ) : (
                      issue.labels.map((label) => (
                        <span
                          key={label.id}
                          className="text-xs px-2 py-0.5 rounded-full font-medium"
                          style={{ 
                            backgroundColor: `${label.color}20`,
                            color: label.color,
                          }}
                        >
                          {label.name}
                        </span>
                      ))
                    )}
                  </div>
                </div>
                <div className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">Priority</p>
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${priorityColors[issue.priority]}`}>
                    {issue.priority}
                  </span>
                </div>
                <div className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">Created</p>
                  <span className="text-sm text-foreground">
                    {safeDistanceToNow(issue.createdAt)}
                  </span>
                </div>
              </div>
            </div>

            {/* Commits card */}
            <div className="github-card">
              <div className="p-4 border-b border-border">
                <h3 className="text-sm font-semibold text-foreground">Linked Commits</h3>
              </div>
              <div className="p-4">
                <p className="text-sm text-muted-foreground">
                  No commits linked
                </p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default IssueDetailPage;