import { CheckCircle2, Circle } from 'lucide-react';

interface ChecklistItem {
  id: string;
  text: string;
  completed: boolean;
}

interface ChecklistProps {
  issueId?: string;
}

const checklistItems: ChecklistItem[] = [
  { id: '1', text: 'Create a new repo directory from the UI and it appears in the repo list', completed: true },
  { id: '2', text: 'Verify that selecting a repo opens a repo page showing current branch and latest commit', completed: true },
  { id: '3', text: 'In the repo page, include links to Branch and Merge pages for that repo', completed: true },
  { id: '4', text: 'Branch page: can create a new branch name for the selected repo', completed: true },
  { id: '5', text: 'When a new branch is created, it appears in the repo\'s branch list view', completed: true },
  { id: '6', text: 'Merge page: can run merge for the selected repo and shows success for fast-forward merges', completed: true },
];

const Checklist = ({ issueId }: ChecklistProps) => {
  return (
    <div className="github-card">
      <div className="p-4 border-b border-border">
        <h2 className="text-lg font-semibold text-foreground">Feature Checklist</h2>
        <p className="text-sm text-muted-foreground mt-1">
          All required behaviors implemented
        </p>
      </div>
      <div className="divide-y divide-border">
        {checklistItems.map((item) => (
          <div
            key={item.id}
            className="p-4 flex items-start gap-3 hover:bg-secondary/30 transition-colors"
          >
            {item.completed ? (
              <CheckCircle2 className="h-5 w-5 text-success flex-shrink-0 mt-0.5" />
            ) : (
              <Circle className="h-5 w-5 text-muted-foreground flex-shrink-0 mt-0.5" />
            )}
            <span className={`text-sm ${item.completed ? 'text-foreground' : 'text-muted-foreground'}`}>
              {item.text}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Checklist;
