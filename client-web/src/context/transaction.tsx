import React, { createContext, useReducer } from "react";
import useContextWrapper from "../hooks/useContextWrapper";
import { TransactionContextType } from "../types/transaction";
import TransactionReducer, {
	initialTransactionState,
} from "../reducers/transaction";
import { fetchTransactionsApi, settleTransactionApi } from "../api/transaction";
import { useUserCtx } from "./user";
import { BaseContextResponse } from "../types";
import { api } from "../api";

export const FETCH_TRANSACTIONS = "FETCH_TRANSACTIONS";
export const SETTLE_TRANSACTION = "SETTLE_TRANSACTION";

const TransactionContext = createContext<TransactionContextType | null>(null);

const TransactionProvider: React.FC<{ children: React.ReactNode }> = ({
	children,
}) => {
	const [state, dispatch] = useReducer(
		TransactionReducer,
		initialTransactionState
	);
	const { user } = useUserCtx();

	const fetchTransactions = async (): Promise<BaseContextResponse> => {
		try {
			const { data: response } = await fetchTransactionsApi(api);
			const payload = {
				toPay: response.data.filter((tx) => tx.payerId === user.uid),
				toReceive: response.data.filter((tx) => tx.payeeId === user.uid),
			};
			console.log("Fetched transactions", payload);
			dispatch({ type: FETCH_TRANSACTIONS, payload });
			return { isSuccessResponse: true, error: null };
		} catch (error) {
			console.error("Failed to fetch transactions", error);
			return { isSuccessResponse: false, error: error };
		}
	};

	const settleTransaction = async (
		transactionId: number
	): Promise<BaseContextResponse> => {
		try {
			const res = await settleTransactionApi(api, transactionId);
			if (res.status !== 200) {
				throw new Error("Failed to settle transaction");
			}
			dispatch({ type: SETTLE_TRANSACTION, payload: transactionId });
			return { isSuccessResponse: true, error: null };
		} catch (error) {
			console.error("Failed to settle transaction", error);
			return { isSuccessResponse: false, error: error };
		}
	};

	return (
		<TransactionContext.Provider
			value={{
				toPay: state.toPay,
				toReceive: state.toReceive,
				fetchTransactions,
				settleTransaction,
			}}
		>
			{children}
		</TransactionContext.Provider>
	);
};

const useTransactionCtx = () => useContextWrapper(TransactionContext);

export { useTransactionCtx, TransactionProvider };
