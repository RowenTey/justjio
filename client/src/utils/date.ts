export function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("en-GB");
}

export function toDayOfWeek(date: string): string {
  return new Date(date).toLocaleDateString("en-GB", {
    weekday: "long",
  });
}
