import { Routes, Route } from "react-router-dom"
import LandingPage from "./pages/landingPage"
import LoginPage from "./pages/loginPage"
import SignUpPage from "./pages/signUp"

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/SignIn" element={<LoginPage />} />
      <Route path="/SignUp" element={<SignUpPage />} />
    </Routes>
  )
}
