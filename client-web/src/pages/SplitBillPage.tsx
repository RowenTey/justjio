import { useEffect, useMemo, useState } from "react";
import { consolidateBillApi, fetchBillApi } from "../api/bill";
import { api } from "../api";
import useMandatoryParam from "../hooks/useMandatoryParam";
import RoomTopBar from "../components/top-bar/TopBarWithBackArrow";
import { IBill } from "../types/bill";
import { formatDate } from "../utils/date";
import { useLocation, useNavigate } from "react-router-dom";

const SplitBillPage: React.FC = () => {
	const roomId = useMandatoryParam("roomId");
	const [bills, setBills] = useState<IBill[]>([]);
	const [expandedOwners, setExpandedOwners] = useState<Set<number>>(new Set());
	const { state } = useLocation();
	const { roomName } = state as { roomName: string };
	const navigate = useNavigate();

	const total = useMemo(
		() => bills.reduce((acc, bill) => acc + bill.amount, 0),
		[bills]
	);

	// Group bills by owner
	const groupedBills = useMemo(() => {
		const grouped = new Map<number, IBill[]>();

		bills.forEach((bill) => {
			const ownerId = bill.owner.id;
			if (!grouped.has(ownerId)) {
				grouped.set(ownerId, []);
			}
			grouped.get(ownerId)?.push(bill);
		});

		console.log("[SplitBillPage] Grouped bills", grouped);
		return grouped;
	}, [bills]);

	// Toggle expand/collapse for an owner
	const toggleExpand = (ownerId: number) => {
		setExpandedOwners((prev) => {
			const newSet = new Set(prev);

			if (newSet.has(ownerId)) {
				newSet.delete(ownerId);
			} else {
				newSet.add(ownerId);
			}

			return newSet;
		});
	};

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

		alert("Bill consolidated successfully");
		navigate(-1);
	};

	return (
		<div className="h-full flex flex-col items-center gap-4 bg-gray-200">
			<RoomTopBar title="Split Bill" shouldCenterTitle={true} />

			<div className="h-full w-4/5 flex flex-col items-center gap-2">
				<h2 className="text-lg font-bold text-justjio-secondary mb-1">
					Bill(s) for:{" "}
					<span className="text-white bg-justjio-secondary rounded-full px-3 py-1">
						{roomName}
					</span>
				</h2>

				<div className="h-[80%] w-full p-3 flex flex-col items-center justify-between bg-justjio-primary rounded-xl">
					<div className="h-[90%] pr-1 w-full flex flex-col items-center gap-2 overflow-y-auto">
						{Array.from(groupedBills.entries()).map(([ownerId, bills]) => {
							const owner = bills[0].owner;
							const totalAmount = bills.reduce(
								(acc, bill) => acc + bill.amount,
								0
							);
							const isExpanded = expandedOwners.has(ownerId);

							return (
								<div
									key={ownerId}
									className="w-full rounded-xl bg-white border-2 border-justjio-secondary text-justjio-secondary"
								>
									<div className="flex gap-1 px-4 py-2 justify-between">
										<p className="text-sm">
											{owner.username} -{" "}
											<span className="font-bold">
												${totalAmount.toFixed(2)}
											</span>
										</p>
										<p
											className="text-sm cursor-pointer"
											onClick={() => toggleExpand(Number(ownerId))}
										>
											{isExpanded ? "▼" : "▶"}
										</p>
									</div>

									{isExpanded && (
										<div className="px-4 py-2 border-t border-justjio-secondary">
											{bills.map((bill) => (
												<div
													key={bill.id}
													className="flex gap-1 justify-between text-sm"
												>
													<p>
														{formatDate(bill.date)}: {bill.name}
													</p>
													<p>${bill.amount.toFixed(2)}</p>
												</div>
											))}
										</div>
									)}
								</div>
							);
						})}
					</div>

					<p className="text-lg text-justjio-secondary">
						Total: <span className="font-bold">{total.toFixed(2)}</span>
					</p>
				</div>

				<button
					className="bg-justjio-secondary hover:bg-purple-900 text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
					onClick={onSubmit}
				>
					Submit
				</button>
			</div>
		</div>
	);
};

export default SplitBillPage;
