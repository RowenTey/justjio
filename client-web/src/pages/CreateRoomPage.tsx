import { SubmitHandler, useForm } from "react-hook-form";
import InputField from "../components/InputField";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import { useRoomCtx } from "../context/room";
import useLoadingAndError from "../hooks/useLoadingAndError";
import { useNavigate } from "react-router-dom";
import Spinner from "../components/Spinner";

type CreateRoomFormData = {
	roomName: string;
	date: string;
	venue: string;
	time: string;
	message?: string;
};

const CreateRoomPage = () => {
	const { loading, startLoading, stopLoading, error, setErrorMsg } =
		useLoadingAndError();
	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<CreateRoomFormData>();
	const { createRoom } = useRoomCtx();
	const navigate = useNavigate();

	const onSubmit: SubmitHandler<CreateRoomFormData> = async (data) => {
		startLoading();

		const parsedDate = new Date(data.date);
		data.date = parsedDate.toISOString();

		console.log("[CreateRoomPage] Submitted data: ", data);
		const res = await createRoom(data);

		if (!res.isSuccessResponse) {
			setErrorMsg("An error occurred: " + res.error);
			stopLoading();
			return;
		}

		stopLoading();
		navigate("/");
	};

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<RoomTopBar title="Create Room" />

			<form
				onSubmit={handleSubmit(onSubmit)}
				id="create-room-form"
				className="flex flex-col justify-center items-center gap-2 w-[85%]"
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
					name="date"
					type="date"
					label="Date"
					placeholder="Enter date"
					errors={errors}
					register={register}
					validation={{ required: "Date is required" }}
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
					name="time"
					type="time"
					label="Time"
					placeholder="Enter time"
					errors={errors}
					register={register}
					validation={{ required: "Time is required" }}
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
					className="bg-justjio-secondary hover:shadow-lg hover:border-2 hover:border-white text-white font-bold py-2 px-4 rounded-full mt-4 w-2/5"
					form="create-room-form"
				>
					{loading ? (
						<Spinner spinnerSize={{ width: "w-6", height: "h-6" }} />
					) : (
						"Submit"
					)}
				</button>

				{error && <p className="text-red-500 text-wrap text-center">{error}</p>}
			</form>
		</div>
	);
};

export default CreateRoomPage;
