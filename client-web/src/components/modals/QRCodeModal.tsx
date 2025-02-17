import React from "react";
import { QRCodeSVG } from "qrcode.react";
import ModalWrapper, { ModalWrapperProps } from "../ModalWrapper";

interface QRCodeModalProps {
	url: string;
}

const QRCodeModalContent: React.FC<QRCodeModalProps & ModalWrapperProps> = ({
	url,
	closeModal,
}) => {
	return (
		<div className="w-full flex flex-col items-center gap-3">
			<h2 className="text-3xl font-bold text-secondary mb-2">Room QR Code</h2>
			<QRCodeSVG value={url} />
			<button
				className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
				onClick={closeModal}
			>
				Close
			</button>
		</div>
	);
};

const QRCodeModal = ModalWrapper(QRCodeModalContent);

export default QRCodeModal;
