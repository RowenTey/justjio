import React from "react";

import { ArrowRightStartOnRectangleIcon } from "@heroicons/react/24/outline";
import Spinner from "../Spinner";

interface HomeTopBarProps {
	isLoading: boolean;
	username: string;
	onLogout: () => void;
}

const HomeTopBar: React.FC<HomeTopBarProps> = ({
	isLoading,
	username,
	onLogout,
}) => {
	return (
		<div className="relative top-0 flex h-[8%] items-center justify-between w-full py-4 pl-4 pr-6 bg-purple-200">
			<div className="flex items-center gap-1">
				<img
					src="/favicon.svg"
					alt="JustJio Logo"
					className="w-8 h-8 rounded-full"
				/>

				<h1 className="text-lg font-bold text-justjio-secondary">
					Welcome, {username}!
				</h1>
			</div>

			<button
				onClick={onLogout}
				className="flex items-center justify-center w-9 h-9 p-1 bg-justjio-secondary rounded-full hover:shadow-lg hover:border-2 hover:border-white"
			>
				{isLoading ? (
					<Spinner
						spinnerColor="border-white"
						spinnerSize={{ width: "w-6", height: "h-6" }}
					/>
				) : (
					<ArrowRightStartOnRectangleIcon className="w-6 h-6 text-white" />
				)}
			</button>
		</div>
	);
};

export default HomeTopBar;
