import { useForm } from "react-hook-form";
import ModalWrapper from "../ModalWrapper";
import SearchableDropdown from "../SearchableDropdown";

interface InviteAttendeesFormData {
	invitees: string;
}

const InviteAttendeesModalContent: React.FC = () => {
	const {
		register,
		handleSubmit,
		setValue,
		formState: { errors },
	} = useForm<InviteAttendeesFormData>();

	return (
		<>
			<h2 className="text-3xl font-bold text-justjio-secondary">
				Invite People
			</h2>
			<form
				onSubmit={handleSubmit((data: InviteAttendeesFormData) => {
					console.log(data);
				})}
				className="w-full flex flex-col items-center justify-center gap-3"
				id="invite-people-form"
			>
				<SearchableDropdown
					name="invitees"
					errors={errors}
					register={register}
					onSelect={(selected) => {
						console.log(selected);
						setValue(
							"invitees",
							selected.map((option) => option.value).join(",")
						);
					}}
					options={[
						{ label: "John Doe", value: "1" },
						{ label: "Jane Doe", value: "2" },
						{ label: "John Smith", value: "3" },
					]}
					validation={{ required: "Invitees are required" }}
				/>

				<button
					className="w-32 py-2 mt-2 rounded-full text-black font-semibold bg-justjio-primary"
					form="invite-people-form"
				>
					Submit
				</button>
			</form>
		</>
	);
};

const InviteAttendeesModal = ModalWrapper(InviteAttendeesModalContent);

export default InviteAttendeesModal;
