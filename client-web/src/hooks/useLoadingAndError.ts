import { useState } from "react";

const useLoadingAndError = (count: number = 1) => {
	const [loadingStates, setLoadingStates] = useState<boolean[]>(
		new Array(count).fill(false)
	);
	const [errorStates, setErrorStates] = useState<string[]>(
		new Array(count).fill("")
	);

	const startLoading = (index: number = 0) => {
		if (index >= 0 && index < count) {
			setLoadingStates((prev) => {
				const newStates = [...prev];
				newStates[index] = true;
				return newStates;
			});
		}
	};

	const stopLoading = (index: number = 0) => {
		if (index >= 0 && index < count) {
			setLoadingStates((prev) => {
				const newStates = [...prev];
				newStates[index] = false;
				return newStates;
			});
		}
	};

	const setErrorMsg = (msg: string, index: number = 0) => {
		if (index >= 0 && index < count) {
			setErrorStates((prev) => {
				const newStates = [...prev];
				newStates[index] = msg;
				return newStates;
			});
		}
	};

	const clearError = (index: number = 0) => {
		if (index >= 0 && index < count) {
			setErrorStates((prev) => {
				const newStates = [...prev];
				newStates[index] = "";
				return newStates;
			});
		}
	};

	return {
		loadingStates,
		errorStates,
		startLoading,
		stopLoading,
		setErrorMsg,
		clearError,
	};
};

export default useLoadingAndError;
