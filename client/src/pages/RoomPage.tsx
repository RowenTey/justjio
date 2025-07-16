/* eslint-disable react-hooks/exhaustive-deps */
import RoomTopBar from "../components/top-bar/RoomTopBar";
import ButtonCard from "../components/ButtonCard";
import {
  ChatBubbleLeftIcon,
  DocumentDuplicateIcon,
  DocumentPlusIcon,
  XMarkIcon,
  ArrowRightStartOnRectangleIcon,
  CalendarIcon,
  ClockIcon,
  MapPinIcon,
  UserPlusIcon,
  InformationCircleIcon,
} from "@heroicons/react/24/outline";
import PeopleBox from "../components/PeopleBox";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useEffect, useState } from "react";
import Spinner from "../components/Spinner";
import { fetchRoomApi, fetchRoomAttendeesApi, joinRoomApi } from "../api/room";
import { api } from "../api";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IRoom } from "../types/room";
import { useUserCtx } from "../context/user";
import { IUser } from "../types/user";
import { formatDate } from "../utils/date";
import { useRoomCtx } from "../context/room";
import { channelTypes, useWs } from "../context/ws";
import useMandatoryParam from "../hooks/useMandatoryParam";
import InviteAttendeesModal from "../components/modals/InviteAttendeesModal";
import { getRedirectPath, setRedirectPath } from "../utils/redirect";
import ConfirmJoinModal from "../components/modals/ConfirmJoinModal";
import { useToast } from "../context/toast";
import { AxiosError } from "axios";
import { isRoomBillConsolidatedApi } from "../api/bill";
import QRCodeModal from "../components/modals/QRCodeModal";

const initialRoom: IRoom = {
  id: "0",
  name: "Room",
  date: "",
  time: "",
  venue: "",
  venuePlaceId: "",
  venueUrl: "",
  imageUrl: "",
  attendeesCount: 1,
  hostId: 0,
  host: {
    id: 0,
    username: "",
    email: "",
    pictureUrl: "",
  },
  createdAt: "",
  updatedAt: "",
  isClosed: false,
  isPrivate: false,
  url: "",
  description: "",
};

const RoomPage = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [room, setRoom] = useState<IRoom>(initialRoom);
  const [attendees, setAttendees] = useState<IUser[]>([]);
  const [numNewMessages, setNumNewMessages] = useState<number>(0);
  const [isConfirmJoinModalVisible, setConfirmJoinModalVisible] =
    useState(false);
  const [isRoomBillConsolidated, setIsRoomBillConsolidated] = useState(false);
  const roomId = useMandatoryParam("roomId");
  const [subscribe, unsubscribe] = useWs();
  const { closeRoom, leaveRoom } = useRoomCtx();
  const { user } = useUserCtx();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const [searchParams, setSearchParams] = useSearchParams();

  const handleConfirmJoin = async () => {
    startLoading();
    try {
      const res = await joinRoomApi(api, roomId);
      showToast("Successfully joined room!", false);
      console.log("[RoomPage] Joined room", res.data.data);

      searchParams.delete("join");
      setSearchParams(searchParams);

      fetchData();
    } catch (error) {
      switch ((error as AxiosError).response?.status) {
        case 404:
          showToast("Room not found, please try again later.", true);
          break;
        case 409:
          showToast("User already in room!", true);
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

  const handleCloseRoom = async () => {
    startLoading();
    const res = await closeRoom(roomId);

    if (!res.isSuccessResponse) {
      switch ((res.error as AxiosError).response?.status) {
        case 403:
          showToast("You are not the host of this room!", true);
          break;
        case 404:
          showToast("Room not found!", true);
          break;
        case 409:
          showToast("Cannot close room with unconsolidated bills!", true);
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
      stopLoading();
      return;
    }

    stopLoading();
    navigate("/", { state: { from: `/room/${roomId}` } });
  };

  const handleLeaveRoom = async () => {
    startLoading();
    const res = await leaveRoom(roomId);

    if (!res.isSuccessResponse) {
      switch ((res.error as AxiosError).response?.status) {
        case 404:
          showToast("Room not found!", true);
          break;
        case 409:
          showToast("Cannot leave room with unconsolidated bills!", true);
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
      stopLoading();
      return;
    }

    stopLoading();
    navigate("/", { state: { from: `/room/${roomId}` } });
  };

  const fetchData = async () => {
    const fetchRoom = async (roomId: string) => {
      const res = await fetchRoomApi(api, roomId);
      setRoom(res.data.data);
    };

    const fetchAttendees = async (roomId: string) => {
      const res = await fetchRoomAttendeesApi(api, roomId);
      setAttendees(res.data.data);
    };

    const getBillConsolidationStatus = async (roomId: string) => {
      const res = await isRoomBillConsolidatedApi(api, roomId);
      setIsRoomBillConsolidated(res.data.data.isConsolidated);
    };

    startLoading();
    Promise.all([
      fetchRoom(roomId),
      fetchAttendees(roomId),
      getBillConsolidationStatus(roomId),
    ])
      .then(() => stopLoading())
      .catch(() => stopLoading());
  };

  useEffect(() => {
    searchParams.get("join") === "true" && setConfirmJoinModalVisible(true);

    // Clear the redirect path after successful login/signup
    if (getRedirectPath() === window.location.pathname) {
      setRedirectPath("");
    }

    fetchData();
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

  if (loadingStates[0]) {
    return <Spinner bgClass="bg-primary" />;
  }

  return (
    <div className="h-full flex flex-col items-center gap-1 bg-gray-200">
      <RoomTopBar room={room} showEditBtn={user.id === room.hostId} />

      <RoomDetails room={room} />

      <RoomAttendees attendees={attendees} hostId={room.hostId} />

      <RoomActionWidgets
        userId={user.id}
        roomId={roomId}
        room={room}
        attendees={attendees}
        isHost={user.id === room.hostId}
        isRoomBillConsolidated={isRoomBillConsolidated}
        numNewMessages={numNewMessages}
        onCloseRoom={handleCloseRoom}
        onLeaveRoom={handleLeaveRoom}
      />

      <ConfirmJoinModal
        isVisible={isConfirmJoinModalVisible}
        closeModal={() => setConfirmJoinModalVisible(false)}
        rejectJoin={() => navigate("/", { state: { from: `/room/${roomId}` } })}
        confirmJoin={() => {
          handleConfirmJoin();
          setConfirmJoinModalVisible(false);
        }}
      />
    </div>
  );
};

const RoomDetails: React.FC<{ room: IRoom }> = ({ room }) => {
  return (
    <div className="h-[41%] w-full px-5 flex flex-col gap-2">
      <h3 className="text-secondary font-bold">
        {room.isPrivate ? "Private" : "Public"} Event{" "}
        <span className="font-thin italic">
          - {new Date(room.date) > new Date() ? "Upcoming" : "Passed"}
        </span>
      </h3>

      <div className="w-full h-[90%] flex flex-col items-center rounded-2xl overflow-hidden">
        <div className="w-full h-[55%] overflow-hidden -mt-1">
          <img
            src={room.imageUrl}
            alt=""
            className="w-full h-full object-cover"
          />
        </div>

        <div className="w-full h-[45%] flex flex-col justify-between bg-white rounded-b-2xl shadow-lg px-3 py-2 leading-tight text-gray-500">
          <div className="w-full flex items-center gap-6">
            <div className="flex gap-2 items-center">
              <CalendarIcon className="h-6 w-6" />
              <p>{formatDate(room.date)}</p>
            </div>
            <div className="flex gap-2 items-center">
              <ClockIcon className="h-6 w-6" />
              <p>{room.time}</p>
            </div>
          </div>
          <div className="flex gap-2 items-center ">
            <InformationCircleIcon className="h-6 w-6" />
            <p className="w-[85%] overflow-x-scroll overflow-y-hidden">
              {room.description || "N/A"}
            </p>
          </div>
          <a
            href={room.venueUrl}
            target="_blank"
            className="w-fit flex items-center gap-2 text-[#8A38F5] hover:underline"
          >
            <MapPinIcon className="w-6 h-6" />
            <span>{room.venue}</span>
          </a>
        </div>
      </div>
    </div>
  );
};

type RoomAttendeesProps = {
  hostId: number;
  attendees: IUser[];
};

const RoomAttendees: React.FC<RoomAttendeesProps> = ({ hostId, attendees }) => {
  return (
    <div className="h-[33%] w-full px-5 flex flex-col gap-2">
      <h3 className="text-secondary font-bold">
        {attendees.length} Attendee(s)
      </h3>
      <div
        className="h-[90%] flex flex-col gap-2 p-2 
						rounded-xl bg-primary overflow-y-auto"
      >
        {attendees.map((attendee) => (
          <PeopleBox
            key={attendee.id}
            name={attendee.username}
            pictureUrl={attendee.pictureUrl}
            isHost={attendee.id === hostId}
          />
        ))}
      </div>
    </div>
  );
};

type RoomActionWidgetsProps = {
  userId: number;
  roomId: string;
  room: IRoom;
  isHost: boolean;
  isRoomBillConsolidated: boolean;
  attendees: IUser[];
  numNewMessages: number;
  onCloseRoom: () => void;
  onLeaveRoom: () => void;
};

const RoomActionWidgets: React.FC<RoomActionWidgetsProps> = ({
  userId,
  roomId,
  room,
  isHost,
  isRoomBillConsolidated,
  attendees,
  numNewMessages,
  onCloseRoom,
  onLeaveRoom,
}) => {
  const [isInviteModalVisible, setIsInviteModalVisible] = useState(false);
  const [isQRCodeModalVisible, setIsQRCodeModalVisible] = useState(false);
  const showSplitBill = isHost ? !isRoomBillConsolidated : false;
  const showInviteFriends = room.isPrivate ? isHost : true;

  return (
    <div className="w-full mt-1 h-[10%] flex justify-evenly items-baseline">
      {showSplitBill && (
        <ButtonCard
          title="Split Bill"
          Icon={DocumentDuplicateIcon}
          isLink={true}
          linkProps={{
            to: `/room/${roomId}/bill/split`,
            from: `/room/${roomId}`,
            state: { roomName: room.name },
          }}
        />
      )}
      {!isRoomBillConsolidated && (
        <ButtonCard
          title="Create Bill"
          Icon={DocumentPlusIcon}
          isLink={true}
          linkProps={{
            to: `/room/${roomId}/bill/create`,
            from: `/room/${roomId}`,
            state: { attendees, roomName: room.name, currentUserId: userId },
          }}
        />
      )}

      <ButtonCard
        title="Chat"
        Icon={ChatBubbleLeftIcon}
        numNotifications={numNewMessages}
        linkProps={{
          to: `/room/${roomId}/chat`,
          from: `/room/${roomId}`,
        }}
      />
      {showInviteFriends && (
        <ButtonCard
          title="Invite Friends"
          Icon={UserPlusIcon}
          onClick={() => setIsInviteModalVisible(true)}
          isLink={false}
        />
      )}

      {/* TODO: Show prompt for close and leave room */}
      {isHost ? (
        <ButtonCard
          title="Close Room"
          Icon={XMarkIcon}
          onClick={onCloseRoom}
          isLink={false}
        />
      ) : (
        <ButtonCard
          title="Leave Room"
          Icon={ArrowRightStartOnRectangleIcon}
          onClick={onLeaveRoom}
          isLink={false}
        />
      )}

      <InviteAttendeesModal
        isVisible={isInviteModalVisible}
        setIsQRCodeModalVisible={setIsQRCodeModalVisible}
        closeModal={() => setIsInviteModalVisible(false)}
        roomId={roomId}
      />

      <QRCodeModal
        url={window.location.href + "?join=true"}
        isVisible={isQRCodeModalVisible}
        closeModal={() => setIsQRCodeModalVisible(false)}
      />
    </div>
  );
};

export default RoomPage;
