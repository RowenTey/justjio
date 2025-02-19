import { AxiosError } from "axios";

declare module "*.jpg";
declare module "*.png";

// general
export interface BaseContextResponse<T = never> {
	isSuccessResponse: boolean;
	data?: T;
	error: AxiosError | null;
}

// auth
export interface AuthState {
	accessToken: string | undefined;
	authenticated: boolean;
}

export type AuthContextType = {
	getAccessToken: () => string | undefined;
	isAuthenticated: () => boolean;
	logout: () => Promise<boolean>;
	login: (username: string, password: string) => Promise<BaseContextResponse>;
	googleLogin: (code: string) => Promise<BaseContextResponse>;
};
