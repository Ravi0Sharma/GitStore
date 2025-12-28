import { GitBranch } from "lucide-react";
import { useNavigate, Link} from "react-router-dom"
import { routes } from "../routes";


const Navbar = () => {
  const navigate = useNavigate()
  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-md border-b border-border/50">
      <div className="container mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-primary/10 border border-primary/20">
              <GitBranch className="w-5 h-5 text-primary" />
            </div>
            <Link to={routes.home}><span className="text-xl font-bold text-foreground">GitStore</span>
            </Link>
            
          </div>

          {/* CTA Buttons */}
          <div className="flex items-center gap-4">
           <button onClick={() => navigate(routes.signIn)}>
              Log in
            </button>
            <button  onClick={() => navigate(routes.signUp) } className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors">
              Get Started
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;

