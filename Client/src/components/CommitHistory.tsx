import { Commit } from '../types/git';
import { GitCommit, User, Clock } from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';

interface CommitHistoryProps {
  commits: Commit[];
}

const CommitHistory = ({ commits }: CommitHistoryProps) => {
  if (commits.length === 0) {
    return (
      <div className="github-card p-6 text-center">
        <GitCommit className="h-8 w-8 text-muted-foreground mx-auto mb-3" />
        <p className="text-muted-foreground">No commits linked to this issue yet.</p>
      </div>
    );
  }

  return (
    <div className="github-card">
      <div className="p-4 border-b border-border">
        <h3 className="text-lg font-semibold text-foreground flex items-center gap-2">
          <GitCommit className="h-5 w-5 text-muted-foreground" />
          Commit History
          <span className="text-sm font-normal text-muted-foreground">
            ({commits.length} commit{commits.length !== 1 ? 's' : ''})
          </span>
        </h3>
      </div>
      
      <div className="divide-y divide-border">
        {commits.map((commit, index) => (
          <div
            key={commit.hash}
            className="p-4 flex items-start gap-4 hover:bg-secondary/30 transition-colors"
          >
            <div className="relative">
              <div className="w-8 h-8 rounded-full bg-secondary flex items-center justify-center">
                <GitCommit className="h-4 w-4 text-muted-foreground" />
              </div>
              {index < commits.length - 1 && (
                <div className="absolute top-8 left-1/2 w-0.5 h-full -translate-x-1/2 bg-border" />
              )}
            </div>
            
            <div className="flex-1 min-w-0">
              <p className="text-foreground font-medium truncate">
                {commit.message}
              </p>
              <div className="flex flex-wrap items-center gap-3 mt-1 text-sm text-muted-foreground">
                <span className="font-mono text-primary bg-primary/10 px-2 py-0.5 rounded">
                  {commit.hash}
                </span>
                <span className="flex items-center gap-1">
                  <User className="h-3 w-3" />
                  {commit.author}
                </span>
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {safeDistanceToNow(commit.date)}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CommitHistory;