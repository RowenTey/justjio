import { BaseContextResponse } from ".";
import { BaseUserInfo } from "./user";

export interface IRoom {
  id: string;
  name: string;
  time: string;
  venue: string;
  date: string;
  hostId: number;
  host: BaseUserInfo;
  attendeesCount: number;
  url: string;
  createdAt: string;
  updatedAt: string;
  isClosed: boolean;
}

export interface IRoomInvite {
  id: number;
  roomId: string;
  message: string;
  room: IRoom;
}

export interface RoomState {
  rooms: IRoom[];
}

export interface RoomContextType {
  rooms: IRoom[];
  fetchRooms: () => Promise<BaseContextResponse>;
  createRoom: (
    roomData: Partial<IRoom>,
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
