import React from "react";
import { Link } from "react-router-dom";

interface LinkProps {
  to: string;
  from: string;
  state?: object;
}

interface ButtonCardProps {
  title: string;
  Icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  numNotifications?: number;
  onClick?: () => void;
  isLink?: boolean;
  linkProps?: LinkProps;
}

const ButtonCard: React.FC<ButtonCardProps> = ({
  title,
  Icon,
  numNotifications,
  onClick = () => {},
  isLink = true,
  linkProps: { to, from, state } = { to: "/", from: "/", state: {} },
}) => {
  const shouldShowNotification = numNotifications && numNotifications > 0;

  const content = (
    <div
      className="relative flex flex-col items-center justify-center w-12 cursor-pointer"
      onClick={isLink ? undefined : onClick}
    >
      <div className="flex items-center justify-center w-12 h-12 p-1 bg-secondary rounded-lg hover:shadow-lg hover:border hover:border-white">
        <Icon className="w-8 h-8 text-white font-bold" />
      </div>

      <p className="text-sm text-center leading-[1.1] font-medium text-black text-wrap mt-1">
        {title}
      </p>

      {shouldShowNotification === true ? (
        <div className="absolute -top-1 -right-1 w-4 h-4 bg-red-600 rounded-full flex items-center justify-center text-white text-xs font-bold p-1">
          <span>{numNotifications}</span>
        </div>
      ) : null}
    </div>
  );

  return isLink ? (
    <Link to={to} state={{ from, ...state }}>
      {content}
    </Link>
  ) : (
    content
  );
};

export default ButtonCard;
