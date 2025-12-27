import { Label, IssueStatus, DEFAULT_LABELS } from '../types/git';
import { ChevronDown } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from './dropDown'

type SortOption = 'priority' | 'newest' | 'oldest';

interface IssueFiltersProps {
  sortBy: SortOption;
  onSortChange: (sort: SortOption) => void;
  statusFilter: IssueStatus | 'all';
  onStatusFilterChange: (status: IssueStatus | 'all') => void;
  labelFilter: string | null;
  onLabelFilterChange: (labelId: string | null) => void;
}

const IssueFilters = ({
  sortBy,
  onSortChange,
  statusFilter,
  onStatusFilterChange,
  labelFilter,
  onLabelFilterChange,
}: IssueFiltersProps) => {
  const sortLabels = {
    priority: 'Priority',
    newest: 'Newest',
    oldest: 'Oldest',
  };

  const statusLabels = {
    all: 'All',
    open: 'Open',
    closed: 'Closed',
  };

  const selectedLabel = DEFAULT_LABELS.find((l) => l.id === labelFilter);

  return (
    <div className="flex flex-wrap items-center gap-3 mb-4">
      {/* Sort dropdown */}
      <DropdownMenu>
        <DropdownMenuTrigger className="github-btn-secondary flex items-center gap-2 text-sm">
          Sort: {sortLabels[sortBy]}
          <ChevronDown className="h-4 w-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="bg-card border-border">
          <DropdownMenuItem onClick={() => onSortChange('priority')}>
            Priority
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => onSortChange('newest')}>
            Newest
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => onSortChange('oldest')}>
            Oldest
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Status filter */}
      <DropdownMenu>
        <DropdownMenuTrigger className="github-btn-secondary flex items-center gap-2 text-sm">
          Status: {statusLabels[statusFilter]}
          <ChevronDown className="h-4 w-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="bg-card border-border">
          <DropdownMenuItem onClick={() => onStatusFilterChange('all')}>
            All
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => onStatusFilterChange('open')}>
            Open
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => onStatusFilterChange('closed')}>
            Closed
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Label filter */}
      <DropdownMenu>
        <DropdownMenuTrigger className="github-btn-secondary flex items-center gap-2 text-sm">
          Label: {selectedLabel?.name || 'All'}
          <ChevronDown className="h-4 w-4" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="bg-card border-border">
          <DropdownMenuItem onClick={() => onLabelFilterChange(null)}>
            All labels
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          {DEFAULT_LABELS.map((label) => (
            <DropdownMenuItem
              key={label.id}
              onClick={() => onLabelFilterChange(label.id)}
            >
              <span
                className="w-3 h-3 rounded-full mr-2"
                style={{ backgroundColor: label.color }}
              />
              {label.name}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default IssueFilters;