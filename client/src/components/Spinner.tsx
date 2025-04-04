interface SpinnerProps {
  bgClass?: string;
  spinnerColor?: string;
  spinnerSize?: {
    width: string;
    height: string;
  };
}

const Spinner: React.FC<SpinnerProps> = ({
  bgClass = "bg-inherit",
  spinnerColor = "border-gray-900",
  spinnerSize = { width: "w-32", height: "h-32" },
}) => (
  <div className={`h-full flex justify-center items-center ${bgClass}`}>
    <div
      className={`animate-spin rounded-full ${spinnerSize.height} ${spinnerSize.width} border-t-2 border-b-2 ${spinnerColor}`}
    ></div>
  </div>
);

export default Spinner;
