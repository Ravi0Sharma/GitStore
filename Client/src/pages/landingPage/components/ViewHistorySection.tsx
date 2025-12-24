
import { History, GitCommit } from "lucide-react";

interface ViewHistorySectionProps {
  glowColor?: string;
}

const ViewHistorySection = ({ glowColor = "accent" }: ViewHistorySectionProps) => {
  const getGlowStyle = () => {
    switch (glowColor) {
      case "primary":
        return "0 0 60px hsl(15 90% 55% / 0.4), 0 0 120px hsl(15 90% 55% / 0.2)";
      case "accent":
        return "0 0 60px hsl(350 85% 50% / 0.4), 0 0 120px hsl(350 85% 50% / 0.2)";
      case "blue":
        return "0 0 60px hsl(210 100% 50% / 0.4), 0 0 120px hsl(210 100% 50% / 0.2)";
      case "green":
        return "0 0 60px hsl(142 76% 36% / 0.4), 0 0 120px hsl(142 76% 36% / 0.2)";
      case "purple":
        return "0 0 60px hsl(270 70% 50% / 0.4), 0 0 120px hsl(270 70% 50% / 0.2)";
      default:
        return "0 0 60px hsl(350 85% 50% / 0.4), 0 0 120px hsl(350 85% 50% / 0.2)";
    }
  };

  const commits = [
    { hash: "a3f2b1c", message: "Add user authentication", time: "2 hours ago" },
    { hash: "7d4e9f2", message: "Fix navigation styling", time: "5 hours ago" },
    { hash: "1b8c3a5", message: "Update dependencies", time: "1 day ago" },
  ];

  return (
    <section className="relative py-24">
      <div className="container mx-auto px-6">
        <div className="max-w-4xl mx-auto">
          {/* Section Header */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white/10 backdrop-blur-sm border border-white/20 mb-6 shadow-lg">
              <History className="w-4 h-4 text-white" />
              <span className="text-white text-sm font-semibold">View History</span>
            </div>
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              <span className="text-foreground">Track every </span>
              <span className="text-white drop-shadow-lg">change</span>
            </h2>
            <p className="text-foreground/90 text-lg max-w-2xl mx-auto drop-shadow-sm">
              View your complete commit history with detailed information about each change. 
              Navigate through time to understand how your project evolved.
            </p>
          </div>

          {/* Layered Rectangle Component with Glow */}
          <div 
            className="relative"
            style={{ filter: `drop-shadow(${getGlowStyle()})` }}
          >
            {/* Base Rectangle (Background layer) */}
            <div className="absolute inset-0 translate-x-4 translate-y-4 bg-card/40 rounded-2xl border border-border/20" />
            
            {/* Second Rectangle (Middle layer - image preview) */}
            <div className="absolute inset-0 translate-x-2 translate-y-2 bg-card/60 rounded-2xl border border-border/30 overflow-hidden">
              {/* Simulated preview content */}
              <div className="absolute inset-0 bg-gradient-to-br from-muted/20 to-transparent" />
            </div>

            {/* Main Rectangle (Front layer) */}
            <div className="relative bg-card rounded-2xl overflow-hidden border border-border/50">
              {/* Header */}
              <div className="px-6 py-4 border-b border-border/30 flex items-center gap-3">
                <History className="w-5 h-5 text-card-foreground" />
                <span className="text-card-foreground font-semibold">Commit History</span>
              </div>

              {/* Commit List */}
              <div className="divide-y divide-border/20">
                {commits.map((commit, index) => (
                  <div
                    key={index}
                    className="px-6 py-4 flex items-center gap-4 hover:bg-muted/5 transition-colors"
                  >
                    <div className="w-8 h-8 rounded-full bg-muted/30 flex items-center justify-center">
                      <GitCommit className="w-4 h-4 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <code className="text-xs font-mono text-primary font-semibold bg-primary/15 px-2 py-0.5 rounded border border-primary/20">
                          {commit.hash}
                        </code>
                        <span className="text-card-foreground text-sm truncate">
                          {commit.message}
                        </span>
                      </div>
                      <span className="text-xs text-card-foreground/65">{commit.time}</span>
                    </div>
                  </div>
                ))}
              </div>

              {/* Footer hint */}
              <div className="px-6 py-4 border-t border-border/30 bg-muted/30">
                <p className="text-xs text-card-foreground/75 text-center font-medium">
                  Click on any commit to view detailed changes and file diffs
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ViewHistorySection;
