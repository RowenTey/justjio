/* eslint-disable react-hooks/exhaustive-deps */
import React, { useEffect, useState } from "react";
import RoomTopBar from "../components/RoomTopBar";
import { fetchRoomInvitesApi } from "../api/room";
import { api } from "../api";
import { IRoomInvite } from "../types/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { useRoomCtx } from "../context/room";
import { formatDate } from "../utils/date";
import { CalendarIcon, MapPinIcon } from "@heroicons/react/24/solid";
import { ClockIcon } from "@heroicons/react/24/outline";

const RoomInvitesPage: React.FC = () => {
	const { loading, startLoading, stopLoading } = useLoadingAndError();
	// TODO: refactor to context
	const [invites, setInvites] = useState<IRoomInvite[]>([]);

	const { respondToInvite } = useRoomCtx();

	const onRespondToInvite = async (roomId: string, accept: boolean) => {
		const res = await respondToInvite(roomId, accept);

		// TODO: implement error handling
		if (!res.isSuccessResponse) {
			console.error(res.error);
			return;
		}

		setInvites((prevInvites) =>
			prevInvites.filter((invite) => invite.roomId !== roomId)
		);
	};

	useEffect(() => {
		const fetchInvites = async () => {
			return fetchRoomInvitesApi(api);
		};

		startLoading();
		fetchInvites()
			.then((res) => setInvites(res.data.data))
			.then(stopLoading);
	}, []);

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<RoomTopBar title="Room Invites" />

			{loading ? (
				<Spinner spinnerSize={{ width: "w-16", height: "h-16" }} />
			) : (
				<RoomInvites invites={invites} handleInvite={onRespondToInvite} />
			)}
		</div>
	);
};

const RoomInvites: React.FC<{
	invites: IRoomInvite[];
	handleInvite: (roomId: string, accept: boolean) => void;
}> = ({ invites, handleInvite }) => {
	return (
		<div className="w-full flex flex-col pb-4 items-center gap-4 overflow-y-auto">
			{invites.map((invite) => (
				<RoomInviteCard
					key={invite.id}
					invite={invite}
					handleInvite={handleInvite}
				/>
			))}
		</div>
	);
};

const RoomInviteCard: React.FC<{
	invite: IRoomInvite;
	handleInvite: (roomId: string, accept: boolean) => void;
}> = ({ invite, handleInvite }) => {
	const onActionClick = (roomId: string, accept: boolean) => () => {
		handleInvite(roomId, accept);
	};

	return (
		<div className="flex flex-col gap-2 w-[75%] bg-white rounded-md shadow-md px-4 py-3 text-black border-2 border-justjio-secondary">
			<h3 className="text-lg font-bold">{invite.room.name}</h3>

			<div className="flex justify-between items-center">
				<div className="flex flex-col">
					<div className="leading-snug">
						<div className="flex items-center gap-2 mb-2">
							<img
								src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
								alt=""
								className="w-8 h-8 rounded-full"
							/>
							<p>{invite.room.host.username}</p>
						</div>
						<div className="flex items-center gap-2">
							<CalendarIcon className="w-6 h-6 text-justjio-secondary" />
							<p>{formatDate(invite.room.date)}</p>
						</div>
						<div className="flex items-center gap-2">
							<ClockIcon className="w-6 h-6 text-justjio-secondary" />
							<p>{invite.room.time}</p>
						</div>
						<div className="flex items-center gap-2">
							<MapPinIcon className="w-6 h-6 text-justjio-secondary" />
							<p>{invite.room.venue}</p>
						</div>
						{invite.message && <p className="mt-2">{invite.message}</p>}
					</div>
				</div>

				<div className="flex flex-col gap-4 mt-3 text-white">
					<button
						className="bg-green-600 hover:bg-green-500 rounded-full px-3 py-1"
						onClick={onActionClick(invite.roomId, true)}
					>
						Accept
					</button>
					<button
						className="bg-red-600 hover:bg-red-500 rounded-full px-3 py-1"
						onClick={onActionClick(invite.roomId, false)}
					>
						Decline
					</button>
				</div>
			</div>
		</div>
	);
};

export default RoomInvitesPage;
