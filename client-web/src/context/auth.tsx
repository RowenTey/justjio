import React, { createContext, useState } from "react";
import { AuthContextType, AuthState, BaseContextResponse } from "../types";
import { googleLoginApi, loginApi, LoginResponse } from "../api/auth";
import { api } from "../api";
import { useUserCtx } from "./user";
import useContextWrapper from "../hooks/useContextWrapper";
import { AxiosError } from "axios";
import { DecodedJWTToken, jwtDecode } from "../utils/jwt";

export const LOGOUT = "LOGOUT";

const AuthContext = createContext<AuthContextType | null>(null);

const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [authState, setAuthState] = useState<AuthState>({
    accessToken: undefined,
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
      const data = await loginApi(api, username, password, false);
      console.log("Login data", data.headers);
      handleLoginResponse(data.data);
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
      const { data: res } = await googleLoginApi(api, code);
      console.log("Google login response: ", res);
      handleLoginResponse(res);
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
          accessToken: undefined,
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
