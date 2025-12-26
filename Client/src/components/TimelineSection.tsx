import { GitBranch, GitCommit, FileText, GitMerge, Clock, Database } from "lucide-react";

const TimelineSection = () => {
  const features = [
    { icon: FileText, label: "Create Issues", dot: 1 },
    { icon: GitCommit, label: "Commit Changes", dot: 2 },
    { icon: GitMerge, label: "Merge Branches", dot: 3 },
    { icon: Clock, label: "View History", dot: 4 },
  ];

  return (
    <section className="relative pt-8 pb-16 md:pt-12 md:pb-24 overflow-hidden mx-auto w-[60vw] max-w-4xl min-w-[320px]">
  
            {/* Feature List - Single Row with Numbered Dots */}
            <div className="flex flex-wrap lg:flex-nowrap items-start gap-6 mt-12">
              {features.map((feature, index) => (
                <div
                  key={index}
                  className="flex flex-col items-center gap-3 flex-1 min-w-[140px]"
                >
                  {/* Numbered Dot */}
                  <div 
                    className="w-8 h-8 rounded-full bg-accent/20 border-2 border-accent flex items-center justify-center text-accent font-bold text-sm"
                    style={{ boxShadow: "0 0 12px hsl(350 85% 50% / 0.4)" }}
                  >
                    {feature.dot}
                  </div>
                  
                  {/* Connecting line */}
                  <div className="w-0.5 h-4 bg-accent/50" />
                  
                  {/* Feature Card */}
                  <div className="p-4 rounded-xl bg-secondary/50 border border-border/50 hover:bg-secondary/70 transition-colors w-full">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-lg bg-accent/10 border border-accent/20 flex items-center justify-center">
                        <feature.icon className="w-5 h-5 text-accent" />
                      </div>
                      <span className="text-foreground font-medium text-sm">{feature.label}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* GitDb Description Section */}
            <div className="mt-16 p-6 rounded-2xl bg-secondary/30 border border-border/50 backdrop-blur-sm">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/30 flex items-center justify-center flex-shrink-0"
                  style={{ boxShadow: "0 0 20px hsl(15 90% 55% / 0.3)" }}
                >
                  <Database className="w-6 h-6 text-primary" />
                </div>
                <div>
                  <h3 className="text-foreground font-semibold text-lg mb-2">GitDb - Integrated Database</h3>
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    GitDb is a lightweight, embedded key–value database built for version-control workloads, 
                    with fast append-only writes and quick recovery.
                  </p>
                </div>
              </div>
            </div>
    </section>
  );
};

export default TimelineSection;