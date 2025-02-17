import { useEffect, useRef } from "react";

export interface ModalWrapperProps {
	isVisible: boolean;
	closeModal: () => void;
}

const ModalWrapper = <P extends object>(
	WrappedComponent: React.ComponentType<P & ModalWrapperProps>
) => {
	const ModalComponent: React.FC<P & ModalWrapperProps> = ({
		isVisible,
		closeModal,
		...props
	}) => {
		const modalContentRef = useRef<HTMLDivElement>(null);

		useEffect(() => {
			const handleClickOutside = (event: MouseEvent) => {
				if (
					modalContentRef.current &&
					!modalContentRef.current.contains(event.target as Node)
				) {
					closeModal();
				}
			};

			if (isVisible) {
				document.addEventListener("mousedown", handleClickOutside);
			}

			return () => {
				document.removeEventListener("mousedown", handleClickOutside);
			};
		}, [isVisible, closeModal]);

		if (!isVisible) {
			return null;
		}

		return (
			<div className="fixed inset-0 flex items-center justify-center z-50 bg-black bg-opacity-50">
				<div
					ref={modalContentRef}
					className="w-80 p-6 bg-gray-200 rounded-xl border-[1px] border-secondary flex flex-col gap-3 items-center justify-center"
				>
					<WrappedComponent
						{...(props as P)}
						isVisible={isVisible}
						closeModal={closeModal}
					/>
				</div>
			</div>
		);
	};

	return ModalComponent;
};

export default ModalWrapper;
