export interface ISubscription {
  id: string;
  userId: number;
  endpoint: string;
  auth: string;
  p256dh: string;
}
