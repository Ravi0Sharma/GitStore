import { Routes, Route } from "react-router-dom"
import LandingPage from "./pages/landingPage"
import LoginPage from "./pages/loginPage"
import SignUpPage from "./pages/signUp"
import Dashboard from "./pages/Dashboard"

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/signin" element={<LoginPage />} />
      <Route path="/signup" element={<SignUpPage />} />
      <Route path="/dashboard" element={<Dashboard />} />
    </Routes>
  )
}
