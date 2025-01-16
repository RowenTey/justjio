import { Route, Routes } from "react-router-dom";
import TabLayout from "../layout/TabLayout";
import HomePage from "../pages/HomePage";
import LoginPage from "../pages/LoginPage";
import SignUpPage from "../pages/SignUpPage";
import ProtectedRoutes from "./ProtectedRoutes";
import RoomPage from "../pages/RoomPage";
import CreateRoomPage from "../pages/CreateRoomPage";
import RoomInvitesPage from "../pages/RoomInvitesPage";
import ProfilePage from "../pages/ProfilePage";
import RoomChatPage from "../pages/RoomChatPage";

const AppRouter = () => {
	return (
		<Routes>
			<Route element={<ProtectedRoutes />}>
				<Route element={<TabLayout />}>
					<Route path="/" element={<HomePage />} />
					<Route path="/profile" element={<ProfilePage />} />
				</Route>
				<Route path="/rooms/create" element={<CreateRoomPage />} />
				<Route path="/rooms/invites" element={<RoomInvitesPage />} />
				<Route path="/room/:roomId" element={<RoomPage />} />
				<Route path="/room/:roomId/chat" element={<RoomChatPage />} />
			</Route>
			<Route path="/login" element={<LoginPage />} />
			<Route path="/signup" element={<SignUpPage />} />
		</Routes>
	);
};

export default AppRouter;
