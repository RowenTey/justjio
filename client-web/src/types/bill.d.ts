export interface IBill {
	id: string;
	name: string;
	amount: number;
	date: string;
	includeOwner: boolean;
	roomId: string;
	payers: number[];
	ownerId: number;
	owner: {
		id: number;
		username: string;
	};
	consolidationId?: string;
}

export interface IConsolidation {
	id: number;
	createdAt: string;
}
