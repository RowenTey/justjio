import { Outlet, Navigate } from "react-router-dom";
import { useAuth } from "../context/auth";
import { useEffect, useState } from "react";
import { useUserCtx } from "../context/user";
import Spinner from "../components/Spinner";

const ProtectedRoutes = () => {
	const { isAuthenticated } = useAuth();
	const { user } = useUserCtx();
	const [allowed, setAllowed] = useState(false);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		if (!isAuthenticated()) {
			setAllowed(false);
			setLoading(false);
			return;
		}

		// Keep loading if user is still not set by UserContext yet
		if (!user || !user.id || user.id === -1) {
			setAllowed(false);
			return;
		}

		setAllowed(true);
		setLoading(false);
	}, [user, isAuthenticated]);

	if (loading) {
		return <Spinner bgClass="bg-primary" />;
	}

	return allowed ? <Outlet /> : <Navigate to="/login" />;
};

export default ProtectedRoutes;
