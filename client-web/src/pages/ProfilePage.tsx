/* eslint-disable react-hooks/exhaustive-deps */
import React, { useEffect, useMemo, useState } from "react";
import { useUserCtx } from "../context/user";
import { getNumFriendsApi } from "../api/user";
import { api } from "../api";
import { fetchNumRoomsApi } from "../api/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { fetchTransactionsApi } from "../api/transaction";
import { ITransaction } from "../types/transaction";
import { useNavigate } from "react-router-dom";

const ProfilePage: React.FC = () => {
	const { loading, startLoading, stopLoading } = useLoadingAndError();
	const { user } = useUserCtx();
	const [numFriends, setNumFriends] = useState<number | undefined>(undefined);
	const [numRooms, setNumRooms] = useState<number | undefined>(undefined);
	const [transactions, setTransactions] = useState<ITransaction[]>([]);

	const groupedTransactions = useMemo(() => {
		return transactions.reduce((acc, tx) => {
			const date = new Date(tx.paidOn);
			const key = `${date.getDate()} ${date.toLocaleString("default", {
				month: "short",
			})}`;

			if (!acc[key]) {
				acc[key] = [];
			}
			acc[key].push(tx);
			return acc;
		}, {} as Record<string, ITransaction[]>);
	}, [transactions]);

	useEffect(() => {
		const fetchData = async () => {
			const numFriendsPromise = getNumFriendsApi(api, user.id);
			const numRooomsPromise = fetchNumRoomsApi(api);
			const fetchTransactionsPromise = fetchTransactionsApi(api, true);
			return Promise.all([
				numFriendsPromise,
				numRooomsPromise,
				fetchTransactionsPromise,
			]);
		};

		startLoading();
		fetchData()
			.then((res) => {
				setNumFriends(res[0].data.data.numFriends);
				setNumRooms(res[1].data.data.count);
				setTransactions(res[2].data.data);
			})
			.then(() => stopLoading());
	}, [user.id]);

	if (loading) {
		return <Spinner bgClass="bg-gray-200" />;
	}

	return (
		<div className="h-full flex flex-col items-center bg-gray-200">
			<ProfileTopBar />
			<ProfileContainer
				username={user.username}
				numFriends={numFriends || 0}
				numRooms={numRooms || 0}
			/>
			<TransactionHistory
				username={user.username}
				transactions={groupedTransactions}
			/>
		</div>
	);
};

const ProfileTopBar: React.FC = () => {
	return (
		<div className="relative top-0 flex h-[8%] items-center justify-center w-full py-4 px-6 bg-purple-200">
			<h1 className="text-xl font-bold text-secondary">Profile</h1>
		</div>
	);
};

type ProfileContainerProps = {
	username: string;
	numFriends: number;
	numRooms: number;
};

const ProfileContainer: React.FC<ProfileContainerProps> = ({
	username,
	numFriends,
	numRooms,
}) => {
	const navigate = useNavigate();

	return (
		<div className="w-[90%] p-4 flex justify-center items-center bg-white rounded-xl mt-4 shadow-xl">
			<div
				className="w-1/5 flex flex-col gap-3 items-center cursor-pointer"
				onClick={() => navigate("/friends")}
			>
				<p className="text-4xl font-extrabold text-secondary hover:scale-110">
					{numFriends}
				</p>
				<p className="text-lg font-semibold text-black">Friends</p>
			</div>

			<div className="w-3/5 flex flex-col gap-2 items-center">
				<img
					src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
					alt=""
					className="w-24 h-24 rounded-full"
				/>

				<div className="flex justify-center items-center bg-secondary border-black border-[1.5px] rounded-3xl px-2 w-[75%]">
					<h3 className="text-lg text-white font-semibold">{username}</h3>
				</div>

				<button className="bg-primary text-black text-sm font-semibold px-4 py-1 rounded-3xl mt-1">
					Edit Profile
				</button>
			</div>

			<div
				className="w-1/5 flex flex-col gap-3 items-center cursor-pointer"
				onClick={() => navigate("/rooms")}
			>
				<p className="text-4xl font-extrabold text-secondary hover:scale-110">
					{numRooms}
				</p>
				<p className="text-lg font-semibold text-black">Rooms</p>
			</div>
		</div>
	);
};

type TransactionHistoryProps = {
	username: string;
	transactions: Record<string, ITransaction[]>;
};

const TransactionHistory: React.FC<TransactionHistoryProps> = ({
	username,
	transactions,
}) => {
	return (
		<div className="w-[95%] h-[60%] p-4 flex flex-col justify-center mt-1 text-black">
			<p className="text-md font-semibold mb-1">Transaction History</p>

			<div
				className={`w-full h-[95%] overflow-y-auto flex flex-col gap-2 py-1 px-1 bg-primary shadow-xl rounded-md ${
					Object.keys(transactions).length === 0
						? "items-center justify-center"
						: ""
				}`}
			>
				{Object.keys(transactions).length === 0 ? (
					<p className="text-center text-gray-500">
						No transaction history available
					</p>
				) : (
					Object.entries(transactions)
						.sort((a, b) => {
							const dateA = new Date(a[1][0].paidOn);
							const dateB = new Date(b[1][0].paidOn);
							return dateB.getTime() - dateA.getTime();
						})
						.map(([date, transactionsForDate]) => (
							<div key={date} className="flex flex-col gap-1 pl-2">
								<p className="text-sm font-medium ml-1">{date}</p>
								<div className="w-full flex flex-col items-center pr-1">
									{transactionsForDate.map((transaction, index) => (
										<div
											key={index}
											className={`w-full flex justify-between items-center bg-white px-3 py-1 ${
												transactionsForDate.length === 1
													? "rounded-lg" // Single item: fully rounded
													: index === 0
													? "rounded-t-lg border-b-[1px] border-gray-200" // First item
													: index === transactionsForDate.length - 1
													? "rounded-b-lg" // Last item
													: "border-b-[1px] border-gray-200" // Middle items
											}`}
										>
											<p>
												{transaction.payee.username === username
													? transaction.payer.username
													: transaction.payee.username}
											</p>
											<p
												className={`${
													transaction.payee.username === username
														? "text-success"
														: "text-error"
												}`}
											>
												${transaction.amount.toFixed(2)}
											</p>
										</div>
									))}
								</div>
							</div>
						))
				)}
			</div>
		</div>
	);
};

export default ProfilePage;
