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
import CreateBillPage from "../pages/CreateBillPage";
import SplitBillPage from "../pages/SplitBillPage";
import FriendsPage from "../pages/FriendsPage";
import FriendRequestsPage from "../pages/FriendRequestPage";
import RoomsPage from "../pages/RoomsPage";
import NotificationsPage from "../pages/NotificationsPage";
import VerifyOTPPage from "../pages/VerifyOTPPage";
import ForgotPasswordPage from "../pages/ForgotPasswordPage";
import ResetPasswordPage from "../pages/ResetPasswordPage";
import EditProfilePage from "../pages/EditProfilePage";
import EditRoomPage from "../pages/EditRoomPage";

const AppRouter = () => {
  return (
    <Routes>
      <Route element={<ProtectedRoutes />}>
        <Route element={<TabLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
        </Route>
        <Route path="/profile/edit" element={<EditProfilePage />} />
        <Route path="/friends" element={<FriendsPage />} />
        <Route path="/friends/requests" element={<FriendRequestsPage />} />
        <Route path="/rooms" element={<RoomsPage />} />
        <Route path="/rooms/create" element={<CreateRoomPage />} />
        <Route path="/rooms/invites" element={<RoomInvitesPage />} />
        <Route path="/room/:roomId" element={<RoomPage />} />
        <Route path="/room/:roomId/chat" element={<RoomChatPage />} />
        <Route path="/room/:roomId/edit" element={<EditRoomPage />} />
        <Route path="/room/:roomId/bill/create" element={<CreateBillPage />} />
        <Route path="/room/:roomId/bill/split" element={<SplitBillPage />} />
      </Route>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignUpPage />} />
      <Route path="/otp" element={<VerifyOTPPage />} />
      <Route path="/password/forgot" element={<ForgotPasswordPage />} />
      <Route path="/password/reset" element={<ResetPasswordPage />} />
    </Routes>
  );
};

export default AppRouter;
