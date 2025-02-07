import React, { createContext, useContext, useState, ReactNode } from "react";
import Toast from "../components/Toast";

type ToastContextType = {
	showToast: (message: string, isError: boolean, className?: string) => void;
};

const ToastContext = createContext<ToastContextType | undefined>(undefined);

const ToastProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
	const [toast, setToast] = useState<{
		message: string;
		className: string;
		visible: boolean;
	}>({
		message: "",
		className: "",
		visible: false,
	});

	const showToast = (message: string, isError: boolean, className = "") => {
		const baseStyles = isError
			? "bg-error-bg text-error"
			: "bg-success-bg text-success";
		setToast({
			message,
			className: `${baseStyles} ${className}`,
			visible: true,
		});
		setTimeout(() => {
			setToast((prev) => ({ ...prev, visible: false }));
		}, 3000);
	};

	return (
		<ToastContext.Provider value={{ showToast }}>
			{children}
			<Toast
				message={toast.message}
				className={toast.className}
				visible={toast.visible}
			/>
		</ToastContext.Provider>
	);
};

const useToast = (): ToastContextType => {
	const context = useContext(ToastContext);
	if (!context) {
		throw new Error("useToast must be used within a ToastProvider");
	}
	return context;
};

export { useToast, ToastProvider };
