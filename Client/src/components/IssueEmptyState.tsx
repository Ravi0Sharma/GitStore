import { CircleDot, Lightbulb } from 'lucide-react';

interface IssueEmptyStateProps {
  onCreateClick: () => void;
}

const IssueEmptyState = ({ onCreateClick }: IssueEmptyStateProps) => {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-8 text-center">
      <div className="w-24 h-24 rounded-full bg-secondary flex items-center justify-center mb-6">
        <CircleDot className="h-12 w-12 text-muted-foreground" />
      </div>
      
      <h2 className="text-2xl font-semibold text-foreground mb-2">
        No issues yet
      </h2>
      
      <p className="text-muted-foreground mb-8 max-w-md">
        Issues are used to track bugs, enhancements, and other requests. Get started by creating your first issue.
      </p>
      
      <button
        onClick={onCreateClick}
        className="github-btn-warning text-lg px-6 py-3 flex items-center gap-2"
      >
        + Create your first issue
      </button>
      
      <div className="mt-8 bg-secondary/50 border border-border rounded-lg p-4 max-w-md">
        <div className="flex items-start gap-3 text-left">
          <Lightbulb className="h-5 w-5 text-muted-foreground mt-0.5 flex-shrink-0" />
          <div>
            <p className="text-sm font-medium text-muted-foreground">Pro tip</p>
            <p className="text-sm text-muted-foreground mt-1">
              You can use Markdown checklists in your issue descriptions to track progress. Just use <code className="bg-muted px-1 py-0.5 rounded text-xs font-mono">- [ ]</code> for unchecked and <code className="bg-muted px-1 py-0.5 rounded text-xs font-mono">- [x]</code> for checked items.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default IssueEmptyState;