import React, { createContext, useContext, useState, ReactNode } from "react";
import Toast from "../components/Toast";

type ToastContextType = {
	showToast: (message: string, isError: boolean, className?: string) => void;
};

type ToastType = {
	id: number;
	message: string;
	className: string;
	visible: boolean;
};

const ToastContext = createContext<ToastContextType | undefined>(undefined);

const ToastProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
	const [toasts, setToasts] = useState<ToastType[]>([]);

	const showToast = (message: string, isError: boolean, className = "") => {
		const baseStyles = isError
			? "bg-error-bg text-error"
			: "bg-success-bg text-success";
		const newToast: ToastType = {
			id: Date.now(),
			message,
			className: `${baseStyles} ${className}`,
			visible: true,
		};
		setToasts((prevToasts) => [...prevToasts, newToast]);

		setTimeout(() => {
			setToasts((prevToasts) =>
				prevToasts.filter((toast) => toast.id !== newToast.id)
			);
		}, 3000);
	};

	return (
		<ToastContext.Provider value={{ showToast }}>
			{children}
			<div className="fixed top-4 left-0 right-0 flex flex-col items-center gap-2 p-4 pointer-events-none ">
				{toasts.map((toast) => (
					<Toast
						key={toast.id}
						message={toast.message}
						className={toast.className}
						visible={toast.visible}
					/>
				))}
			</div>
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
