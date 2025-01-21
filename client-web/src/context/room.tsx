import { ReactNode, createContext, useReducer } from "react";
import { IRoom, RoomContextType } from "../types/room";
import RoomReducer, { initialRoomState } from "../reducers/room";
import {
	closeRoomApi,
	createRoomApi,
	fetchRecentRoomsApi,
	respondToInviteApi,
} from "../api/room";
import { api } from "../api";
import useContextWrapper from "../hooks/useContextWrapper";

interface RoomProviderProps {
	children: ReactNode;
}

export const CLOSE_ROOM = "CLOSE_ROOM";
export const CREATE_ROOM = "CREATE_ROOM";
export const DECLINE_ROOM = "DECLINE_ROOM";
export const FETCH_ROOMS = "FETCH_ROOMS";
export const JOIN_ROOM = "JOIN_ROOM";

const RoomContext = createContext<RoomContextType | null>(null);
const { Provider } = RoomContext;

const RoomProvider: React.FC<RoomProviderProps> = ({ children }) => {
	const [state, dispatch] = useReducer(RoomReducer, initialRoomState);

	const fetchRooms = async () => {
		try {
			const { data: response } = await fetchRecentRoomsApi(api);
			dispatch({ type: FETCH_ROOMS, payload: response });
		} catch (error) {
			console.error("Failed to fetch rooms", error);
			return { isSuccessResponse: false, error: error };
		}
		return { isSuccessResponse: true, error: null };
	};

	const createRoom = async (roomData: Partial<IRoom>) => {
		try {
			const { data: response } = await createRoomApi(api, roomData);
			const updatedRooms = state.rooms.concat(response.data.room);
			const updatedTotal = updatedRooms.length;
			dispatch({
				type: CREATE_ROOM,
				payload: { rooms: updatedRooms, total: updatedTotal },
			});
		} catch (error) {
			console.error("Failed to create room", error);
			return { isSuccessResponse: false, error: error };
		}
		return { isSuccessResponse: true, error: null };
	};

	const respondToInvite = async (roomId: string, accept: boolean) => {
		try {
			const { data: response } = await respondToInviteApi(
				api,
				roomId.toString(),
				accept
			);

			if (accept) {
				const updatedRooms = state.rooms.concat(response.data.room as IRoom);
				const updatedTotal = updatedRooms.length;
				dispatch({
					type: JOIN_ROOM,
					payload: { rooms: updatedRooms, total: updatedTotal },
				});
			} else {
				dispatch({ type: DECLINE_ROOM });
			}
		} catch (error) {
			console.error("Failed to respond to invite", error);
			return { isSuccessResponse: false, error: error };
		}
		return { isSuccessResponse: true, error: null };
	};

	const closeRoom = async (roomId: string) => {
		try {
			await closeRoomApi(api, roomId);

			const updatedRooms = state.rooms.filter((room) => room.id !== roomId);
			const updatedTotal = updatedRooms.length;
			dispatch({
				type: CLOSE_ROOM,
				payload: { rooms: updatedRooms, total: updatedTotal },
			});
		} catch (error) {
			console.error("Failed to close room", error);
			return { isSuccessResponse: false, error: error };
		}
		return { isSuccessResponse: true, error: null };
	};

	const value = {
		total: state.total,
		rooms: state.rooms,
		fetchRooms,
		createRoom,
		respondToInvite,
		closeRoom,
	};

	return <Provider value={value}>{children}</Provider>;
};

const useRoomCtx = () => useContextWrapper(RoomContext);

export { useRoomCtx, RoomProvider };
