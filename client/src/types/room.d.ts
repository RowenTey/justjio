import { BaseContextResponse } from ".";
import { BaseUserInfo } from "./user";

export interface IRoom {
  id: string;
  name: string;
  time: string;
  venue: string;
  venueUrl: string;
  date: string;
  imageUrl: string;
  hostId: number;
  host: BaseUserInfo;
  description?: string;
  attendeesCount: number;
  url: string;
  createdAt: string;
  updatedAt: string;
  isPrivate: boolean;
  isClosed: boolean;
}

export interface IRoomInvite {
  id: number;
  roomId: string;
  message: string;
  room: IRoom;
}

export interface IVenue {
  name: string;
  address: string;
  googleMapsPlaceId: string;
}

export interface RoomState {
  rooms: IRoom[];
}

export interface RoomContextType {
  rooms: IRoom[];
  fetchRooms: () => Promise<BaseContextResponse>;
  createRoom: (
    roomData: Partial<IRoom>,
    placeId: string,
    attendeesId: string[],
    message?: string,
  ) => Promise<BaseContextResponse>;
  respondToInvite: (
    roomId: string,
    accept: boolean,
  ) => Promise<BaseContextResponse>;
  closeRoom: (roomId: string) => Promise<BaseContextResponse>;
  leaveRoom: (roomId: string) => Promise<BaseContextResponse>;
}

interface RoomsPayload {
  data: IRoom[];
}

type RoomActionTypes =
  | {
      type:
        | "FETCH_ROOMS"
        | "CREATE_ROOM"
        | "JOIN_ROOM"
        | "CLOSE_ROOM"
        | "LEAVE_ROOM";
      payload: RoomsPayload;
    }
  | { type: "DECLINE_ROOM"; payload?: never }
  | { type: "LOGOUT"; payload?: never };
