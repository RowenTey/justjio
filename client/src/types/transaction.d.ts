import { TransactionDto } from "./models";

export interface TransactionState {
  toPay: TransactionDto[];
  toReceive: TransactionDto[];
}

export type TransactionContextType = {
  toPay: TransactionDto[];
  toReceive: TransactionDto[];
  fetchTransactions: () => Promise<BaseContextResponse>;
  settleTransaction: (transactionId: number) => Promise<BaseContextResponse>;
};

export type TransactionActionTypes =
  | {
      type: "FETCH_TRANSACTIONS";
      payload: { toPay: TransactionDto[]; toReceive: TransactionDto[] };
    }
  | { type: "SETTLE_TRANSACTION"; payload: number };
