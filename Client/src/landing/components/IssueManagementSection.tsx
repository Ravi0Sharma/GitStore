
import { CircleDot, CheckCircle2, Clock, Tag, User, MessageSquare } from "lucide-react";

const IssueManagementSection = () => {
  const mockIssues = [
    {
      id: 1,
      title: "Implement user authentication",
      status: "open",
      priority: "high",
      labels: ["feature", "security"],
      assignee: "JD",
      comments: 3,
    },
    {
      id: 2,
      title: "Fix navigation bug on mobile",
      status: "in-progress",
      priority: "medium",
      labels: ["bug", "mobile"],
      assignee: "AK",
      comments: 5,
    },
    {
      id: 3,
      title: "Update documentation",
      status: "closed",
      priority: "low",
      labels: ["docs"],
      assignee: "MS",
      comments: 2,
    },
  ];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "open":
        return <CircleDot className="w-4 h-4 text-primary" />;
      case "in-progress":
        return <Clock className="w-4 h-4 text-primary" />;
      case "closed":
        return <CheckCircle2 className="w-4 h-4 text-green-500" />;
      default:
        return <CircleDot className="w-4 h-4 text-muted-foreground" />;
    }
  };

  const getLabelColor = (label: string) => {
    const labelLower = label.toLowerCase();
    if (labelLower === 'feature') {
      return 'bg-blue-100 text-blue-700 border-blue-200';
    } else if (labelLower === 'security') {
      return 'bg-red-100 text-red-700 border-red-200';
    } else if (labelLower === 'bug') {
      return 'bg-orange-100 text-orange-700 border-orange-200';
    } else if (labelLower === 'mobile') {
      return 'bg-purple-100 text-purple-700 border-purple-200';
    } else if (labelLower === 'docs') {
      return 'bg-green-100 text-green-700 border-green-200';
    } else {
      return 'bg-muted/50 text-card-foreground/75 border-border';
    }
  };

  const capitalizeFirst = (str: string) => {
    return str.charAt(0).toUpperCase() + str.slice(1);
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "high":
        return "bg-primary/20 text-primary border-primary/30";
      case "medium":
        return "bg-primary/15 text-primary/90 border-primary/25";
      case "low":
        return "bg-muted text-muted-foreground border-border";
      default:
        return "bg-muted text-muted-foreground border-border";
    }
  };

  return (
    <section className="relative py-24">
      <div className="container mx-auto px-6">
        <div className="max-w-4xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/20 border border-primary/40 mb-6 shadow-sm">
              <Tag className="w-4 h-4 text-primary" />
              <span className="text-primary text-sm font-semibold">Issue Management</span>
            </div>
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              <span className="text-foreground">Manage issues with </span>
              <span className="text-primary">clarity</span>
            </h2>
            <p className="text-card-foreground/80 text-lg max-w-2xl mx-auto">
              Track progress, assign team members, and organize work with labels and priorities. 
              Everything you need to stay on top of your project.
            </p>
          </div>

          {/* Issue List Card */}
          <div className="bg-card rounded-2xl overflow-hidden card-glow">
            {/* Header */}
            <div className="px-6 py-4 border-b border-border/30 flex items-center justify-between">
              <div className="flex items-center gap-4">
                <span className="text-card-foreground font-semibold text-sm">All Issues</span>
                <span className="text-card-foreground/60 text-xs">3 total</span>
              </div>
              <div className="flex items-center gap-2">
                <button className="px-3 py-1.5 text-xs rounded-md bg-muted/40 text-card-foreground/80 hover:bg-muted/60 transition-colors font-medium">
                  Filter
                </button>
                <button className="px-3 py-1.5 text-xs rounded-md bg-muted/40 text-card-foreground/80 hover:bg-muted/60 transition-colors font-medium">
                  Sort
                </button>
              </div>
            </div>

            {/* Issue Items */}
            <div className="divide-y divide-border/20">
              {mockIssues.map((issue) => (
                <div
                  key={issue.id}
                  className="px-6 py-4 hover:bg-muted/5 transition-colors cursor-pointer"
                >
                  <div className="flex items-start gap-3">
                    {/* Status Icon */}
                    <div className="mt-0.5">{getStatusIcon(issue.status)}</div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-card-foreground font-medium text-sm truncate">
                          {issue.title}
                        </span>
                        <span className={`px-2 py-0.5 text-xs rounded border ${getPriorityColor(issue.priority)}`}>
                          {issue.priority}
                        </span>
                      </div>
                      
                      {/* Labels */}
                      <div className="flex items-center gap-2 mb-2">
                        {issue.labels.map((label) => (
                          <span
                            key={label}
                            className={`px-2 py-0.5 text-xs rounded-full border font-medium ${getLabelColor(label)}`}
                          >
                            {capitalizeFirst(label)}
                          </span>
                        ))}
                      </div>

                      {/* Meta */}
                      <div className="flex items-center gap-4 text-xs text-card-foreground/65">
                        <div className="flex items-center gap-1">
                          <User className="w-3 h-3" />
                          <span>{issue.assignee}</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <MessageSquare className="w-3 h-3" />
                          <span>{issue.comments}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default IssueManagementSection;
