import { Outlet } from "react-router-dom";
import AuthNavbar from "../components/AuthNavbar";

export default function AuthedLayout() {
  return (
    <div className="min-h-screen bg-background pt-20">
      <AuthNavbar />
      <Outlet />
    </div>
  );
}

