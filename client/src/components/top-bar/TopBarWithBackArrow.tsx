import { ArrowLeftIcon } from "@heroicons/react/24/outline";
import React from "react";
import { useLocation, useNavigate } from "react-router-dom";

type TopBarWithBackArrowProps = {
  title: string;
  shouldCenterTitle?: boolean;
};

const TopBarWithBackArrow: React.FC<TopBarWithBackArrowProps> = ({
  title,
  shouldCenterTitle = true,
}) => {
  const navigate = useNavigate();
  const { state } = useLocation();

  return (
    <div
      className={`relative top-0 flex h-[8%] items-center w-full py-4 pl-3 pr-6 bg-purple-200 ${
        shouldCenterTitle ? "justify-center" : "justify-between"
      }`}
    >
      <button
        onClick={() => {
          if (!state?.from) {
            navigate("/");
          } else {
            navigate(-1);
          }
        }}
        className={`flex items-center justify-center p-1 bg-transparent hover:scale-110 ${
          shouldCenterTitle ? "absolute left-3" : ""
        }`}
      >
        <ArrowLeftIcon className="w-6 h-6 text-secondary" />
      </button>

      <h1
        className={`text-xl font-bold text-secondary ${
          shouldCenterTitle ? "ml-4" : ""
        }`}
      >
        {title}
      </h1>
    </div>
  );
};

export default TopBarWithBackArrow;
