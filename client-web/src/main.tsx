import ReactDOM from "react-dom/client";
import React from "react";
import App from "./App.tsx";
import { AppContextProvider } from "./context/index.tsx";
import { GoogleOAuthProvider } from "@react-oauth/google";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <GoogleOAuthProvider clientId={import.meta.env.VITE_GOOGLE_CLIENT_ID}>
      <AppContextProvider>
        <App />
      </AppContextProvider>
    </GoogleOAuthProvider>
  </React.StrictMode>,
);
