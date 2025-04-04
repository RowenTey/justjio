import { Link } from "react-router-dom";
import {
  BellIcon as BellSolid,
  HomeIcon as HomeSolid,
  UserIcon as UserSolid,
} from "@heroicons/react/24/solid";
import {
  BellIcon as BellOutline,
  HomeIcon as HomeOutline,
  UserIcon as UserOutline,
} from "@heroicons/react/24/outline";
import NavIcon from "../components/NavIcon";

const TabBar = () => {
  return (
    <nav className="h-[7.5%] p-2 bg-primary flex items-center justify-evenly">
      <Link to="/" state={{ from: "/" }}>
        <NavIcon to="/" SolidIcon={HomeSolid} OutlineIcon={HomeOutline} />
      </Link>
      <Link to="/profile" state={{ from: "/" }}>
        <NavIcon
          to="/profile"
          SolidIcon={UserSolid}
          OutlineIcon={UserOutline}
        />
      </Link>
      <Link to="/notifications" state={{ from: "/" }}>
        <NavIcon
          to="/notifications"
          SolidIcon={BellSolid}
          OutlineIcon={BellOutline}
        />
      </Link>
    </nav>
  );
};

export default TabBar;
