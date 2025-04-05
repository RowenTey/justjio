/* eslint-disable react-hooks/exhaustive-deps */
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/auth";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useUserCtx } from "../context/user";
import HomeTopBar from "../components/top-bar/HomeTopBar";
import RoomCard from "../components/RoomCard";
import ButtonCard from "../components/ButtonCard";
import { EnvelopeIcon, PlusIcon } from "@heroicons/react/24/outline";
import TransactionContainer from "../components/TransactionContainer";
import { useRoomCtx } from "../context/room";
import { useEffect, useState } from "react";
import { IRoom } from "../types/room";
import Spinner from "../components/Spinner";
import { useTransactionCtx } from "../context/transaction";
import { useSubscription } from "../context/subscriptions";

const HomePage = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [logoutLoading, setLogoutLoading] = useState(false);
  const { logout } = useAuth();
  const { unsubscribe } = useSubscription();
  const { rooms, fetchRooms } = useRoomCtx();
  const { fetchTransactions } = useTransactionCtx();
  const { user } = useUserCtx();
  const navigate = useNavigate();

  const onLogout = async () => {
    setLogoutLoading(true);
    await unsubscribe();
    await logout();
    setLogoutLoading(false);
    navigate("/login", { state: { from: "/" } });
  };

  useEffect(() => {
    async function fetchData() {
      const roomPromise = fetchRooms();
      const transactionPromise = fetchTransactions();
      return await Promise.all([roomPromise, transactionPromise]);
    }

    console.log("[HomePage] Fetching data...");
    startLoading(0);
    fetchData()
      .then(() => stopLoading(0))
      .then(() => console.log("[HomePage] Data fetched"))
      .catch(() => stopLoading(0));
  }, []);

  if (loadingStates[0]) {
    return <Spinner bgClass="bg-gray-200" />;
  }

  return (
    <div className="h-full flex flex-col items-center bg-gray-200">
      <HomeTopBar
        isLoading={logoutLoading}
        username={user.username || "guest"}
        onLogout={onLogout}
      />
      <TransactionActionWidgets />
      <RoomActionWidgets />
      <RecentRoomsWidget rooms={rooms} />
    </div>
  );
};

const TransactionActionWidgets: React.FC = () => {
  const { toPay, toReceive } = useTransactionCtx();

  return (
    <div className="w-[90%] h-[30%] flex justify-between mt-3">
      <TransactionContainer
        title="To Pay:"
        emptyText="No bills to pay"
        transactions={toPay}
        isPayer={true}
      />
      <TransactionContainer
        title="To Receive:"
        emptyText="No bills to receive"
        transactions={toReceive}
        isPayer={false}
      />
    </div>
  );
};

const RoomActionWidgets: React.FC = () => {
  return (
    <div className="h-[15%] flex justify-evenly w-full mt-4">
      <ButtonCard
        title="Create Room"
        Icon={PlusIcon}
        isLink={true}
        linkProps={{
          to: "/rooms/create",
          from: "/",
        }}
      />
      <ButtonCard
        title="Room Invites"
        Icon={EnvelopeIcon}
        isLink={true}
        linkProps={{
          to: "/rooms/invites",
          from: "/",
        }}
      />
    </div>
  );
};

const RecentRoomsWidget: React.FC<{ rooms: IRoom[] }> = ({ rooms }) => {
  return (
    <div className="w-full h-[60%] mt-1 flex flex-col items-center">
      <h1 className="text-secondary text-[2.5rem] font-bold">Recent Rooms</h1>
      <div
        className={`relative bottom-0 h-[95%] w-full px-[1.875rem] flex items-center gap-7 overflow-x-auto ${
          rooms.length === 1 ? "justify-center" : ""
        }`}
      >
        {rooms.length > 0 ? (
          rooms.map((room) => (
            <RoomCard key={room.id} id={room.id} name={room.name} />
          ))
        ) : (
          <p className="text-gray-500 text-2xl font-bold text-center">
            No rooms to display. Go create or join one!
          </p>
        )}
      </div>
    </div>
  );
};

export default HomePage;
