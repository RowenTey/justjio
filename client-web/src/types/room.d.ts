/* eslint-disable no-mixed-spaces-and-tabs */
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
	total: number;
	rooms: IRoom[];
}

export interface RoomContextType {
	total: number;
	rooms: Room[];
	fetchRooms: () => Promise<BaseContextResponse>;
	createRoom: (roomData: Partial<IRoom>) => Promise<BaseContextResponse>;
	respondToInvite: (
		roomId: string,
		accept: boolean
	) => Promise<BaseContextResponse>;
	closeRoom: (roomId: string) => Promise<BaseContextResponse>;
}

interface FetchRoomsPayload {
	data: IRoom[];
}

interface ModifyRoomsPayload {
	rooms: IRoom[];
	total: number;
}

type RoomActionTypes =
	| { type: "FETCH_ROOMS"; payload: FetchRoomsPayload }
	| {
			type: "CREATE_ROOM" | "JOIN_ROOM" | "CLOSE_ROOM";
			payload: ModifyRoomsPayload;
	  }
	| { type: "DECLINE_ROOM"; payload?: never }
	| { type: "LOGOUT"; payload?: never };
