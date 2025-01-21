import React, { createContext, useState } from "react";
import { BaseUserInfo, UserContextType } from "../types/user";
import useContextWrapper from "../hooks/useContextWrapper";

const initialUserState: BaseUserInfo = {
	id: -1,
	email: "",
	username: "",
};

const UserContext = createContext<UserContextType | null>(null);

const UserProvider: React.FC<{ children: React.ReactNode }> = ({
	children,
}) => {
	const [user, setUser] = useState<BaseUserInfo>(initialUserState);

	return (
		<UserContext.Provider value={{ user, setUser }}>
			{children}
		</UserContext.Provider>
	);
};

const useUserCtx = () => useContextWrapper(UserContext);

export { useUserCtx, initialUserState, UserProvider };
