import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { onAuthStateChanged, User } from "firebase/auth";
import { firebaseAuth, signOutUser } from "../firebase";
import Navbar from "../components/Navbar";
import Starfield from "../components/Starfield";
import { LogOut } from "lucide-react";

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

  const handleSignOut = async () => {
    await signOutUser();
    navigate("/signin");
  };

  if (loading) {
    return (
      <div className="min-h-screen gradient-bg relative overflow-x-hidden flex items-center justify-center">
        <Starfield />
        <div className="relative z-10 text-foreground">Laddar...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen gradient-bg relative overflow-x-hidden flex items-center justify-center px-6 py-12">
      <Starfield />
      <Navbar />

      <div className="relative z-10 w-full max-w-md">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 shadow-lg">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Dashboard
            </h1>
            <p className="text-muted-foreground">
              Välkommen! Du är inloggad.
            </p>
          </div>

          <div className="space-y-6">
            <div className="p-4 rounded-lg bg-secondary/50 border border-border/50">
              <p className="text-sm text-muted-foreground mb-1">E-postadress</p>
              <p className="text-foreground font-medium">{user?.email}</p>
            </div>

            <button
              onClick={handleSignOut}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg bg-gradient-to-r from-red-500/20 to-red-600/20 border border-red-500/50 text-red-400 font-semibold hover:opacity-90 transition-opacity focus:outline-none focus:ring-2 focus:ring-red-500/50"
            >
              <LogOut className="w-5 h-5" />
              Sign out
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

