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
import { fetchRoomApi, fetchRoomAttendeesApi, joinRoomApi } from "../api/room";
import { api } from "../api";
import { useNavigate, useSearchParams } from "react-router-dom";
import { IRoom } from "../types/room";
import { useUserCtx } from "../context/user";
import { IUser } from "../types/user";
import { formatDate, toDayOfWeek } from "../utils/date";
import { useRoomCtx } from "../context/room";
import { channelTypes, useWs } from "../context/ws";
import useMandatoryParam from "../hooks/useMandatoryParam";
import InviteAttendeesModal from "../components/modals/InviteAttendeesModal";
import QRCodeModal from "../components/modals/QRCodeModal";
import { getRedirectPath, setRedirectPath } from "../utils/redirect";
import ConfirmJoinModal from "../components/modals/ConfirmJoinModal";
import { useToast } from "../context/toast";
import { AxiosError } from "axios";
import { isRoomBillConsolidatedApi } from "../api/bill";

const initialRoom: IRoom = {
  id: "0",
  name: "Room",
  date: "",
  time: "",
  venue: "",
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
  url: "",
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
  const { closeRoom } = useRoomCtx();
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
      stopLoading();

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
      stopLoading();
    }
  };

  const handleCloseRoom = async () => {
    startLoading();
    const res = await closeRoom(roomId);

    if (!res.isSuccessResponse) {
      console.error("Failed to close room", res.error);
      return;
    }

    stopLoading();
    navigate("/");
  };

  const fetchData = async () => {
    const fetchRoom = async (roomId: string) => {
      const res = await fetchRoomApi(api, roomId);
      console.log("[RoomPage] Room data", res.data.data);
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
    <div className="h-full flex flex-col items-center gap-3 bg-gray-200">
      <RoomTopBar title={room.name} shouldCenterTitle={true} />

      <RoomDetails room={room} />

      <RoomAttendees
        isHost={user.id === room.hostId}
        attendees={attendees}
        roomId={roomId}
        hostId={room.hostId}
      />

      <RoomActionWidgets
        userId={user.id}
        roomId={roomId}
        room={room}
        attendees={attendees}
        isHost={user.id === room.hostId}
        isRoomBillConsolidated={isRoomBillConsolidated}
        numNewMessages={numNewMessages}
        onCloseRoom={handleCloseRoom}
      />

      <ConfirmJoinModal
        isVisible={isConfirmJoinModalVisible}
        closeModal={() => setConfirmJoinModalVisible(false)}
        rejectJoin={() => navigate("/")}
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

type RoomAttendeesProps = {
  roomId: string;
  isHost: boolean;
  hostId: number;
  attendees: IUser[];
};

const RoomAttendees: React.FC<RoomAttendeesProps> = ({
  roomId,
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
								cursor-pointer hover:border hover:border-white"
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
              pictureUrl={attendee.pictureUrl}
              isHost={attendee.id === hostId}
            />
          ))}
        </div>
      </div>

      <InviteAttendeesModal
        isVisible={isModalVisible}
        closeModal={() => setIsModalVisible(false)}
        roomId={roomId}
      />
    </>
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
}) => {
  const [isQRCodeModalVisible, setIsQRCodeModalVisible] = useState(false);
  const showSplitBill = isHost
    ? isRoomBillConsolidated
      ? false
      : true
    : false;

  return (
    <>
      <div className="w-full mt-3 h-[10%] flex justify-evenly items-baseline">
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
        <ButtonCard
          title="Generate QR"
          Icon={QrCodeIcon}
          onClick={() => setIsQRCodeModalVisible(true)}
          isLink={false}
        />

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
            onClick={() => {}}
            isLink={false}
          />
        )}

        <QRCodeModal
          url={window.location.href + "?join=true"}
          isVisible={isQRCodeModalVisible}
          closeModal={() => setIsQRCodeModalVisible(false)}
        />
      </div>
    </>
  );
};

export default RoomPage;
