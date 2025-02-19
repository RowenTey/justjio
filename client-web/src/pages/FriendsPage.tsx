/* eslint-disable react-hooks/exhaustive-deps */
import { useEffect, useState } from "react";
import { IUser } from "../types/user";
import { sendFriendRequestApi } from "../api/user";
import { useUserCtx } from "../context/user";
import { api } from "../api";
import { TrashIcon } from "@heroicons/react/24/outline";
import SearchUserModal from "../components/modals/SearchUserModal";
import { useToast } from "../context/toast";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { AxiosError } from "axios";
import FriendsTopBar from "../components/top-bar/FriendsTopBar";

const FriendsPage = () => {
	const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
	const { user, friends, fetchFriends, removeFriend } = useUserCtx();
	const [isSearchModalVisible, setIsSearchModalVisible] = useState(false);
	const { showToast } = useToast();

	useEffect(() => {
		startLoading();
		fetchFriends(user.id).then(() => stopLoading());
	}, [user.id]);

	const handleSendFriendRequest = async (newFriend: IUser) => {
		startLoading();
		try {
			await sendFriendRequestApi(api, user.id, newFriend.id);
			showToast("Friend request sent!", false);
		} catch (error) {
			console.error("An error occurred while sending friend request: ", error);
			switch ((error as AxiosError).response?.status) {
				case 400:
					showToast("Bad request, please check request body.", true);
					break;
				case 404:
					showToast("User not found, please try again later.", true);
					break;
				case 409:
					showToast(
						(error as AxiosError<{ message: string }>).response?.data
							?.message || "An error occurred, please try again later.",
						true
					);
					break;
				case 500:
				default:
					showToast("An error occurred, please try again later.", true);
					break;
			}
		} finally {
			stopLoading();
		}
	};

	const handleRemoveFriend = async (friendId: number) => {
		startLoading();
		const res = await removeFriend(user.id, friendId);
		if (!res.isSuccessResponse) {
			switch (res.error?.response?.status) {
				case 400:
					showToast("Bad request, please check request body.", true);
					break;
				case 404:
					showToast("User not found, please try again later.", true);
					break;
				case 500:
				default:
					showToast("An error occurred, please try again later.", true);
					break;
			}
			stopLoading();
			return;
		}
		showToast("Friend removed!", false);
		stopLoading();
	};

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<FriendsTopBar userId={user.id} title="Friends" />

			<div className="w-full h-full flex flex-col items-center px-4 gap-3">
				<div
					className={`w-full h-[85%] overflow-y-auto flex flex-col items-center ${
						loadingStates[0]
							? "justify-center"
							: friends.length > 0
							? ""
							: "justify-center"
					} gap-4`}
				>
					{loadingStates[0] ? (
						<Spinner spinnerSize={{ width: "w-10", height: "h-10" }} />
					) : friends.length > 0 ? (
						friends.map((friend) => (
							<div
								key={friend.id}
								className="w-4/5 flex items-center justify-between py-2 px-3 bg-white rounded-xl shadow-md"
							>
								<div className="flex items-center gap-2">
									<img
										// src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
										src={friend.pictureUrl}
										alt="Profile Image"
										className="w-7 h-7 rounded-full"
									/>
									<p className="text-black">{friend.username}</p>
								</div>
								<TrashIcon
									className="h-6 w-6 text-error hover:scale-110 cursor-pointer"
									onClick={() => handleRemoveFriend(friend.id)}
								/>
							</div>
						))
					) : (
						<p className="text-lg font-semibold text-gray-500">
							No friends found
						</p>
					)}
				</div>

				<button
					onClick={() => setIsSearchModalVisible(true)}
					className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
				>
					Add Friend
				</button>
			</div>

			<SearchUserModal
				isVisible={isSearchModalVisible}
				closeModal={() => setIsSearchModalVisible(false)}
				userId={user.id}
				sendFriendRequest={handleSendFriendRequest}
			/>
		</div>
	);
};

export default FriendsPage;
