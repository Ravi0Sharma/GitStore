import { GitBranch, LogOut, User, CircleDot, GitMerge } from "lucide-react";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { signOutUser } from "../firebase";
import { onAuthStateChanged, User as FirebaseUser } from "firebase/auth";
import { firebaseAuth } from "../firebase";
import { useEffect, useState } from "react";
import { routes } from "../routes";

const AuthNavbar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState<FirebaseUser | null>(null);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(firebaseAuth, (currentUser) => {
      setUser(currentUser);
    });
    return () => unsubscribe();
  }, []);

  const handleSignOut = async () => {
    await signOutUser();
    navigate(routes.signIn);
  };

  // Extract repoId from path if we're on a repo page
  const repoMatch = location.pathname.match(/\/dashboard\/repo\/([^/]+)/);
  const repoId = repoMatch ? repoMatch[1] : null;
  const isRepoPage = repoId !== null;

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-md border-b border-border/50">
      <div className="container mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-primary/10 border border-primary/20">
              <GitBranch className="w-5 h-5 text-primary" />
            </div>
            <Link to={routes.dashboard}>
              <span className="text-xl font-bold text-foreground">GitStore</span>
            </Link>
          </div>

          {/* Navigation Links - Middle */}
          <div className="flex items-center gap-6">
            <Link
              to={routes.dashboard}
              className={`text-sm font-medium transition-colors ${
                location.pathname === routes.dashboard
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Dashboard
            </Link>
            
            {isRepoPage && repoId && (
              <>
                <Link
                  to={routes.dashboardRepoIssues(repoId)}
                  className={`text-sm font-medium transition-colors flex items-center gap-1 ${
                    location.pathname.includes('/issues')
                      ? 'text-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <CircleDot className="w-4 h-4" />
                  Issues
                </Link>
                <Link
                  to={routes.dashboardRepoBranches(repoId)}
                  className={`text-sm font-medium transition-colors flex items-center gap-1 ${
                    location.pathname.includes('/branches')
                      ? 'text-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <GitBranch className="w-4 h-4" />
                  Branches
                </Link>
                <Link
                  to={routes.dashboardRepoMerge(repoId)}
                  className={`text-sm font-medium transition-colors flex items-center gap-1 ${
                    location.pathname.includes('/merge')
                      ? 'text-foreground'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <GitMerge className="w-4 h-4" />
                  Merge
                </Link>
              </>
            )}
          </div>

          {/* User Menu - Right */}
          <div className="flex items-center gap-4">
            {user && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <User className="w-4 h-4" />
                <span className="text-foreground">{user.email}</span>
              </div>
            )}
            <button
              onClick={handleSignOut}
              className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              <LogOut className="w-4 h-4" />
              Sign out
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default AuthNavbar;

