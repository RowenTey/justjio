import { useEffect, useState } from "react";
import { IUser } from "../types/user";
import {
	sendFriendRequestApi,
	fetchFriendsApi,
	removeFriendApi,
	countPendingFriendRequestsApi,
} from "../api/user";
import { useUserCtx } from "../context/user";
import { api } from "../api";
import { TrashIcon } from "@heroicons/react/24/outline";
import SearchUserModal from "../components/modals/SearchUserModal";
import { ArrowLeftIcon, UserGroupIcon } from "@heroicons/react/24/solid";
import { useNavigate } from "react-router-dom";
import { useToast } from "../context/toast";

type FriendsTopBarProps = {
	title: string;
	userId: number;
};

const FriendsTopBar: React.FC<FriendsTopBarProps> = ({ userId, title }) => {
	const navigate = useNavigate();
	const [numFriendRequests, setNumFriendRequests] = useState(0);

	useEffect(() => {
		const fetchFriendRequests = async () => {
			const res = await countPendingFriendRequestsApi(api, userId);
			setNumFriendRequests(res.data.data.count);
		};

		fetchFriendRequests();
	}, [userId]);

	return (
		<div
			className={`relative top-0 flex h-[8%] items-center w-full py-4 px-3 bg-purple-200 justify-between`}
		>
			<button
				onClick={() => navigate(-1)}
				className={`flex items-center justify-center p-1 hover:scale-110 `}
			>
				<ArrowLeftIcon className="w-6 h-6 text-black" />
			</button>

			<h1 className={`text-xl font-bold text-secondary`}>{title}</h1>

			<button
				onClick={() => navigate("/friendRequests")}
				className={`flex items-center justify-center p-1`}
			>
				<UserGroupIcon className="w-8 h-8 text-secondary hover:scale-110" />
				{numFriendRequests > 0 && (
					<div className="absolute top-[4px] right-[7px] w-4 h-4 bg-red-600 rounded-full flex items-center justify-center text-white text-xs font-bold p-1">
						{numFriendRequests}
					</div>
				)}
			</button>
		</div>
	);
};

const FriendsPage = () => {
	const [friends, setFriends] = useState<IUser[]>([]);
	const { user } = useUserCtx();
	const [isSearchModalVisible, setIsSearchModalVisible] = useState(false);
	const { showToast } = useToast();

	useEffect(() => {
		const fetchFriends = async () => {
			const res = await fetchFriendsApi(api, user.id);
			setFriends(res.data.data);
		};

		fetchFriends();
	}, [user.id]);

	const handleSendFriendRequest = async (newFriend: IUser) => {
		try {
			const res = await sendFriendRequestApi(api, user.id, newFriend.id);
			if (res.status !== 200) {
				switch (res.status) {
					case 400:
						showToast("Bad request, please check request body.", true);
						break;
					case 404:
						showToast("User not found, please try again later.", true);
						break;
					case 409:
						showToast(res.data.message, true);
						break;
					case 500:
					default:
						showToast("An error occurred, please try again later.", true);
						break;
				}
				return;
			}

			showToast("Friend request sent!", false);
			setIsSearchModalVisible(false);
		} catch (error) {
			console.error(error);
			showToast("An error occurred, please try again later.", true);
		}
	};

	const handleRemoveFriend = async (friendId: number) => {
		try {
			const res = await removeFriendApi(api, user.id, friendId);
			if (res.status !== 200) {
				switch (res.status) {
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
				return;
			}

			showToast("Friend removed!", false);
			setFriends((prevFriends) =>
				prevFriends.filter((friend) => friend.id !== friendId)
			);
		} catch (error) {
			showToast("An error occurred, please try again later.", true);
		}
	};

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<FriendsTopBar userId={user.id} title="Friends" />

			<div className="w-full h-full flex flex-col items-center px-4 gap-3">
				<div className="w-full h-[85%] overflow-y-auto flex flex-col items-center justify-center gap-4">
					{friends.length > 0 ? (
						friends.map((friend) => (
							<div
								key={friend.id}
								className="w-4/5 flex items-center justify-between py-2 px-3 bg-white rounded-xl shadow-md"
							>
								<div className="flex items-center gap-2">
									<img
										src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
										alt="Profile Image"
										className="w-7 h-7 rounded-full"
									/>
									<p className="text-black">{friend.username}</p>
								</div>
								<TrashIcon
									className="h-6 w-6 text-red-500 hover:text-red-600 cursor-pointer"
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
