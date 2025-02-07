/* eslint-disable react-hooks/exhaustive-deps */
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import ButtonCard from "../components/ButtonCard";
import {
	ChatBubbleLeftIcon,
	PlusIcon,
	DocumentDuplicateIcon,
	DocumentPlusIcon,
	XMarkIcon,
	QrCodeIcon,
	ArrowRightStartOnRectangleIcon,
} from "@heroicons/react/24/outline";
import PeopleBox from "../components/PeopleBox";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useEffect, useState } from "react";
import Spinner from "../components/Spinner";
import { fetchRoomApi, fetchRoomAttendeesApi } from "../api/room";
import { api } from "../api";
import { useNavigate } from "react-router-dom";
import { IRoom } from "../types/room";
import { useUserCtx } from "../context/user";
import { IUser } from "../types/user";
import { formatDate, toDayOfWeek } from "../utils/date";
import { useRoomCtx } from "../context/room";
import { channelTypes, useWs } from "../context/ws";
import useMandatoryParam from "../hooks/useMandatoryParam";
import InviteAttendeesModal from "../components/modals/InviteAttendeesModal";

const RoomPage = () => {
	const { loading, startLoading, stopLoading } = useLoadingAndError();
	const [room, setRoom] = useState<IRoom | undefined>(undefined);
	const [attendees, setAttendees] = useState<IUser[]>([]);
	const [numNewMessages, setNumNewMessages] = useState<number>(0);
	const roomId = useMandatoryParam("roomId");
	const { closeRoom } = useRoomCtx();
	const [subscribe, unsubscribe] = useWs();
	const { user } = useUserCtx();
	const navigate = useNavigate();

	const onCloseRoom = async () => {
		startLoading();
		const res = await closeRoom(roomId);

		if (!res.isSuccessResponse) {
			console.error("Failed to close room", res.error);
			return;
		}

		stopLoading();
		navigate("/");
	};

	useEffect(() => {
		const fetchRoom = async (roomId: string) => {
			const res = await fetchRoomApi(api, roomId);
			console.log("[RoomPage] Room data", res.data.data);
			setRoom(res.data.data);
		};
		const fetchAttendees = async (roomId: string) => {
			const res = await fetchRoomAttendeesApi(api, roomId);
			setAttendees(res.data.data);
		};

		startLoading();
		Promise.all([fetchRoom(roomId), fetchAttendees(roomId)]).then(stopLoading);
	}, []);

	useEffect(() => {
		const channel = channelTypes.createMessage();

		subscribe(channel, (message) => {
			console.log("[RoomPage] Received message", message);
			setNumNewMessages((numNewMessages) => numNewMessages + 1);
		});

		return () => {
			unsubscribe(channel);
		};
	}, [roomId, subscribe, unsubscribe]);

	if (loading || room === undefined) {
		return <Spinner bgClass="bg-primary" />;
	}

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<RoomTopBar title={room.name} shouldCenterTitle={true} />

			<RoomDetails room={room} />

			<RoomAttendees
				isHost={user.id === room.hostId}
				attendees={attendees}
				hostId={room.hostId}
			/>

			<RoomActionWidgets
				isHost={user.id === room.hostId}
				roomId={roomId}
				numNewMessages={numNewMessages}
				onSplitBillClicked={() =>
					navigate(`/room/${roomId}/bill/split`, {
						state: { roomName: room.name },
					})
				}
				onCreateBillClicked={() =>
					navigate(`/room/${roomId}/bill/create`, {
						state: { attendees, roomName: room.name, currentUserId: user.id },
					})
				}
				onChatClicked={() => navigate(`/room/${roomId}/chat`)}
				onCloseRoom={onCloseRoom}
			/>
		</div>
	);
};

const RoomDetails: React.FC<{ room: IRoom }> = ({ room }) => {
	return (
		<div className="h-[25%] w-full px-5 flex flex-col gap-2">
			<h3 className="text-secondary font-bold">
				{new Date(room.date) > new Date() ? "Upcoming" : "Passed"} Event
			</h3>

			<div className="h-[90%] flex justify-between bg-white gap-6 rounded-lg px-3 py-2 leading-tight">
				<div className="w-2/5 flex flex-col gap-2 justify-center font-bold text-black">
					<div className="flex flex-col">
						<p>{toDayOfWeek(room.date)}</p>
						<p>{formatDate(room.date)}</p>
						<p>{room.time}</p>
					</div>
				</div>

				<div className="w-3/5 flex flex-col gap-2 font-bold justify-center">
					<div className="w-full py-2 px-3 bg-secondary rounded-xl text-white">
						<p>Venue: {room.venue}</p>
					</div>
					<div className="w-full py-2 px-3 bg-secondary rounded-xl text-white">
						<p>Attendees: {room.attendeesCount}</p>
					</div>
				</div>
			</div>
		</div>
	);
};

interface RoomAttendeesProps {
	isHost: boolean;
	hostId: number;
	attendees: IUser[];
}

const RoomAttendees: React.FC<RoomAttendeesProps> = ({
	isHost,
	hostId,
	attendees,
}) => {
	const [isModalVisible, setIsModalVisible] = useState(false);

	return (
		<>
			<div className="h-[40%] w-full px-5 flex flex-col gap-2 mt-1">
				<div className="w-full flex justify-between items-center">
					<h3 className="text-secondary font-bold">Attendees</h3>

					{isHost && (
						<div
							className="flex items-center justify-center 
								rounded-full bg-secondary p-1 w-8 h-8 
								cursor-pointer hover:border-2 hover:border-white hover:shadow-lg"
							onClick={() => setIsModalVisible(true)}
						>
							<PlusIcon className="h-7 w-7 text-white" />
						</div>
					)}
				</div>
				<div
					className="h-[90%] flex flex-col gap-2 p-2 
						rounded-xl bg-primary overflow-y-auto"
				>
					{attendees.map((attendee) => (
						<PeopleBox
							key={attendee.id}
							name={attendee.username}
							isHost={attendee.id === hostId}
						/>
					))}
				</div>
			</div>

			<InviteAttendeesModal
				isVisible={isModalVisible}
				closeModal={() => setIsModalVisible(false)}
			/>
		</>
	);
};

interface RoomActionWidgetsProps {
	isHost: boolean;
	roomId: string;
	numNewMessages: number;
	onSplitBillClicked?: () => void;
	onCreateBillClicked: () => void;
	onChatClicked: () => void;
	onCloseRoom: () => void;
}

const RoomActionWidgets: React.FC<RoomActionWidgetsProps> = ({
	isHost,
	numNewMessages,
	onSplitBillClicked,
	onCreateBillClicked,
	onChatClicked,
	onCloseRoom,
}) => {
	const showSplitBillBtn = isHost && onSplitBillClicked !== undefined;

	return (
		<>
			<div className="w-full mt-3 h-[10%] flex justify-evenly items-baseline">
				{showSplitBillBtn && (
					<ButtonCard
						title="Split Bill"
						Icon={DocumentDuplicateIcon}
						onClick={onSplitBillClicked}
					/>
				)}
				<ButtonCard
					title="Create Bill"
					Icon={DocumentPlusIcon}
					onClick={onCreateBillClicked}
				/>
				<ButtonCard
					title="Chat"
					Icon={ChatBubbleLeftIcon}
					numNotifications={numNewMessages}
					onClick={onChatClicked}
				/>
				<ButtonCard title="Generate QR" Icon={QrCodeIcon} onClick={() => {}} />

				{/* TODO: Show prompt for close and leave room */}
				{isHost ? (
					<ButtonCard
						title="Close Room"
						Icon={XMarkIcon}
						onClick={onCloseRoom}
					/>
				) : (
					<ButtonCard
						title="Leave Room"
						Icon={ArrowRightStartOnRectangleIcon}
						onClick={() => {}}
					/>
				)}
			</div>
		</>
	);
};

export default RoomPage;
