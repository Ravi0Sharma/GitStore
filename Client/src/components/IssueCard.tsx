import { Issue } from '../types/git';
import { CircleDot, CheckCircle2, MessageSquare } from 'lucide-react';
import { safeDistanceToNow } from '../utils/dateHelpers';
import { Link } from 'react-router-dom';
import { routes } from '../routes';

interface IssueCardProps {
  issue: Issue;
  repoId: string;
}

const priorityColors = {
  low: 'bg-muted text-muted-foreground',
  medium: 'bg-primary/20 text-primary',
  high: 'bg-destructive/20 text-destructive',
};

const IssueCard = ({ issue, repoId }: IssueCardProps) => {
  return (
    <Link
      to={routes.dashboardRepoIssueDetail(repoId, issue.id)}
      className="flex items-start gap-4 p-4 border-b border-border hover:bg-secondary/30 transition-colors"
    >
      <div className="mt-1">
        {issue.status === 'open' ? (
          <CircleDot className="h-5 w-5 text-success" />
        ) : (
          <CheckCircle2 className="h-5 w-5 text-purple-500" />
        )}
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="text-foreground font-medium hover:text-primary transition-colors">
            {issue.title}
          </h3>
          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${priorityColors[issue.priority]}`}>
            {issue.priority}
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
        
        <p className="text-sm text-muted-foreground mt-1">
          #{issue.id.slice(-4)} opened {safeDistanceToNow(issue.createdAt)} by {issue.author}
        </p>
      </div>
      
      <div className="flex items-center gap-4 text-muted-foreground">
        <div className="flex items-center gap-1">
          <img
            src={issue.authorAvatar}
            alt={issue.author}
            className="w-5 h-5 rounded-full"
          />
        </div>
        {issue.commentCount > 0 && (
          <div className="flex items-center gap-1 text-sm">
            <MessageSquare className="h-4 w-4" />
            {issue.commentCount}
          </div>
        )}
      </div>
    </Link>
  );
};

export default IssueCard;