export const getRedirectPath = () => {
	return localStorage.getItem("redirectPath");
};

export const setRedirectPath = (path: string) => {
	localStorage.setItem("redirectPath", path);
};
