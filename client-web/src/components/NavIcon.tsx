import { useLocation } from "react-router-dom";

interface NavIconProps {
	to: string;
	SolidIcon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
	OutlineIcon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
}

const NavIcon: React.FC<NavIconProps> = ({ to, SolidIcon, OutlineIcon }) => {
	const location = useLocation();
	const isActive = location.pathname === to;

	return (
		<div className="mx-2 text-secondary hover:underline flex items-center">
			{isActive ? (
				<SolidIcon className="h-6 w-6 hover:scale-110" />
			) : (
				<OutlineIcon className="h-6 w-6 hover:scale-110" />
			)}
		</div>
	);
};

export default NavIcon;
