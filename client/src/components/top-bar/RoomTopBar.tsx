import { ArrowLeftIcon, PencilIcon } from "@heroicons/react/24/outline";
import React from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { IRoom } from "../../types/room";

type RoomTopBarProps = {
  room: IRoom;
  showEditBtn: boolean;
};

const RoomTopBar: React.FC<RoomTopBarProps> = ({ room, showEditBtn }) => {
  const navigate = useNavigate();
  const { state } = useLocation();

  return (
    <div className="relative top-0 flex h-[8%] items-center w-full py-4 bg-purple-200 justify-center">
      <button
        onClick={() => {
          if (!state.from) {
            navigate("/");
          } else {
            navigate(-1);
          }
        }}
        className="flex items-center justify-center p-1 bg-transparent hover:scale-110 absolute left-3"
      >
        <ArrowLeftIcon className="w-6 h-6 text-secondary" />
      </button>

      <h1 className={"text-xl font-bold text-secondary"}>{room.name}</h1>

      {showEditBtn && (
        <Link
          to={`/room/${room.id}/edit`}
          state={{ from: `/room/${room.id}`, room }}
          className="flex items-center justify-center p-1 bg-transparent hover:scale-110 absolute right-3"
        >
          <PencilIcon className="w-6 h-6 text-secondary" />
        </Link>
      )}
    </div>
  );
};

export default RoomTopBar;
