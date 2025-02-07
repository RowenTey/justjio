import { useEffect, useState } from "react";
import { IUser } from "../../types/user";
import ModalWrapper from "../ModalWrapper";
import { searchFriendsApi } from "../../api/user";
import { api } from "../../api";
import { UserPlusIcon } from "@heroicons/react/24/outline";

type SearchUserModalProps = {
	userId: number;
	sendFriendRequest: (user: IUser) => void;
};

const SearchUserModalContent: React.FC<SearchUserModalProps> = ({
	userId,
	sendFriendRequest,
}) => {
	const [searchTerm, setSearchTerm] = useState("");
	const [searchResults, setSearchResults] = useState<IUser[]>([]);

	useEffect(() => {
		if (searchTerm.trim() === "") {
			setSearchResults([]);
			return;
		}

		const fetchUsers = async () => {
			const res = await searchFriendsApi(api, userId, searchTerm);
			setSearchResults(res.data.data);
		};

		// Debounce search input
		const debounceTimeout = setTimeout(fetchUsers, 300);
		return () => clearTimeout(debounceTimeout);
	}, [userId, searchTerm]);

	return (
		<div className="w-full flex flex-col gap-3">
			<h3 className="text-center text-3xl font-bold text-secondary">
				Add Friends
			</h3>

			<input
				type="text"
				placeholder="Search users..."
				className="bg-white text-black px-2 py-1 rounded-lg shadow-lg"
				value={searchTerm}
				onChange={(e) => setSearchTerm(e.target.value)}
			/>

			<div className="max-h-40 overflow-y-auto flex flex-col gap-2">
				{searchResults.length > 0 ? (
					searchResults.map((user) => (
						<div
							key={user.id}
							className="flex justify-between items-center bg-white px-3 py-2 rounded-lg shadow-md"
						>
							<div className="flex items-center gap-2">
								<img
									src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
									alt="Profile Image"
									className="w-7 h-7 rounded-full"
								/>
								<p className="text-secondary">{user.username}</p>
							</div>
							<UserPlusIcon
								onClick={() => sendFriendRequest(user)}
								className="h-6 w-6 text-green-500 hover:text-green-600 cursor-pointer"
							/>
						</div>
					))
				) : (
					<p className="text-gray-500 text-center">No users found</p>
				)}
			</div>
		</div>
	);
};

const SearchUserModal = ModalWrapper(SearchUserModalContent);

export default SearchUserModal;
