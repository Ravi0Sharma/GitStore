import { Outlet } from "react-router-dom";
import AuthNavbar from "../components/AuthNavbar";

export default function AuthedLayout() {
  return (
    <>
      <AuthNavbar />
      <Outlet />
    </>
  );
}

