/* eslint-disable react-hooks/exhaustive-deps */
import React, { useEffect, useState } from "react";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import {
  fetchExploreMoreRoomsApi,
  fetchRoomInvitesApi,
  joinRoomApi,
} from "../api/room";
import { api } from "../api";
import { IRoom, IRoomInvite } from "../types/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import { useRoomCtx } from "../context/room";
import { formatDate } from "../utils/date";
import {
  ClockIcon,
  EnvelopeOpenIcon,
  MapPinIcon,
  SignalSlashIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import { useToast } from "../context/toast";
import { CalendarIcon } from "@heroicons/react/24/outline";
import { AxiosError } from "axios";

const RoomInvitesPage: React.FC = () => {
  const { loadingStates, startLoading, stopLoading } = useLoadingAndError();
  const [invites, setInvites] = useState<IRoomInvite[]>([]);
  const [exploreMoreRooms, setExploreMoreRooms] = useState<IRoom[]>([]);
  const { showToast } = useToast();

  const { respondToInvite } = useRoomCtx();

  const onRespondToInvite = async (roomId: string, accept: boolean) => {
    const res = await respondToInvite(roomId, accept);

    if (!res.isSuccessResponse) {
      console.error("An error occured while responding to invite: ", res.error);
      switch (res.error?.response?.status) {
        case 400:
          showToast("Bad request, please check request body.", true);
          break;
        case 404:
          showToast("Room not found, please try again later.", true);
          break;
        case 500:
        default:
          showToast("An error occurred, please try again later.", true);
          break;
      }
      return;
    }

    setInvites((prevInvites) =>
      prevInvites.filter((invite) => invite.roomId !== roomId),
    );
    showToast("Invite responded successfully!", false);
  };

  const onJoinRoom = async (roomId: string) => {
    startLoading();
    try {
      await joinRoomApi(api, roomId);
      setExploreMoreRooms((prevRooms) =>
        prevRooms.filter((room) => room.id !== roomId),
      );
      showToast("Successfully joined room!", false);
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

  useEffect(() => {
    const fetchData = async () => {
      startLoading();
      try {
        const [invitesResponse, exploreRoomsResponse] = await Promise.all([
          fetchRoomInvitesApi(api),
          fetchExploreMoreRoomsApi(api),
        ]);

        setInvites(invitesResponse.data.data);
        setExploreMoreRooms(exploreRoomsResponse.data.data);
      } catch (error) {
        console.error("Error fetching data:", error);
      } finally {
        stopLoading();
      }
    };

    fetchData();
  }, []);

  return (
    <div className="h-full flex flex-col items-center gap-4 bg-gray-200">
      <RoomTopBar title="Room Invites" />

      {loadingStates[0] ? (
        <Spinner spinnerSize={{ width: "w-16", height: "h-16" }} />
      ) : (
        <div className="w-full overflow-y-auto">
          <div className="w-full">
            <RoomInvites invites={invites} handleInvite={onRespondToInvite} />
          </div>

          <h2 className="text-xl font-bold text-secondary ml-5 mb-4">
            Explore More
          </h2>
          <div className="w-full">
            <ExploreMoreSection
              rooms={exploreMoreRooms}
              handleJoin={onJoinRoom}
            />
          </div>
        </div>
      )}
    </div>
  );
};

const RoomInvites: React.FC<{
  invites: IRoomInvite[];
  handleInvite: (roomId: string, accept: boolean) => void;
}> = ({ invites, handleInvite }) => {
  return (
    <div
      className={`w-full h-full flex flex-col pb-4 items-center gap-4 overflow-y-auto ${
        invites.length === 0 ? "justify-center" : ""
      }`}
    >
      {invites.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-full">
          <EnvelopeOpenIcon
            strokeWidth={0.5}
            className="w-24 h-24 text-gray-500"
          />
          <p className="text-lg font-semibold text-gray-500">
            You do not have any room invites now
          </p>
        </div>
      ) : (
        invites.map((invite) => (
          <RoomInviteCard
            key={invite.id}
            invite={invite}
            isPrivate={invite.room.isPrivate}
            handleAction={handleInvite}
            isExploreMore={false}
          />
        ))
      )}
    </div>
  );
};

const ExploreMoreSection: React.FC<{
  rooms: IRoom[];
  handleJoin: (roomId: string) => void;
}> = ({ rooms, handleJoin }) => {
  return (
    <div
      className={`w-full h-full flex flex-col pb-4 items-center gap-4 overflow-y-auto ${
        rooms.length === 0 ? "justify-center" : ""
      }`}
    >
      {rooms.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-full">
          <SignalSlashIcon
            strokeWidth={0.5}
            className="w-24 h-24 text-gray-500"
          />
          <p className="text-lg font-semibold text-gray-500">
            Nothing to explore at the moment
          </p>
        </div>
      ) : (
        rooms.map((room, idx) => {
          const invite: IRoomInvite = {
            id: -1,
            roomId: room.id,
            room,
          };

          return (
            <RoomInviteCard
              key={idx}
              invite={invite}
              isPrivate={false}
              handleAction={handleJoin}
              isExploreMore={true}
            />
          );
        })
      )}
    </div>
  );
};

const RoomInviteCard: React.FC<{
  invite: IRoomInvite;
  isPrivate: boolean;
  handleAction: (roomId: string, accept: boolean) => void;
  isExploreMore: boolean;
}> = ({ invite, isPrivate, handleAction, isExploreMore }) => {
  return (
    <div className="flex flex-col w-[85%] h-48 bg-white rounded-2xl shadow-md text-black overflow-hidden">
      <div className="relative h-[60%]">
        <img
          src={invite.room.imageUrl}
          alt=""
          className="h-full w-full object-cover"
        />
        <div className="absolute inset-0 bg-black bg-opacity-45 z-0"></div>
        {isPrivate && (
          <div className="absolute top-4 -right-10 w-32 bg-primary text-secondary py-[0.15rem] text-center transform rotate-45 text-xs font-semibold shadow-md z-10">
            private
          </div>
        )}
        <div className="absolute bottom-2 left-3 text-primary flex flex-col z-10">
          <h3 className="text-2xl font-bold ">{invite.room.name}</h3>
          <div className="flex gap-2 items-center">
            <div className="flex gap-[0.3rem] items-center">
              <img
                src={invite.room.host.pictureUrl}
                alt="Host Profile Image"
                className="w-5 h-5 rounded-full border-none"
              />
              <p className="text-xs font-semibold">
                Hosted by {invite.room.host.username}
              </p>
            </div>
            <div className="flex gap-1 items-center">
              <UserCircleIcon className="w-6 h-6" />
              <p className="text-xs font-semibold">
                {invite.room.attendeesCount} attendee(s)
              </p>
            </div>
          </div>
        </div>
      </div>
      <div className="h-[40%] flex justify-between px-2 py-1 text-xs text-gray-500">
        <div className="flex flex-col gap-1 justify-center font-semibold">
          <div className="flex items-center gap-1">
            <ClockIcon className="w-4 h-4" />
            <p>{invite.room.time}</p>
          </div>
          <div className="flex items-center gap-1">
            <CalendarIcon className="w-4 h-4" />
            <p>{formatDate(invite.room.date)}</p>
          </div>
          <a
            href={invite.room.venueUrl}
            target="_blank"
            className="flex items-center gap-1 text-[#8A38F5]"
          >
            <MapPinIcon className="w-4 h-4" />
            <span className="hover:underline">{invite.room.venue}</span>
          </a>
        </div>
        {isExploreMore ? (
          <div className="flex items-center justify-center gap-2 mr-2 font-bold">
            <button
              onClick={() => handleAction(invite.roomId, true)}
              className="rounded-3xl p-1 px-3 bg-[#8A38F5] text-white"
            >
              Join
            </button>
          </div>
        ) : (
          <div className="flex flex-col justify-center gap-2 mr-2 font-bold">
            <button
              onClick={() => handleAction(invite.roomId, true)}
              className="rounded-3xl p-1 px-3 bg-[#8A38F5] text-white"
            >
              Accept
            </button>
            <button
              onClick={() => handleAction(invite.roomId, false)}
              className="rounded-3xl p-1 px-3 bg-[#D9D9D9] text-black"
            >
              Decline
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default RoomInvitesPage;
