import React from "react";
import { useNavigate } from "react-router-dom";
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
	const navigate = useNavigate();

	return (
		<div
			className="min-w-36 w-36 h-36 flex flex-col items-center justify-center bg-purple-200 rounded-lg shadow-md p-4 cursor-pointer transition-transform transform hover:scale-105"
			onClick={() => navigate(`/room/${id}`)}
		>
			<img src={img} alt="Group" className="w-16 h-16" />
			<p className="mt-3 text-lg font-bold text-center leading-[1.1] text-justjio-secondary">
				{name}
			</p>
		</div>
	);
};

export default RoomCard;
