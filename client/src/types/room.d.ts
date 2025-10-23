import { BaseContextResponse } from ".";
import type { CreateRoomRequest, RoomListDto } from "./models";

export interface IVenue {
  name: string;
  address: string;
  googleMapsPlaceId: string;
}

export type RoomCtxState = {
  rooms: RoomListDto[];
};

export interface RoomContextType {
  rooms: RoomListDto[];
  fetchRooms: () => Promise<BaseContextResponse>;
  createRoom: (data: CreateRoomRequest) => Promise<BaseContextResponse>;
  respondToInvite: (
    roomId: string,
    accept: boolean,
  ) => Promise<BaseContextResponse>;
  closeRoom: (roomId: string) => Promise<BaseContextResponse>;
  leaveRoom: (roomId: string) => Promise<BaseContextResponse>;
}

export type RoomActionTypes =
  | {
      type:
        | "FETCH_ROOMS"
        | "CREATE_ROOM"
        | "JOIN_ROOM"
        | "CLOSE_ROOM"
        | "LEAVE_ROOM";
      payload: RoomListDto[];
    }
  | { type: "DECLINE_ROOM"; payload?: never }
  | { type: "LOGOUT"; payload?: never };
