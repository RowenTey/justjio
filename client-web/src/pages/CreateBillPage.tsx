import { SubmitHandler, useForm } from "react-hook-form";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import useLoadingAndError from "../hooks/useLoadingAndError";
import InputField from "../components/InputField";
import Spinner from "../components/Spinner";
import { useLocation, useNavigate } from "react-router-dom";
import React, { useEffect, useState } from "react";
import { IUser } from "../types/user";
import PeopleBox from "../components/PeopleBox";
import Checkbox from "../components/Checkbox";
import useMandatoryParam from "../hooks/useMandatoryParam";
import { createBillApi } from "../api/bill";
import { api } from "../api";

type CreateBillFormData = {
	name: string;
	amount: number;
};

type CreateBillPageProps = {
	attendees: IUser[];
	roomName: string;
	currentUserId: number;
};

type SimplifiedUser = {
	id: number;
	username: string;
};

const CreateBillPage = () => {
	const { loading, startLoading, stopLoading, error, setErrorMsg } =
		useLoadingAndError();
	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<CreateBillFormData>();
	const navigate = useNavigate();
	const roomId = useMandatoryParam("roomId");
	const { state } = useLocation();
	const { attendees, roomName, currentUserId } = state as CreateBillPageProps;
	const simplifiedAttendees = attendees
		.map<SimplifiedUser>((attendee) => ({
			id: attendee.id,
			username: attendee.username,
		}))
		.filter((attendee) => attendee.id !== currentUserId);
	const [payers, setPayers] = useState<SimplifiedUser[]>(simplifiedAttendees);
	const [includeOwner, setIncludeOwner] = useState(true);

	const onSubmit: SubmitHandler<CreateBillFormData> = async (data) => {
		startLoading();

		const billData = {
			name: data.name,
			amount: Number(data.amount),
			includeOwner: includeOwner,
			roomId,
			payers: payers.map((payer) => payer.id),
		};
		console.log("[CreateBillPage] Submitted data: ", billData);

		const res = await createBillApi(api, billData);

		if (res.status !== 200) {
			setErrorMsg("An error occurred: " + res.data.message);
			stopLoading();
			return;
		}

		stopLoading();
		navigate(-1);
	};

	useEffect(() => {
		console.log("[CreateBillPage] Payers: ", payers);
	}, [payers]);

	return (
		<div className="h-full flex flex-col items-center bg-gray-200">
			<RoomTopBar title="Create Bill" shouldCenterTitle={true} />

			<div className="flex flex-col justify-center items-center gap-4 w-4/5 h-[92%]">
				<h2 className="text-lg font-bold text-secondary">
					Bill for:{" "}
					<span className="text-white font-semibold bg-secondary rounded-full px-3 py-1 ml-1">
						{roomName}
					</span>
				</h2>

				<form
					onSubmit={handleSubmit(onSubmit)}
					id="create-room-form"
					className="flex flex-col items-center gap-2 w-[85%] max-h-[85%]"
				>
					<InputField
						name="name"
						type="text"
						label="Bill Name"
						placeholder="Enter bill name"
						errors={errors}
						register={register}
						validation={{ required: "Bill Name is required" }}
					/>

					<InputField
						name="amount"
						type="number"
						label="Amount"
						isFloat={true}
						placeholder="Enter amount"
						pattern="[0-9]*"
						errors={errors}
						register={register}
						validation={{ required: "amount is required" }}
					/>

					<SelectMembersInput
						initialPayers={simplifiedAttendees}
						payers={payers}
						onPayersChange={setPayers}
					/>

					<Checkbox
						isChecked={includeOwner}
						onChange={() => setIncludeOwner(!includeOwner)}
						label="Include myself in the bill"
					/>

					<button
						className="bg-secondary hover:bg-tertiary text-white 
							font-bold py-2 px-4 rounded-full mt-2 w-3/5"
						form="create-room-form"
					>
						{loading ? (
							<Spinner spinnerSize={{ width: "w-6", height: "h-6" }} />
						) : (
							"Submit"
						)}
					</button>

					{error && (
						<p className="text-red-500 text-wrap text-center">{error}</p>
					)}
				</form>
			</div>
		</div>
	);
};

const SelectMembersInput: React.FC<{
	initialPayers: SimplifiedUser[];
	payers: SimplifiedUser[];
	onPayersChange: React.Dispatch<React.SetStateAction<SimplifiedUser[]>>;
}> = ({ initialPayers, payers, onPayersChange }) => {
	const [allSelected, setAllSelected] = useState(true);

	useEffect(() => {
		setAllSelected(payers.length === initialPayers.length);
	}, [payers, initialPayers]);

	const handleCheckChange = (user: SimplifiedUser, shouldRemove: boolean) => {
		onPayersChange((prevPayers) =>
			shouldRemove
				? prevPayers.filter((payer) => payer.id !== user.id)
				: [...prevPayers, user]
		);
	};

	const handleSelectAll = () => {
		onPayersChange(allSelected ? [] : initialPayers);
		setAllSelected(!allSelected);
	};

	return (
		<div className="flex flex-col gap-2 w-full my-1 h-full">
			<div className="flex justify-between items-center w-full">
				<label className="font-semibold text-md text-secondary">
					Select payers:
				</label>
				<Checkbox
					isChecked={allSelected}
					label={allSelected ? "Deselect All" : "Select All"}
					onChange={handleSelectAll}
				/>
			</div>
			<div className="pr-2 flex flex-col gap-2 h-full overflow-y-auto">
				{initialPayers.map((user) => (
					<SelectBox
						key={user.id}
						user={user}
						onCheckChange={handleCheckChange}
						isChecked={payers.some((payer) => payer.id === user.id)}
					/>
				))}
			</div>
		</div>
	);
};

const SelectBox: React.FC<{
	user: SimplifiedUser;
	onCheckChange: (user: SimplifiedUser, shouldRemove: boolean) => void;
	isChecked: boolean;
}> = ({ user, onCheckChange, isChecked }) => {
	const [check, setCheck] = useState(isChecked);

	useEffect(() => {
		setCheck(isChecked);
	}, [isChecked]);

	const onChange = () => {
		onCheckChange(user, check);
		setCheck(!check);
	};

	return (
		<PeopleBox
			name={user.username}
			isHost={false}
			isChecked={check}
			onClick={onChange}
		/>
	);
};

export default CreateBillPage;
