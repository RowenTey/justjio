import { AxiosError } from "axios";

declare module "*.jpg";
declare module "*.png";

export type Optional<T> = T | null;

// general
export interface BaseContextResponse<T = never> {
  isSuccessResponse: boolean;
  data?: T;
  error: Optional<AxiosError>;
}

// auth
export interface AuthState {
  accessToken: Optional<string>;
  authenticated: boolean;
}

export type AuthContextType = {
  getAccessToken: () => Optional<string>;
  isAuthenticated: () => boolean;
  logout: () => Promise<boolean>;
  login: (username: string, password: string) => Promise<BaseContextResponse>;
  googleLogin: (code: string) => Promise<BaseContextResponse>;
};
