import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { onAuthStateChanged, User } from "firebase/auth";
import { firebaseAuth } from "../firebase";
import Starfield from "../components/Starfield";
import RepoList from "../components/RepoList";
import { BookMarked, CircleDot, GitBranch, GitMerge } from "lucide-react";


export default function Dashboard() {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(firebaseAuth, (currentUser) => {
      if (currentUser) {
        setUser(currentUser);
      } else {
        navigate("/signin");
      }
      setLoading(false);
    });

    return () => unsubscribe();
  }, [navigate]);

  if (loading) {
    return (
      <div className="min-h-screen gradient-bg relative overflow-x-hidden flex items-center justify-center">
        <Starfield />
        <div className="relative z-10 text-foreground">Laddar...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen gradient-bg relative overflow-x-hidden">
      <Starfield />

      <div className="relative z-10 container mx-auto px-6 py-24">
        <div className="max-w-6xl mx-auto">
          {/* Welcome Section */}
          <div className="mb-12">
            <h1 className="text-4xl font-bold text-foreground mb-2">Dashboard</h1>
            <p className="text-muted-foreground">
              Welcome back, {user?.email}
            </p>
          </div>

          {/* Quick Links Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
            <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-6">
              <div className="flex items-center gap-4 mb-4">
                <div className="p-3 rounded-lg bg-primary/10 border border-primary/20">
                  <BookMarked className="w-6 h-6 text-primary" />
                </div>
                <h3 className="text-xl font-semibold text-foreground">
                  Repositories
                </h3>
              </div>
              <p className="text-muted-foreground">View and manage your repositories below</p>
            </div>

            <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-6">
              <div className="flex items-center gap-4 mb-4">
                <div className="p-3 rounded-lg bg-accent/10 border border-accent/20">
                  <CircleDot className="w-6 h-6 text-accent" />
                </div>
                <h3 className="text-xl font-semibold text-foreground">
                  Issues
                </h3>
              </div>
              <p className="text-muted-foreground">Access issues from repository pages</p>
            </div>

            <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-6">
              <div className="flex items-center gap-4 mb-4">
                <div className="p-3 rounded-lg bg-green-500/10 border border-green-500/20">
                  <GitBranch className="w-6 h-6 text-green-400" />
                </div>
                <h3 className="text-xl font-semibold text-foreground">
                  Branches
                </h3>
              </div>
              <p className="text-muted-foreground">Manage branches from repository pages</p>
            </div>

            <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-6">
              <div className="flex items-center gap-4 mb-4">
                <div className="p-3 rounded-lg bg-blue-500/10 border border-blue-500/20">
                  <GitMerge className="w-6 h-6 text-blue-400" />
                </div>
                <h3 className="text-xl font-semibold text-foreground">
                  Merge
                </h3>
              </div>
              <p className="text-muted-foreground">Merge branches from repository pages</p>
            </div>
          </div>

          {/* Repositories Section */}
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-foreground mb-6">Recent Repositories</h2>
            <RepoList />
          </div>
        </div>
      </div>
    </div>
  );
}

