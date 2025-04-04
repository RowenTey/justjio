import React from "react";

type ResponseModalProps = {
  title: string;
  message: string;
  type: "error" | "success";
  modalVisible: boolean;
  closeModal: () => void;
};

const ResponseModal: React.FC<ResponseModalProps> = ({
  title,
  message,
  type,
  modalVisible,
  closeModal,
}) => {
  if (!modalVisible) {
    return null;
  }

  return (
    <div className="fixed inset-0 flex items-center justify-center z-50 bg-black bg-opacity-50">
      <div
        className={`w-80 p-6 bg-white rounded-xl shadow-lg border-4 ${
          type === "error" ? "border-red-500" : "border-green-500"
        } flex flex-col gap-3 items-center justify-center`}
      >
        <h2
          className={`text-2xl font-bold ${
            type === "error" ? "text-red-500" : "text-green-500"
          }`}
        >
          {title}
        </h2>
        <p className="text-lg text-center text-black">{message}</p>
        <button
          className={`w-32 py-2 rounded-lg text-white font-semibold ${
            type === "error" ? "bg-red-500" : "bg-green-500"
          }`}
          onClick={closeModal}
        >
          {type === "error" ? "Retry" : "Ok"}
        </button>
      </div>
    </div>
  );
};

export default ResponseModal;
