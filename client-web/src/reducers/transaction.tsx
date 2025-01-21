import { FETCH_TRANSACTIONS, SETTLE_TRANSACTION } from "../context/transaction";
import { TransactionActionTypes, TransactionState } from "../types/transaction";

export const initialTransactionState: TransactionState = {
	toPay: [],
	toReceive: [],
};

const TransactionReducer = (
	state: TransactionState,
	action: TransactionActionTypes
): TransactionState => {
	const { type, payload } = action;

	switch (type) {
		case FETCH_TRANSACTIONS:
			return {
				...state,
				toPay: payload.toPay,
				toReceive: payload.toReceive,
			};
		case SETTLE_TRANSACTION:
			return {
				...state,
				toPay: state.toPay.filter((tx) => tx.id !== payload),
			};
		default:
			throw new Error(`No case for type ${type} found in TransactionReducer.`);
	}
};

export default TransactionReducer;
