import React from 'react';

interface MarkdownRendererProps {
  content: string;
  onChecklistChange?: (newContent: string) => void;
}

const MarkdownRenderer = ({ content, onChecklistChange }: MarkdownRendererProps) => {
  const handleCheckboxToggle = (lineIndex: number) => {
    if (!onChecklistChange) return;
    
    const lines = content.split('\n');
    const line = lines[lineIndex];
    
    if (line.includes('- [ ]')) {
      lines[lineIndex] = line.replace('- [ ]', '- [x]');
    } else if (line.includes('- [x]')) {
      lines[lineIndex] = line.replace('- [x]', '- [ ]');
    }
    
    onChecklistChange(lines.join('\n'));
  };

  const renderContent = () => {
    const lines = content.split('\n');
    const elements: React.ReactNode[] = [];
    let currentList: React.ReactNode[] = [];
    let inList = false;

    lines.forEach((line, index) => {
      // Headings
      if (line.startsWith('## ')) {
        if (inList && currentList.length > 0) {
          elements.push(<ul key={`list-${index}`} className="space-y-2 my-3">{currentList}</ul>);
          currentList = [];
          inList = false;
        }
        elements.push(
          <h2 key={index} className="text-lg font-semibold text-foreground mt-4 mb-2">
            {line.replace('## ', '')}
          </h2>
        );
        return;
      }

      if (line.startsWith('# ')) {
        if (inList && currentList.length > 0) {
          elements.push(<ul key={`list-${index}`} className="space-y-2 my-3">{currentList}</ul>);
          currentList = [];
          inList = false;
        }
        elements.push(
          <h1 key={index} className="text-xl font-bold text-foreground mt-4 mb-2">
            {line.replace('# ', '')}
          </h1>
        );
        return;
      }

      // Checklist items
      if (line.includes('- [ ]') || line.includes('- [x]')) {
        inList = true;
        const isChecked = line.includes('- [x]');
        const text = line.replace(/- \[[ x]\]\s*/, '');
        
        currentList.push(
          <li key={index} className="flex items-center gap-3">
            <button
              onClick={() => handleCheckboxToggle(index)}
              className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                isChecked 
                  ? 'bg-success border-success text-success-foreground' 
                  : 'border-muted-foreground hover:border-primary'
              }`}
            >
              {isChecked && (
                <svg className="w-3 h-3" viewBox="0 0 12 12" fill="none">
                  <path
                    d="M2 6L5 9L10 3"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              )}
            </button>
            <span className={isChecked ? 'text-muted-foreground line-through' : 'text-foreground'}>
              {text}
            </span>
          </li>
        );
        return;
      }

      // Regular list items
      if (line.startsWith('- ')) {
        inList = true;
        currentList.push(
          <li key={index} className="flex items-center gap-2 text-foreground">
            <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground" />
            {line.replace('- ', '')}
          </li>
        );
        return;
      }

      // Empty line or end of list
      if (line.trim() === '' || !line.startsWith('-')) {
        if (inList && currentList.length > 0) {
          elements.push(<ul key={`list-${index}`} className="space-y-2 my-3">{currentList}</ul>);
          currentList = [];
          inList = false;
        }
      }

      // Regular paragraph (skip empty lines)
      if (line.trim() !== '' && !line.startsWith('-')) {
        // Handle inline code
        const processedLine = line.replace(
          /`([^`]+)`/g,
          '<code class="bg-muted px-1.5 py-0.5 rounded text-sm font-mono">$1</code>'
        );
        
        elements.push(
          <p
            key={index}
            className="text-foreground my-2"
            dangerouslySetInnerHTML={{ __html: processedLine }}
          />
        );
      }
    });

    // Don't forget remaining list items
    if (currentList.length > 0) {
      elements.push(<ul key="list-final" className="space-y-2 my-3">{currentList}</ul>);
    }

    return elements;
  };

  return <div className="prose prose-invert max-w-none">{renderContent()}</div>;
};

export default MarkdownRenderer;