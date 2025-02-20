import { BrowserRouter as Router } from "react-router-dom";
import AppRouter from "./navigation/AppRouter";

const App = () => {
  return (
    <Router>
      <div className="min-h-dvh w-dvw flex items-center justify-center bg-gray-400">
        <div className="w-dvw xs:w-[376px] h-dvh border-0 xs:border-x-1 border-gray-800">
          <AppRouter />
        </div>
      </div>
    </Router>
  );
};

export default App;
