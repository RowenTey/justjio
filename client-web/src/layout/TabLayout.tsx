import { Outlet } from "react-router-dom";
import TabBar from "../navigation/TabBar";

const TabLayout = () => {
  return (
    <div className="h-full">
      <div className="h-[92.5%]">
        <Outlet />
      </div>
      <TabBar />
    </div>
  );
};

export default TabLayout;
