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
			className={`max-w-[275px] text-center text-wrap transition-opacity duration-300 ease-in-out px-4 py-2 text-sm font-semibold rounded-2xl shadow-lg  ${
				visible ? "opacity-100" : "opacity-0"
			} ${className}`}
		>
			{message}
		</div>
	);
};

export default Toast;
