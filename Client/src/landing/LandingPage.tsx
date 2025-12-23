import Navbar from "./components/Navbar";
import HeroSection from "./components/HeroSection";
import TimelineSection from "./components/TimelineSection";
import IssuesSection from "./components/IssuesSection";
import IssueManagementSection from "./components/IssueManagementSection";
import ViewHistorySection from "./components/ViewHistorySection";
import Starfield from "./components/Starfield";

const LandingPage = () => {
  return (
    <div className="min-h-screen gradient-bg relative overflow-x-hidden">
      {/* Starfield Background */}
      <Starfield />

      {/* Navigation */}
      <Navbar />

      {/* Hero Section */}
      <HeroSection />

      {/* Timeline with Feature Card */}
      <TimelineSection />

      {/* Issues Section */}
      <IssuesSection />

      {/* Issue Management Section */}
      <IssueManagementSection />

      {/* View History Section */}
      <ViewHistorySection glowColor="accent" />

      {/* Footer */}
      <footer className="relative z-10 border-t border-border/30 py-12">
        <div className="container mx-auto px-6">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <span className="text-foreground font-semibold">GitStore</span>
              <span className="text-muted-foreground">© 2025</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-muted-foreground">
              <a href="#" className="hover:text-foreground transition-colors">Privacy</a>
              <a href="#" className="hover:text-foreground transition-colors">Terms</a>
              <a href="#" className="hover:text-foreground transition-colors">Contact</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default LandingPage;

