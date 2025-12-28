import { useState } from 'react';
import { Label, Priority, DEFAULT_LABELS } from '../types/git';
import { X, ChevronDown } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
} from './dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './dropDown'

interface CreateIssueModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (title: string, body: string, priority: Priority, labels: Label[]) => void;
}

const CreateIssueModal = ({ open, onClose, onSubmit }: CreateIssueModalProps) => {
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [priority, setPriority] = useState<Priority>('medium');
  const [selectedLabels, setSelectedLabels] = useState<Label[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    onSubmit(title, body, priority, selectedLabels);
    setTitle('');
    setBody('');
    setPriority('medium');
    setSelectedLabels([]);
    onClose();
  };

  const toggleLabel = (label: Label) => {
    setSelectedLabels((prev) => {
      const exists = prev.find((l) => l.id === label.id);
      if (exists) {
        return prev.filter((l) => l.id !== label.id);
      }
      return [...prev, label];
    });
  };

  const priorityLabels = {
    low: 'Low',
    medium: 'Medium',
    high: 'High',
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm text-foreground max-w-2xl">
        <DialogHeader>
          <div className="text-foreground text-xl font-semibold">Create New Issue</div>
        </DialogHeader>
        
        <form onSubmit={handleSubmit} className="space-y-4 mt-4">
          <div>
            <input
              type="text"
              placeholder="Issue title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full text-lg px-4 py-2 bg-background/50 text-foreground border border-border/50 rounded-md placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary/40"
              required
            />
          </div>
          
          <div>
            <textarea
              placeholder="Add a description... (Markdown supported)&#10;&#10;You can use checklists:&#10;- [ ] Unchecked item&#10;- [x] Checked item"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="w-full min-h-[200px] font-mono text-sm px-4 py-2 bg-background/50 text-foreground border border-border/50 rounded-md placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary/40 resize-y"
            />
          </div>
          
          <div className="flex flex-wrap gap-3">
            {/* Priority selector */}
            <DropdownMenu>
              <DropdownMenuTrigger className="flex items-center gap-2 text-sm px-4 py-2 bg-background/50 hover:bg-background/70 text-foreground border border-border/50 rounded-md transition-colors">
                <span className={`w-2 h-2 rounded-full ${
                  priority === 'low' ? 'bg-blue-500' : 
                  priority === 'medium' ? 'bg-yellow-500' : 
                  'bg-red-500'
                }`}></span>
                Priority: {priorityLabels[priority]}
                <ChevronDown className="h-4 w-4" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="bg-secondary/30 backdrop-blur-sm text-foreground border-border/50 shadow-lg">
                <DropdownMenuItem onClick={() => setPriority('low')} className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-blue-500"></span>
                  Low
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setPriority('medium')} className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-yellow-500"></span>
                  Medium
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setPriority('high')} className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-full bg-red-500"></span>
                  High
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            {/* Labels multi-select */}
            <DropdownMenu>
              <DropdownMenuTrigger className="flex items-center gap-2 text-sm px-4 py-2 bg-background/50 hover:bg-background/70 text-foreground border border-border/50 rounded-md transition-colors">
                Labels ({selectedLabels.length})
                <ChevronDown className="h-4 w-4" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="bg-secondary/30 backdrop-blur-sm text-foreground border-border/50 shadow-lg">
                {DEFAULT_LABELS.map((label) => {
                  const isSelected = selectedLabels.some((l) => l.id === label.id);
                  return (
                    <DropdownMenuItem
                      key={label.id}
                      onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
                        e.preventDefault();
                        toggleLabel(label);
                      }}
                      className="flex items-center gap-2"
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        readOnly
                        className="w-4 h-4 accent-primary"
                      />
                      <span
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: label.color }}
                      />
                      {label.name}
                    </DropdownMenuItem>
                  );
                })}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Selected labels display */}
          {selectedLabels.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {selectedLabels.map((label) => {
                // Ensure label color has good contrast - if color is too light, use dark text
                const isLightColor = label.color && (
                  label.color.toLowerCase().includes('fff') ||
                  label.color.toLowerCase().includes('white') ||
                  label.color.toLowerCase().includes('#f') ||
                  label.color.toLowerCase().includes('rgb(255')
                );
                return (
                  <span
                    key={label.id}
                    className={`inline-flex items-center gap-1 text-xs px-2 py-1 rounded-full font-medium border border-border/30 ${
                      isLightColor ? 'text-foreground' : ''
                    }`}
                    style={{ 
                      backgroundColor: isLightColor 
                        ? `${label.color}30` 
                        : `${label.color}20`,
                      color: isLightColor 
                        ? 'hsl(var(--foreground))' // use theme foreground for light backgrounds
                        : label.color || 'hsl(var(--muted-foreground))', // use label color or fallback
                    }}
                  >
                    {label.name}
                    <button
                      type="button"
                      onClick={() => toggleLabel(label)}
                      className="hover:opacity-80 text-foreground/70 hover:text-foreground"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </span>
                );
              })}
            </div>
          )}
          
          <div className="flex justify-end gap-3 pt-4 border-t border-border/50">
            <button type="button" onClick={onClose} className="px-4 py-2 bg-background/50 hover:bg-background/70 text-foreground border border-border/50 rounded-md transition-colors">
              Cancel
            </button>
            <button type="submit" className="px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-md transition-colors font-medium shadow-sm">
              Create Issue
            </button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default CreateIssueModal;