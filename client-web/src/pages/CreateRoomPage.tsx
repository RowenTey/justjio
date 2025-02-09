/* eslint-disable react-hooks/exhaustive-deps */
import { SubmitHandler, useForm } from "react-hook-form";
import InputField from "../components/InputField";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import { useRoomCtx } from "../context/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useNavigate } from "react-router-dom";
import Spinner from "../components/Spinner";
import SearchableDropdown from "../components/SearchableDropdown";
import { useUserCtx } from "../context/user";
import { useToast } from "../context/toast";
import { useEffect } from "react";

type CreateRoomFormData = {
	name: string;
	date: string;
	venue: string;
	time: string;
	invitees: string;
	message?: string;
};

const CreateRoomPage = () => {
	const { loading, startLoading, stopLoading } = useLoadingAndError();
	const {
		register,
		handleSubmit,
		setValue,
		formState: { errors },
	} = useForm<CreateRoomFormData>();
	const { user, friends, fetchFriends } = useUserCtx();
	const { createRoom } = useRoomCtx();
	const { showToast } = useToast();
	const navigate = useNavigate();

	useEffect(() => {
		if (friends.length !== 0) return;

		startLoading();
		fetchFriends(user.id).then(stopLoading);
	}, [user, friends]);

	const onSubmit: SubmitHandler<CreateRoomFormData> = async (data) => {
		startLoading();

		console.log("[CreateRoomPage] Date before: ", data.date);
		const parsedDate = new Date(data.date);
		data.date = parsedDate.toISOString();

		console.log("[CreateRoomPage] Submitted data: ", data);
		const roomData = {
			name: data.name,
			date: data.date,
			venue: data.venue,
			time: data.time,
		};
		const res = await createRoom(
			roomData,
			data.invitees.split(","),
			data.message
		);

		if (!res.isSuccessResponse) {
			switch (res.error?.response?.status) {
				case 400:
					showToast("Bad request, please check request body.", true);
					break;
				case 404:
					showToast("User not found, please try again later.", true);
					break;
				case 500:
				default:
					showToast("An error occurred, please try again later.", true);
					break;
			}
			stopLoading();
			return;
		}

		stopLoading();
		showToast("Room created successfully!", false);
		navigate("/");
	};

	return (
		<div className="h-full flex flex-col items-center bg-gray-200">
			<RoomTopBar title="Create Room" />

			<form
				onSubmit={handleSubmit(onSubmit)}
				id="create-room-form"
				className={`flex flex-col justify-center items-center p-2 gap-2 w-[85%] h-[92%] overflow-y-auto ${
					Object.keys(errors).length > 0 && "sm:pt-[15%]"
				}`}
			>
				<InputField
					name="name"
					type="text"
					label="Room Name"
					placeholder="Enter room name"
					errors={errors}
					register={register}
					validation={{ required: "Room Name is required" }}
				/>

				<InputField
					name="venue"
					type="text"
					label="Venue"
					placeholder="Enter venue"
					errors={errors}
					register={register}
					validation={{ required: "Venue is required" }}
				/>

				<InputField
					name="date"
					type="date"
					label="Date"
					placeholder="Enter date"
					errors={errors}
					register={register}
					validation={{
						required: "Date is required",
						min: {
							value: new Date().toISOString().split("T")[0],
							message: "Date must be in the future",
						},
					}}
					min={new Date().toISOString().split("T")[0]}
					defaultValue={new Date().toISOString().split("T")[0]}
				/>

				<InputField
					name="time"
					type="time"
					label="Time"
					placeholder="Enter time"
					min="00:00"
					max="23:59"
					errors={errors}
					register={register}
					validation={{
						required: "Time is required",
					}}
				/>

				<SearchableDropdown
					label="Invitees"
					name="invitees"
					errors={errors}
					register={register}
					onSelect={(selected) => {
						setValue(
							"invitees",
							selected.map((option) => option.value).join(",")
						);
					}}
					options={friends.map((friend) => ({
						label: friend.username,
						value: friend.id,
					}))}
					validation={{}}
				/>

				<InputField
					name="message"
					type="text"
					label="Message"
					placeholder="Enter invite message (optional)"
					errors={errors}
					register={register}
					validation={{}}
				/>

				<button
					className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-4 w-2/5"
					form="create-room-form"
				>
					{loading ? (
						<Spinner spinnerSize={{ width: "w-6", height: "h-6" }} />
					) : (
						"Submit"
					)}
				</button>
			</form>
		</div>
	);
};

export default CreateRoomPage;
