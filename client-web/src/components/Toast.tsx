import React, { useEffect } from "react";

type ToastProps = {
	message: string;
	visible: boolean;
	className: string;
	onClose?: () => void;
};

const Toast: React.FC<ToastProps> = ({
	message,
	className,
	visible,
	onClose,
}) => {
	useEffect(() => {
		if (visible) {
			const timer = setTimeout(() => {
				onClose && onClose();
			}, 3000);
			return () => clearTimeout(timer);
		}
	}, [visible, onClose]);

	return (
		<div
			className={`fixed top-4 left-0 right-0 flex justify-center transition-opacity duration-300 ease-in-out ${
				visible ? "opacity-100" : "opacity-0 pointer-events-none"
			}`}
		>
			<div
				className={`px-4 py-2 text-sm font-semibold rounded-2xl shadow-lg ${className}`}
			>
				{message}
			</div>
		</div>
	);
};

export default Toast;
