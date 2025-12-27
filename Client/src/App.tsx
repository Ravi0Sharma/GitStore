import { Routes, Route } from "react-router-dom"
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
    <Routes>
      {/* Auth & Landing routes - unchanged */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/signin" element={<LoginPage />} />
      <Route path="/signup" element={<SignUpPage />} />
      
      {/* Private routes - require authentication */}
      <Route element={<AuthedLayout />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/repo/:repoId" element={<RepoPage />} />
        <Route path="/repo/:repoId/issues" element={<IssuesPage />} />
        <Route path="/repo/:repoId/issues/:issueId" element={<IssueDetailPage />} />
        <Route path="/repo/:repoId/branches" element={<BranchPage />} />
        <Route path="/repo/:repoId/merge" element={<MergePage />} />
      </Route>
      
      {/* Catch all */}
      <Route path="*" element={<NotFound />} />
    </Routes>
  )
}
