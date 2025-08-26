import Divider from '@mui/material/Divider';
import { memo } from 'react';
import GlobalStyles from '@mui/material/GlobalStyles';
import NavHorizontalLayout1 from './horizontal/NavHorizontalLayout1';
import NavVerticalLayout1 from './vertical/NavVerticalLayout1';
import NavVerticalLayout2 from './vertical/NavVerticalLayout2';
import NavHorizontalCollapse from './horizontal/types/NavHorizontalCollapse';
import NavHorizontalGroup from './horizontal/types/NavHorizontalGroup';
import NavHorizontalItem from './horizontal/types/NavHorizontalItem';
import NavHorizontalLink from './horizontal/types/NavHorizontalLink';
import NavVerticalCollapse from './vertical/types/NavVerticalCollapse';
import NavVerticalGroup from './vertical/types/NavVerticalGroup';
import NavVerticalItem from './vertical/types/NavVerticalItem';
import NavVerticalLink from './vertical/types/NavVerticalLink';
import { NavItemType } from './types/NavItemType';
import { registerComponent } from './utils/registerComponent';

const inputGlobalStyles = (
	<GlobalStyles
		styles={() => ({
			'.popper-navigation-list': {
				'& .fuse-list-item': {
					padding: '8px 12px 8px 12px',
					height: 36,
					minHeight: 36,
					'& .fuse-list-item-text': {
						padding: '0 0 0 8px'
					}
				},
				'&.dense': {
					'& .fuse-list-item': {
						minHeight: 32,
						height: 32,
						'& .fuse-list-item-text': {
							padding: '0 0 0 8px'
						}
					}
				}
			}
		})}
	/>
);

/*
Register  Navigation Components
 */
registerComponent('vertical-group', NavVerticalGroup);
registerComponent('vertical-collapse', NavVerticalCollapse);
registerComponent('vertical-item', NavVerticalItem);
registerComponent('vertical-link', NavVerticalLink);
registerComponent('horizontal-group', NavHorizontalGroup);
registerComponent('horizontal-collapse', NavHorizontalCollapse);
registerComponent('horizontal-item', NavHorizontalItem);
registerComponent('horizontal-link', NavHorizontalLink);
registerComponent('divider', () => <Divider className="my-4" />);
registerComponent('vertical-divider', () => <Divider className="my-4" />);
registerComponent('horizontal-divider', () => <Divider className="my-4" />);

export type NavigationProps = {
	className?: string;
	dense?: boolean;
	active?: boolean;
	onItemClick?: (T: NavItemType) => void;
	navigation?: NavItemType[];
	layout?: 'horizontal' | 'vertical' | 'vertical-2';
	firstLevel?: boolean;
	selectedId?: string;
	checkPermission?: boolean;
};

/**
 * Navigation
 * Component for displaying a navigation bar which contains NavItem components
 * and acts as parent for providing props to its children components
 */
function Navigation(props: NavigationProps) {
	const { navigation, layout = 'vertical' } = props;

	if (!navigation || navigation.length === 0) {
		return null;
	}

	return (
		<>
			{inputGlobalStyles}
			{layout === 'horizontal' && (
				<NavHorizontalLayout1
					checkPermission={false}
					{...props}
				/>
			)}
			{layout === 'vertical' && (
				<NavVerticalLayout1
					checkPermission={false}
					{...props}
				/>
			)}
			{layout === 'vertical-2' && (
				<NavVerticalLayout2
					checkPermission={false}
					{...props}
				/>
			)}
		</>
	);
}

export default memo(Navigation);
