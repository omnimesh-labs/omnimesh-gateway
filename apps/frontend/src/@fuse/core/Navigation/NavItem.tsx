import { NavItemType } from './types/NavItemType';
import components from './utils/components';

export type NavItemComponentProps = {
	type: string;
	item: NavItemType;
	dense?: boolean;
	nestedLevel?: number;
	onItemClick?: (T: NavItemType) => void;
	checkPermission?: boolean;
};

/**
Component to render NavItem depending on its type.
*/
function NavItem(props: NavItemComponentProps) {
	const { type } = props;

	const C = components[type];

	return C ? <C {...(props as object)} /> : null;
}

export default NavItem;
