import { CheckIcon } from "@heroicons/react/24/outline";

interface PeopleBoxProps {
  name: string;
  pictureUrl: string;
  isHost?: boolean;
  isChecked?: boolean;
  onClick?: () => void;
}

const PeopleBox: React.FC<PeopleBoxProps> = ({
  name,
  pictureUrl,
  isHost,
  isChecked,
  onClick,
}) => {
  return (
    <div
      className={`flex gap-3 p-2 bg-white rounded-2xl ${
        onClick !== undefined ? "cursor-pointer" : ""
      }`}
      onClick={onClick}
    >
      <img
        // src="https://i.pinimg.com/736x/a8/57/00/a85700f3c614f6313750b9d8196c08f5.jpg"
        alt="Profile Image"
        src={pictureUrl}
        className="w-7 h-7 rounded-full"
      />
      <span className="text-black font-bold">{name}</span>
      {isHost && <span className="text-black ml-auto mr-2 italic">Host</span>}
      {isChecked && <CheckIcon className="w-6 h-6 text-secondary ml-auto" />}
    </div>
  );
};

export default PeopleBox;
