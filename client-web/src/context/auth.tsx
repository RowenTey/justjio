import React, { createContext, useState } from "react";
import { AuthContextType, AuthState, BaseContextResponse } from "../types";
import { loginApi } from "../api/auth";
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
		setUser({
			id: decodedToken.user_id,
			email: decodedToken.user_email,
			username: decodedToken.username,
		});

		setAuthState({
			accessToken,
			authenticated: true,
		});

		return true;
	};

	const login = async (
		username: string,
		password: string
	): Promise<BaseContextResponse> => {
		try {
			const { data: res } = await loginApi(api, username, password, false);

			const { data, token } = res;
			setAuthState({
				accessToken: token,
				authenticated: true,
			});
			localStorage.setItem("accessToken", token);
			setUser({
				id: data.id,
				email: data.email,
				username: data.username,
			});

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
