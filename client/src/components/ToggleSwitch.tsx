interface ToggleSwitchProps {
  isOn: boolean;
  onToggle: () => void;
  option1: string;
  option2: string;
}

const ToggleSwitch = ({
  isOn,
  onToggle,
  option1,
  option2,
}: ToggleSwitchProps) => {
  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        onClick={onToggle}
        className={`inline-flex relative h-6 w-[4.5rem] px-2 items-center shadow-md rounded-full transition-colors duration-300 ${
          isOn ? "bg-secondary" : "bg-[#D9D9D9]"
        }`}
      >
        {/* Text label (fades in/out) */}
        <span
          className={`text-xs font-medium absolute transition-opacity duration-300 ${
            isOn ? "opacity-100 text-white" : "opacity-0"
          }`}
        >
          {option1}
        </span>
        <span
          className={`text-xs font-medium absolute left-8 transition-opacity duration-300 ${
            isOn ? "opacity-0" : "opacity-100 text-black"
          }`}
        >
          {option2}
        </span>

        {/* Toggle thumb (slides left/right) */}
        <span
          className={`inline-block h-4 w-4 rounded-full bg-white shadow-md transition-transform duration-300 ${
            isOn ? "translate-x-[calc(4.5rem-1.8rem)]" : "translate-x-0"
          }`}
        />
      </button>
    </div>
  );
};

export default ToggleSwitch;
