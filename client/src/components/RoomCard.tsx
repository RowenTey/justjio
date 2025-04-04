import React from "react";
import { Link } from "react-router-dom";
import IMAGES from "../assets/images/Images";

interface RoomCardProps {
  id: string;
  name: string;
  img?: string;
}

const RoomCard: React.FC<RoomCardProps> = ({
  id,
  name,
  img = IMAGES.group,
}) => {
  return (
    <Link
      to={`/room/${id}`}
      state={{ from: "/" }}
      className="min-w-36 w-36 h-36 flex flex-col items-center justify-center bg-purple-200 rounded-lg shadow-md p-4 cursor-pointer transition-transform transform hover:scale-105"
    >
      <img src={img} alt="Group" className="w-16 h-16" />
      <p className="mt-3 text-lg font-bold text-center leading-[1.1] text-secondary">
        {name}
      </p>
    </Link>
  );
};

export default RoomCard;
