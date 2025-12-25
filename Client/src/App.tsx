import { Routes, Route } from "react-router-dom"
import LandingPage from "./pages/landingPage/landingPage"
import LoginPage from "./pages/loginPage"

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/SignIn" element={<LoginPage />} />
    </Routes>
  )
}
