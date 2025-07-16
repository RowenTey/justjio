import { ReactNode, createContext, useReducer } from "react";
import { IRoom, RoomContextType } from "../types/room";
import RoomReducer, { initialRoomState } from "../reducers/room";
import {
  closeRoomApi,
  createRoomApi,
  fetchRecentRoomsApi,
  leaveRoomApi,
  respondToInviteApi,
} from "../api/room";
import { api } from "../api";
import useContextWrapper from "../hooks/useContextWrapper";
import { AxiosError } from "axios";
import { BaseContextResponse } from "../types";

interface RoomProviderProps {
  children: ReactNode;
}

export const CLOSE_ROOM = "CLOSE_ROOM";
export const LEAVE_ROOM = "LEAVE_ROOM";
export const CREATE_ROOM = "CREATE_ROOM";
export const DECLINE_ROOM = "DECLINE_ROOM";
export const FETCH_ROOMS = "FETCH_ROOMS";
export const JOIN_ROOM = "JOIN_ROOM";

const RoomContext = createContext<RoomContextType | null>(null);
const { Provider } = RoomContext;

const RoomProvider: React.FC<RoomProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(RoomReducer, initialRoomState);

  const fetchRooms = async (): Promise<BaseContextResponse> => {
    try {
      const { data: response } = await fetchRecentRoomsApi(api);
      dispatch({ type: FETCH_ROOMS, payload: response });
    } catch (error) {
      console.error("Failed to fetch rooms", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
    return { isSuccessResponse: true, error: null };
  };

  const createRoom = async (
    roomData: Partial<IRoom>,
    attendeesId: string[],
  ): Promise<BaseContextResponse> => {
    try {
      const { data: response } = await createRoomApi(
        api,
        roomData,
        attendeesId,
      );
      const updatedRooms = state.rooms.concat(response.data.room);
      dispatch({
        type: CREATE_ROOM,
        payload: { data: updatedRooms },
      });
    } catch (error) {
      console.error("Failed to create room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
    return { isSuccessResponse: true, error: null };
  };

  const respondToInvite = async (
    roomId: string,
    accept: boolean,
  ): Promise<BaseContextResponse> => {
    try {
      const { data: response } = await respondToInviteApi(
        api,
        roomId.toString(),
        accept,
      );

      if (accept) {
        const updatedRooms = state.rooms.concat(response.data.room as IRoom);
        dispatch({
          type: JOIN_ROOM,
          payload: { data: updatedRooms },
        });
      } else {
        dispatch({ type: DECLINE_ROOM });
      }
    } catch (error) {
      console.error("Failed to respond to invite", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
    return { isSuccessResponse: true, error: null };
  };

  const closeRoom = async (roomId: string): Promise<BaseContextResponse> => {
    try {
      await closeRoomApi(api, roomId);

      const updatedRooms = state.rooms.filter((room) => room.id !== roomId);
      dispatch({
        type: CLOSE_ROOM,
        payload: { data: updatedRooms },
      });
    } catch (error) {
      console.error("Failed to close room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
    return { isSuccessResponse: true, error: null };
  };

  const leaveRoom = async (roomId: string): Promise<BaseContextResponse> => {
    try {
      await leaveRoomApi(api, roomId);

      const updatedRooms = state.rooms.filter((room) => room.id !== roomId);
      dispatch({
        type: LEAVE_ROOM,
        payload: { data: updatedRooms },
      });
    } catch (error) {
      console.error("Failed to close room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
    return { isSuccessResponse: true, error: null };
  };

  const value = {
    rooms: state.rooms,
    fetchRooms,
    createRoom,
    respondToInvite,
    closeRoom,
    leaveRoom,
  };

  return <Provider value={value}>{children}</Provider>;
};

const useRoomCtx = () => useContextWrapper(RoomContext);

export { useRoomCtx, RoomProvider };
