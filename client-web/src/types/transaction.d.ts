import { BaseUserInfo } from "./user";

export interface ITransaction {
  id: number;
  consolidationId: number;
  payerId: number;
  payer: BaseUserInfo;
  payeeId: number;
  payee: BaseUserInfo;
  amount: number;
  isPaid: boolean;
  paidOn: string;
}

export interface TransactionState {
  toPay: ITransaction[];
  toReceive: ITransaction[];
}

export type TransactionContextType = {
  toPay: ITransaction[];
  toReceive: ITransaction[];
  fetchTransactions: () => Promise<BaseContextResponse>;
  settleTransaction: (transactionId: number) => Promise<BaseContextResponse>;
};

type TransactionActionTypes =
  | {
      type: "FETCH_TRANSACTIONS";
      payload: { toPay: ITransaction[]; toReceive: ITransaction[] };
    }
  | { type: "SETTLE_TRANSACTION"; payload: number };
