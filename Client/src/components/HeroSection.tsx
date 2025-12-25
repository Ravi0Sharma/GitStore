import { ArrowRight, Play } from "lucide-react";

const HeroSection = () => {
  return (
    <section className="relative flex flex-col items-center justify-center px-6 pt-24 pb-12 md:pt-32 md:pb-16">
      {/* Badge */}
      <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full border border-border/50 bg-secondary/50 backdrop-blur-sm mb-8 animate-fade-in-up">
        <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
        <span className="text-sm text-muted-foreground">
          Web-based Git CLI
        </span>
      </div>

      {/* Main Heading */}
      <h1 className="text-4xl md:text-5xl lg:text-6xl xl:text-7xl font-bold text-center max-w-5xl leading-tight mb-6 animate-fade-in-up" style={{ animationDelay: "0.1s" }}>
        <span className="text-foreground">
          Web-based Git interface for{" "}
        </span>
        <span className="text-gradient-accent">issues</span>
        <span className="text-foreground">, </span>
        <span className="text-gradient-accent">commits</span>
        <span className="text-foreground">, </span>
        <span className="text-gradient-accent">merges</span>
        <span className="text-foreground"> and </span>
        <span className="text-gradient-accent">history</span>
        <span className="text-foreground">. Simple by default.</span>
      </h1>

      {/* Subtitle */}
      <p className="text-muted-foreground text-lg md:text-xl text-center max-w-2xl mb-10 animate-fade-in-up" style={{ animationDelay: "0.2s" }}>
        The modern Git experience. No terminal needed.
      </p>

      {/* CTA Buttons */}
      <div className="flex flex-col sm:flex-row items-center gap-4 animate-fade-in-up" style={{ animationDelay: "0.3s" }}>
        <button className="group flex items-center gap-2 px-6 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-all duration-200 hover:gap-3">
          Get Started
          <ArrowRight className="w-4 h-4 transition-transform group-hover:translate-x-1" />
        </button>
        <button className="flex items-center gap-2 px-6 py-3 border border-border/50 bg-secondary/30 text-foreground rounded-lg font-medium hover:bg-secondary/50 transition-colors">
          <Play className="w-4 h-4" />
          Watch Demo
        </button>
      </div>

      {/* Video Demo Placeholder */}
      <div className="mt-16 md:mt-24 w-full max-w-4xl animate-fade-in-up" style={{ animationDelay: "0.4s" }}>
        <div className="relative rounded-2xl border border-border/50 bg-secondary/20 backdrop-blur-sm overflow-hidden aspect-video">
          {/* Video placeholder content */}
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <div className="w-20 h-20 rounded-full bg-primary/20 border border-primary/30 flex items-center justify-center mb-4 hover:bg-primary/30 transition-colors cursor-pointer group">
              <Play className="w-8 h-8 text-primary group-hover:scale-110 transition-transform" />
            </div>
            <p className="text-muted-foreground text-sm">Video demo coming soon</p>
          </div>
          
          {/* Decorative gradient overlay */}
          <div className="absolute inset-0 bg-gradient-to-t from-background/50 to-transparent pointer-events-none" />
        </div>
      </div>
    </section>
  );
};

export default HeroSection;

