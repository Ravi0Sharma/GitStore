import { GitBranch, LogOut, User } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { signOutUser } from "../firebase";
import { onAuthStateChanged, User as FirebaseUser } from "firebase/auth";
import { firebaseAuth } from "../firebase";
import { useEffect, useState } from "react";

const AuthNavbar = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState<FirebaseUser | null>(null);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(firebaseAuth, (currentUser) => {
      setUser(currentUser);
    });
    return () => unsubscribe();
  }, []);

  const handleSignOut = async () => {
    await signOutUser();
    navigate("/signin");
  };

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-md border-b border-border/50">
      <div className="container mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-primary/10 border border-primary/20">
              <GitBranch className="w-5 h-5 text-primary" />
            </div>
            <Link to="/dashboard">
              <span className="text-xl font-bold text-foreground">GitStore</span>
            </Link>
          </div>

          {/* Navigation Links - Middle */}
          <div className="flex items-center gap-6">
            <Link
              to="/dashboard"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Dashboard
            </Link>
            <Link
              to="/repos"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Repos
            </Link>
            <Link
              to="/issues"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Issues
            </Link>
            <Link
              to="/branches"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Branches
            </Link>
            <Link
              to="/merge"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              Merge
            </Link>
            <Link
              to="/history"
              className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
            >
              History
            </Link>
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

