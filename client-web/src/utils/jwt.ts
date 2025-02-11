export interface DecodedJWTToken {
	user_id: number;
	user_email: string;
	username: string;
	exp?: number;
	iat?: number;
}

// see https://stackoverflow.com/questions/38552003/how-to-decode-jwt-token-in-javascript-without-using-a-library
export function jwtDecode<T>(token: string): T {
	try {
		const base64Url = token.split(".")[1];
		const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
		const jsonPayload = decodeURIComponent(
			atob(base64)
				.split("")
				.map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
				.join("")
		);
		return JSON.parse(jsonPayload);
	} catch (error) {
		throw new Error("Invalid token format");
	}
}
