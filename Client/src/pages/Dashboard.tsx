import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { onAuthStateChanged, User } from "firebase/auth";
import { firebaseAuth, signOutUser } from "../firebase";
import Navbar from "../components/Navbar";
import Starfield from "../components/Starfield";
import  RepoList from "../components/RepoList"


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
    <div className="min-h-screen  bg-background relative overflow-x-hidden flex items-center justify-center px-6 py-12">
      <Starfield />
      <Navbar />

    <RepoList/>
    </div>
  );
}

