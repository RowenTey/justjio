import { useEffect, useState } from "react";
import { countPendingFriendRequestsApi } from "../../api/user";
import { api } from "../../api";
import { Link, useNavigate } from "react-router-dom";
import { ArrowLeftIcon, UserGroupIcon } from "@heroicons/react/24/solid";

type FriendsTopBarProps = {
  title: string;
  userId: number;
};

const FriendsTopBar: React.FC<FriendsTopBarProps> = ({ userId, title }) => {
  const navigate = useNavigate();
  const [numFriendRequests, setNumFriendRequests] = useState(0);

  useEffect(() => {
    const fetchNumFriendRequests = async () => {
      const res = await countPendingFriendRequestsApi(api, userId);
      setNumFriendRequests(res.data.data.count);
    };

    fetchNumFriendRequests();
  }, [userId]);

  return (
    <div
      className={`relative top-0 flex h-[8%] items-center w-full py-4 px-3 bg-purple-200 justify-between`}
    >
      <button
        onClick={() => navigate(-1)}
        className={`flex items-center justify-center bg-transparent p-1 hover:scale-110 `}
      >
        <ArrowLeftIcon className="w-6 h-6 text-black" />
      </button>

      <h1 className="text-xl font-bold text-secondary">{title}</h1>

      <Link
        to="/friends/requests"
        state={{ from: "/friends" }}
        className={`flex items-center justify-center bg-transparent p-1`}
      >
        <UserGroupIcon className="w-8 h-8 text-secondary hover:scale-110" />
        {numFriendRequests > 0 && (
          <div className="absolute top-[4px] right-[7px] w-4 h-4 bg-red-600 rounded-full flex items-center justify-center text-white text-xs font-bold p-1">
            {numFriendRequests}
          </div>
        )}
      </Link>
    </div>
  );
};

export default FriendsTopBar;
