import React from "react";
import { ITransaction } from "../types/transaction";
import { CheckCircleIcon } from "@heroicons/react/24/solid";
import { BellIcon } from "@heroicons/react/24/outline";
import { useTransactionCtx } from "../context/transaction";
import Toast from "./Toast";
import { useToast } from "../context/toast";

type TransactionContainerProps = {
	title: string;
	emptyText: string;
	transactions: ITransaction[];
	isPayer: boolean;
};

const TransactionContainer: React.FC<TransactionContainerProps> = ({
	title,
	emptyText,
	transactions,
	isPayer,
}) => {
	return (
		<div className="flex flex-col gap-1 w-[47.5%] h-full bg-white shadow-md rounded-lg p-2">
			<h2 className="text-xs font-semibold text-gray-500">{title}</h2>
			<div
				className={`h-[82.5%] overflow-y-auto pr-1 flex flex-col w-full gap-2 ${
					transactions.length > 0 ? "" : "justify-center items-center"
				}`}
			>
				{transactions.length > 0 ? (
					transactions.map((transaction) => (
						<TransactionBox
							key={transaction.id}
							transaction={transaction}
							isPayer={isPayer}
						/>
					))
				) : (
					<p className="text-sm font-medium text-gray-700">{emptyText}</p>
				)}
			</div>
		</div>
	);
};

const TransactionBox = ({
	transaction,
	isPayer,
}: {
	transaction: ITransaction;
	isPayer: boolean;
}) => {
	const { settleTransaction } = useTransactionCtx();
	const { showToast } = useToast();

	const handleSettleTransaction = async (transactionId: number) => {
		console.log("[TransactionContainer] Settling transaction", transactionId);
		const res = await settleTransaction(transactionId);

		if (!res.isSuccessResponse) {
			showToast(`Failed to settle transaction: ${res.error}`, true);
			return;
		}

		showToast(`Transaction settled successfully!`, false);
	};

	return (
		<div
			className="flex items-center justify-between w-full 
				py-1 pl-2 pr-1 shadow-md rounded-lg"
		>
			<p className="text-xs font-semibold text-gray-700">
				{isPayer ? transaction.payee.username : transaction.payer.username}
			</p>
			<div className="flex items-center gap-1">
				<p className="text-xs font-semibold text-gray-700">
					${transaction.amount.toFixed(2)}
				</p>
				{isPayer ? (
					<CheckCircleIcon
						className="w-5 h-5 text-secondary cursor-pointer hover:scale-110"
						onClick={() => handleSettleTransaction(transaction.id)}
					/>
				) : (
					<BellIcon className="w-5 h-5 text-secondary cursor-pointer hover:scale-110" />
				)}
			</div>
			<Toast
				message="Transaction settled"
				visible={false}
				className="bg-green-400"
			/>
		</div>
	);
};

export default TransactionContainer;
