import React, { createContext, useState } from "react";
import {
  AuthContextType,
  AuthState,
  BaseContextResponse,
  Optional,
} from "../types";
import { authService } from "../services/auth.service";
import { useUserCtx } from "./user";
import useContextWrapper from "../hooks/useContextWrapper";
import { AxiosError } from "axios";
import { DecodedJWTToken, jwtDecode } from "../utils/jwt";

interface LoginResponse {
  data: {
    id: number;
    username: string;
    email: string;
    pictureUrl: string;
  };
  token: string;
  message: string;
  status: string;
}

export const LOGOUT = "LOGOUT";

const AuthContext = createContext<Optional<AuthContextType>>(null);

const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [authState, setAuthState] = useState<AuthState>({
    accessToken: null,
    authenticated: false,
  });
  const { setUser } = useUserCtx();

  const checkAuth = (): boolean => {
    const accessToken = localStorage.getItem("accessToken");
    if (!accessToken) return false;

    // Decode the token to get the user's info
    const decodedToken = jwtDecode<DecodedJWTToken>(accessToken);
    // Check if token is expired
    if (decodedToken.exp * 1000 < Date.now()) {
      // Clear state and log user out
      logout();
      return false;
    }

    setUser({
      id: decodedToken.user_id,
      email: decodedToken.user_email,
      username: decodedToken.username,
      pictureUrl: decodedToken.picture_url,
    });

    setAuthState({
      accessToken,
      authenticated: true,
    });

    return true;
  };

  const handleLoginResponse = (res: LoginResponse) => {
    const { data, token } = res;
    localStorage.setItem("accessToken", token);
    setAuthState({
      accessToken: token,
      authenticated: true,
    });
    setUser({
      id: data.id,
      email: data.email,
      username: data.username,
      pictureUrl: data.pictureUrl,
    });
  };

  const login = async (
    username: string,
    password: string,
  ): Promise<BaseContextResponse> => {
    try {
      const res = await authService.login({ username, password });
      handleLoginResponse(res as LoginResponse);
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Error logging in: ", error);
      return {
        isSuccessResponse: false,
        error: error as AxiosError,
      };
    }
  };

  const googleLogin = async (code: string): Promise<BaseContextResponse> => {
    try {
      const res = await authService.googleAuth({ code });
      console.log("Google login response: ", res);
      handleLoginResponse(res as LoginResponse);
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Error logging in: ", error);
      return {
        isSuccessResponse: false,
        error: error as AxiosError,
      };
    }
  };

  const logout = async () => {
    return new Promise<boolean>((resolve) => {
      setTimeout(() => {
        setAuthState({
          accessToken: null,
          authenticated: false,
        });
        localStorage.removeItem("accessToken");
        setUser({
          id: -1,
          email: "",
          username: "",
          pictureUrl: "",
        });
        resolve(true);
      }, 500);
    });
  };

  const getAccessToken = () => {
    return authState.accessToken;
  };

  const isAuthenticated = () => {
    return authState.authenticated ? true : checkAuth();
  };

  return (
    <AuthContext.Provider
      value={{
        login,
        logout,
        googleLogin,
        getAccessToken,
        isAuthenticated,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

const useAuth = () => useContextWrapper(AuthContext);

export { useAuth, AuthProvider };
