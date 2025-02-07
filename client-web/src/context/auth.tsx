import React, { createContext, useState } from "react";
import { AuthContextType, AuthState, BaseContextResponse } from "../types";
import { loginApi } from "../api/auth";
import { api } from "../api";
import { initialUserState, useUserCtx } from "./user";
import useContextWrapper from "../hooks/useContextWrapper";
import { AxiosError } from "axios";

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

	const login = async (
		username: string,
		password: string
	): Promise<BaseContextResponse> => {
		let res = null;
		try {
			res = await loginApi(api, username, password, false);

			const { data, token } = res.data;
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
		} catch (error) {
			const err = error as AxiosError;
			console.error(err);
			return { isSuccessResponse: false, data: null, error: err };
		}

		return { isSuccessResponse: true, data: res.data, error: null };
	};

	const logout = async () => {
		return new Promise<boolean>((resolve) => {
			setTimeout(() => {
				setAuthState({
					accessToken: undefined,
					authenticated: false,
				});
				localStorage.removeItem("accessToken");
				setUser(initialUserState);
				resolve(true);
			}, 500);
		});
	};

	const getAccessToken = () => {
		return authState.accessToken;
	};

	const isAuthenticated = () => {
		return authState.authenticated;
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
