import { ReactNode, createContext, useReducer } from "react";
import { RoomContextType } from "../types/room";
import RoomReducer, { INITIAL_ROOM_CTX_STATE } from "../reducers/room";
import { roomService } from "../services/room.service";
import useContextWrapper from "../hooks/useContextWrapper";
import { AxiosError } from "axios";
import { BaseContextResponse, Optional } from "../types";
import {
  CreateRoomRequest,
  RespondToRoomInviteRequest,
  RoomListDto,
} from "../types/models";

interface RoomProviderProps {
  children: ReactNode;
}

export const CLOSE_ROOM = "CLOSE_ROOM";
export const LEAVE_ROOM = "LEAVE_ROOM";
export const CREATE_ROOM = "CREATE_ROOM";
export const DECLINE_ROOM = "DECLINE_ROOM";
export const FETCH_ROOMS = "FETCH_ROOMS";
export const JOIN_ROOM = "JOIN_ROOM";

const RoomContext = createContext<Optional<RoomContextType>>(null);
const { Provider } = RoomContext;

const RoomProvider: React.FC<RoomProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(RoomReducer, INITIAL_ROOM_CTX_STATE);

  const fetchRooms = async (): Promise<BaseContextResponse> => {
    try {
      const response = await roomService.getRooms();
      dispatch({ type: FETCH_ROOMS, payload: response.data as RoomListDto[] });
      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to fetch rooms", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const createRoom = async (
    data: CreateRoomRequest,
  ): Promise<BaseContextResponse> => {
    try {
      const response = await roomService.createRoom(data);

      const updatedRooms = state.rooms.concat({
        id: response.data,
        name: data.name,
        isClosed: false,
        isPrivate: data.isPrivate,
        noOfAttendees: 1,
        imageUrl: data.imageUrl,
      } as RoomListDto);
      dispatch({
        type: CREATE_ROOM,
        payload: updatedRooms,
      });

      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to create room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const respondToInvite = async (
    roomId: string,
    accept: boolean,
  ): Promise<BaseContextResponse> => {
    try {
      const response = await roomService.respondToInvite(roomId, {
        accept,
      } as RespondToRoomInviteRequest);

      if (accept) {
        const updatedRooms = state.rooms.concat({
          id: response.data?.id,
          name: response.data?.name,
          isClosed: false,
          isPrivate: response.data?.isPrivate,
          noOfAttendees: 1,
          imageUrl: response.data?.imageUrl,
          host: response.data?.host,
        } as RoomListDto);
        dispatch({
          type: JOIN_ROOM,
          payload: updatedRooms,
        });
      } else {
        dispatch({ type: DECLINE_ROOM });
      }

      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to respond to invite", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const closeRoom = async (roomId: string): Promise<BaseContextResponse> => {
    try {
      await roomService.closeRoom(roomId);

      const filteredRooms = state.rooms.filter((room) => room.id !== roomId);
      dispatch({
        type: CLOSE_ROOM,
        payload: filteredRooms,
      });

      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to close room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
  };

  const leaveRoom = async (roomId: string): Promise<BaseContextResponse> => {
    try {
      await roomService.leaveRoom(roomId);

      const filteredRooms = state.rooms.filter((room) => room.id !== roomId);
      dispatch({
        type: LEAVE_ROOM,
        payload: filteredRooms,
      });

      return { isSuccessResponse: true, error: null };
    } catch (error) {
      console.error("Failed to close room", error);
      return { isSuccessResponse: false, error: error as AxiosError };
    }
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
