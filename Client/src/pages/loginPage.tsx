import React, { useState } from "react";
import Navbar from "../components/Navbar";
import { Mail, Lock } from "lucide-react";
import Starfield from "../components/Starfield";
import { Link, useNavigate } from "react-router-dom";
import { signIn, toAuthErrorMessage } from "../firebase";
import { routes } from "../routes";

type FormState = {
  email: string;
  password: string;
  rememberMe: boolean;
};

export default function LoginPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<FormState>({ 
    email: "", 
    password: "",
    rememberMe: false
  });
  const [error, setError] = useState<string>("");
  const [loading, setLoading] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.currentTarget;
    setForm((prev) => ({ 
      ...prev, 
      [name]: type === "checkbox" ? checked : value 
    }));
    setError(""); // Clear error when user types
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const result = await signIn(form.email, form.password);
      
      if (result.error) {
        // result.error is the error code, but we need to create a proper error object
        const errorObj = { code: result.error } as any;
        setError(toAuthErrorMessage(errorObj));
        setLoading(false);
      } else {
        navigate(routes.dashboard);
      }
    } catch (err) {
      console.error('Unexpected error in handleSubmit:', err);
      setError(toAuthErrorMessage(err));
      setLoading(false);
    }
  };

  const handleGoogleSignIn = () => {
    // TODO: handle Google sign in
  };

  return (
    <div className="min-h-screen gradient-bg relative overflow-x-hidden flex items-center justify-center px-6 py-12">
      {/* Starfield Background */}
      <Starfield />

      {/* Navigation */}
      <Navbar />

      {/* Login Card */}
      <div className="relative z-10 w-full max-w-md">
        <div className="rounded-2xl border border-border/50 bg-secondary/30 backdrop-blur-sm p-8 shadow-lg">
          {/* Header */}
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-foreground mb-2">
              Sign in
            </h1>
            <p className="text-muted-foreground">
              Welcome back! Please sign in to continue
            </p>
          </div>

          {/* Google Sign In Button */}
          <button
            type="button"
            onClick={handleGoogleSignIn}
            className="w-full flex items-center justify-center gap-3 px-4 py-3 rounded-lg border border-border/50 bg-card/50 text-foreground font-medium hover:bg-card/70 transition-colors mb-6"
          >
            <svg className="w-5 h-5" viewBox="0 0 24 24">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            Google
          </button>

          {/* Divider */}
          <div className="relative mb-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-border/50"></div>
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-secondary/30 text-muted-foreground">
                or sign in with email
              </span>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 rounded-lg bg-red-500/20 border border-red-500/50 text-red-400 text-sm">
              {error}
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-5">
            {/* Email Field */}
            <div>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  id="email"
                  type="email"
                  name="email"
                  required
                  autoComplete="email"
                  value={form.email}
                  onChange={handleChange}
                  placeholder="Email id"
                  className="w-full pl-10 pr-4 py-3 rounded-lg bg-secondary/50 border border-border/50 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 transition-all"
                />
              </div>
            </div>

            {/* Password Field */}
            <div>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
                <input
                  id="password"
                  type="password"
                  name="password"
                  required
                  autoComplete="current-password"
                  value={form.password}
                  onChange={handleChange}
                  placeholder="Password"
                  className="w-full pl-10 pr-4 py-3 rounded-lg bg-secondary/50 border border-border/50 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-accent/50 focus:border-accent/50 transition-all"
                />
              </div>
            </div>

            {/* Remember Me & Forgot Password */}
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  name="rememberMe"
                  checked={form.rememberMe}
                  onChange={handleChange}
                  className="w-4 h-4 rounded border-border/50 bg-secondary/50 text-accent focus:ring-2 focus:ring-accent/50"
                />
                <span className="text-sm text-foreground">Remember me</span>
              </label>
              <a 
                href="#" 
                className="text-sm text-accent hover:text-accent/80 transition-colors"
              >
                Forgot password?
              </a>
            </div>

            {/* Login Button */}
            <button
              type="submit"
              disabled={loading}
              className="w-full px-4 py-3 rounded-lg bg-gradient-to-r from-accent to-primary text-primary-foreground font-semibold hover:opacity-90 transition-opacity focus:outline-none focus:ring-2 focus:ring-accent/50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Loggar in..." : "Login"}
            </button>
          </form>

          {/* Sign Up Link */}
          <div className="mt-6 text-center">
            
            <Link to={routes.signUp}>
              <p className="text-sm text-muted-foreground">
                Don't have an account?{" "}
                <span className="text-accent hover:text-accent/80 font-medium transition-colors cursor-pointer">
                  Sign up
                </span>
              </p>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
