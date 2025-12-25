import { FileText, Plus, CircleDot } from "lucide-react";

const IssuesSection = () => {
  return (
    <section className="relative py-24">
      <div className="container mx-auto px-6">
        <div className="max-w-4xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-accent/20 border border-accent/40 mb-3 shadow-sm">
              <CircleDot className="w-4 h-4 text-accent" />
              <span className="text-accent text-sm font-semibold">Issues</span>
            </div>
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              <span className="text-foreground">Organize your work with </span>
              <span className="text-accent">Issues</span>
            </h2>
            <p className="text-card-foreground/80 text-lg max-w-2xl mx-auto">
              Issues help you track tasks, bugs, and feature requests. They provide a clear way to 
              organize work and collaborate with your team.
            </p>
          </div>

          {/* Empty State Card */}
          <div className="bg-card rounded-2xl p-8 md:p-12 card-glow">
            <div className="flex flex-col items-center justify-center text-center">
              {/* Empty state illustration */}
              <div className="w-24 h-24 rounded-full bg-muted/40 border-2 border-dashed border-border/60 flex items-center justify-center mb-6">
                <FileText className="w-10 h-10 text-card-foreground/40" />
              </div>
              
              <h3 className="text-card-foreground font-semibold text-xl mb-2">
                No issues yet
              </h3>
              <p className="text-card-foreground/70 text-sm mb-6 max-w-md">
                Create your first issue to start tracking work. Issues can be assigned, 
                labeled, and linked to commits for complete traceability.
              </p>

              {/* Create Issue Button */}
              <button className="inline-flex items-center gap-2 px-6 py-3 bg-accent text-accent-foreground rounded-lg font-medium hover:bg-accent/90 transition-colors">
                <Plus className="w-4 h-4" />
                Create your first issue
              </button>

              {/* Documentation hint */}
              <div className="mt-8 p-4 rounded-lg bg-muted/40 border-l-4 border-accent/60 border-t border-r border-b border-border/60 text-left max-w-lg">
                <p className="text-card-foreground/80 text-xs leading-relaxed">
                  <span className="text-accent font-semibold">Pro tip:</span> Issues support 
                  Markdown formatting, mentions, and can be linked to branches and pull requests 
                  for a seamless development workflow.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default IssuesSection;
