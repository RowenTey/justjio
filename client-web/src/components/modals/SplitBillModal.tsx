import { useEffect, useMemo, useState } from "react";
import { consolidateBillApi, fetchBillApi } from "../../api/bill";
import { IBill } from "../../types/bill";
import ModalWrapper from "../ModalWrapper";
import { api } from "../../api";

interface SplitBillModalProps {
	roomId: string;
	closeModal: () => void;
}

const SplitBillModalContent: React.FC<SplitBillModalProps> = ({
	roomId,
	closeModal,
}) => {
	const [bills, setBills] = useState<IBill[]>([]);
	const total = useMemo(
		() => bills.reduce((acc, bill) => acc + bill.amount, 0),
		[bills]
	);

	useEffect(() => {
		const fetchBill = async (roomId: string) => {
			const res = await fetchBillApi(api, roomId);

			if (res.status !== 200) {
				alert("Failed to fetch bill");
				return;
			}

			const { data } = res.data;
			console.log("[SplitBillPage] Bill data", data);

			setBills(data);
		};
		fetchBill(roomId);
	}, [roomId]);

	const onSubmit = async () => {
		const res = await consolidateBillApi(api, roomId);

		if (res.status !== 200) {
			alert("Failed to consolidate bill");
			return;
		}

		setTimeout(() => closeModal(), 1000);
	};

	return (
		<>
			<h2 className="text-3xl font-bold text-justjio-secondary">Split Bill</h2>
			<div className="h-full w-full flex flex-col justify-center items-center gap-3">
				<div className="h-[30%] w-full flex flex-col items-center gap-2 overflow-y-auto">
					{bills.map((bill) => (
						<div
							key={bill.id}
							className="flex gap-1 px-4 py-2 w-full rounded-xl justify-between bg-white border-2 border-justjio-secondary text-justjio-secondary"
						>
							<p className="text-sm">
								{bill.name} - <span>{bill.owner.username}</span>
							</p>
							<p className="text-sm">${bill.amount}</p>
							{/* <p className="text-sm">Created On: {formatDate(bill.date)}</p> */}
						</div>
					))}
				</div>

				<p className="text-lg font-semibold text-justjio-secondary">
					Total: {total}
				</p>

				<button
					className="bg-justjio-secondary hover:bg-purple-900 text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
					onClick={onSubmit}
				>
					Submit
				</button>
			</div>
		</>
	);
};

const SplitBillModal = ModalWrapper(SplitBillModalContent);

export default SplitBillModal;
