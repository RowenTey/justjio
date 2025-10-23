import { LOGOUT } from "../context/auth";
import {
  CLOSE_ROOM,
  CREATE_ROOM,
  DECLINE_ROOM,
  FETCH_ROOMS,
  JOIN_ROOM,
  LEAVE_ROOM,
} from "../context/room";
import { RoomActionTypes, RoomCtxState } from "../types/room";

export const INITIAL_ROOM_CTX_STATE: RoomCtxState = {
  rooms: [],
};

const RoomReducer = (
  state: RoomCtxState,
  action: RoomActionTypes,
): RoomCtxState => {
  const { type, payload } = action;

  switch (type) {
    case FETCH_ROOMS:
      return {
        ...state,
        rooms: payload,
      };
    case CREATE_ROOM:
    case CLOSE_ROOM:
    case LEAVE_ROOM:
    case JOIN_ROOM:
      return {
        ...state,
        rooms: payload,
      };
    case DECLINE_ROOM:
      return {
        ...state,
      };
    case LOGOUT:
      return INITIAL_ROOM_CTX_STATE;
    default:
      throw new Error(`No case for type ${type} found in RoomReducer.`);
  }
};

export default RoomReducer;
