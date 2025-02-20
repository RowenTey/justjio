import React from "react";
import ModalWrapper from "../ModalWrapper";

interface ConfirmJoinModalProps {
  isVisible: boolean;
  rejectJoin: () => void;
  confirmJoin: () => void;
}

const ConfirmJoinModalContent: React.FC<ConfirmJoinModalProps> = ({
  rejectJoin,
  confirmJoin,
}) => {
  return (
    <div className="w-full flex flex-col items-center gap-3">
      <h3 className="text-3xl font-bold text-secondary mb-2">Confirm Join</h3>
      <p className="text-black">Do you want to join this room?</p>
      <div className="flex justify-between w-4/5">
        <button
          className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
          onClick={rejectJoin}
        >
          No
        </button>
        <button
          className="bg-secondary hover:bg-tertiary text-white font-bold py-2 px-4 rounded-full mt-2 w-2/5"
          onClick={confirmJoin}
        >
          Yes
        </button>
      </div>
    </div>
  );
};

const ConfirmJoinModal = ModalWrapper(ConfirmJoinModalContent);

export default ConfirmJoinModal;
