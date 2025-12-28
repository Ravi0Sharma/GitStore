import { Routes, Route } from "react-router-dom"
import { GitProvider } from "./context/GitContext"
import LandingPage from "./pages/landingPage"
import LoginPage from "./pages/loginPage"
import SignUpPage from "./pages/signUp"
import AuthedLayout from "./layouts/AuthedLayout"
import Dashboard from "./pages/Dashboard"
import RepoPage from "./pages/RepoPage"
import IssuesPage from "./pages/IssuesPage"
import IssueDetailPage from "./pages/IssueDetailPage"
import BranchPage from "./pages/BranchPage"
import MergePage from "./pages/MergePage"
import NotFound from "./pages/NotFound"

export default function App() {
  return (
    <GitProvider>
      <Routes>
        {/* Auth & Landing routes - unchanged */}
        <Route path="/" element={<LandingPage />} />
        <Route path="/signin" element={<LoginPage />} />
        <Route path="/signup" element={<SignUpPage />} />
        
        {/* Private routes - require authentication */}
        <Route element={<AuthedLayout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/dashboard/repo/:repoId" element={<RepoPage />} />
          <Route path="/dashboard/repo/:repoId/issues" element={<IssuesPage />} />
          <Route path="/dashboard/repo/:repoId/issues/:issueId" element={<IssueDetailPage />} />
          <Route path="/dashboard/repo/:repoId/branches" element={<BranchPage />} />
          <Route path="/dashboard/repo/:repoId/merge" element={<MergePage />} />
        </Route>
        
        {/* Catch all */}
        <Route path="*" element={<NotFound />} />
      </Routes>
    </GitProvider>
  )
}
