/* eslint-disable react-hooks/exhaustive-deps */
import { useEffect, useState } from "react";
import { useUserCtx } from "../context/user";
import useLoadingAndError from "../hooks/useLoadingAndError";
import Spinner from "../components/Spinner";
import TopBarWithBackArrow from "../components/top-bar/TopBarWithBackArrow";
import { useRoomCtx } from "../context/room";
import IMAGES from "../assets/images/Images";
import { Link } from "react-router-dom";
import { IRoom } from "../types/room";
import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";

const RoomsPage = () => {
	const { loading, startLoading, stopLoading } = useLoadingAndError();
	const [searchTerm, setSearchTerm] = useState("");
	const { user } = useUserCtx();
	const { rooms, fetchRooms } = useRoomCtx();
	const [filteredRooms, setFilteredRooms] = useState<IRoom[]>([]);

	useEffect(() => {
		startLoading();
		fetchRooms().then(stopLoading).catch(stopLoading);

		setFilteredRooms(rooms);
	}, [user.id]);

	useEffect(() => {
		if (searchTerm === "") {
			setFilteredRooms(rooms);
			return;
		}

		const updateFilteredRooms = () => {
			setFilteredRooms(
				rooms.filter((room) => room.name.toLowerCase().startsWith(searchTerm))
			);
		};

		// Debounce search input
		const debounceTimeout = setTimeout(updateFilteredRooms, 300);
		return () => clearTimeout(debounceTimeout);
	}, [searchTerm]);

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<TopBarWithBackArrow title="Rooms" />

			<div className="w-4/5 flex pr-2 items-center bg-white rounded-xl shadow-md border border-gray-300 peer-focus:border-2 peer-focus:border-secondary">
				<input
					type="text"
					placeholder="Search rooms..."
					className="peer w-full pl-3 pr-2 py-2 text-black bg-inherit rounded-l-xl focus:outline-none"
					onChange={(e) => setSearchTerm(e.target.value.toLowerCase())}
				/>
				<MagnifyingGlassIcon className="w-6 h-6 text-gray-500 cursor-pointer" />
			</div>

			<div className="w-full h-[92%] flex flex-col items-center px-4 gap-3">
				{loading ? (
					<Spinner spinnerSize={{ width: "w-10", height: "h-10" }} />
				) : filteredRooms.length > 0 ? (
					filteredRooms.map((room) => (
						<Link
							key={room.id}
							className="w-4/5 flex items-center p-4 gap-3 bg-white rounded-md shadow-md cursor-pointer hover:scale-105"
							to={`/room/${room.id}`}
							state={{ from: "/rooms" }}
						>
							<img
								src={IMAGES.group}
								alt="Room Image"
								className="w-7 h-7 rounded-md"
							/>
							<p className="text-md text-secondary font-semibold">
								{room.name}
							</p>
						</Link>
					))
				) : (
					<p className="text-lg font-semibold text-gray-500">No rooms found</p>
				)}
			</div>
		</div>
	);
};

export default RoomsPage;
