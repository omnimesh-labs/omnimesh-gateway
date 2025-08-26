import SidePanel from '@fuse/core/SidePanel';
import { memo } from 'react';
import NavigationShortcuts from '../../components/navigation/NavigationShortcuts';

/**
 * The left side layout 3.
 */
function LeftSideLayout3() {
	return (
		<SidePanel className="justify-star flex flex-col items-center gap-2 p-2">
			<div>
				<NavigationShortcuts
					className="m-1 flex shrink"
					variant="vertical"
				/>
			</div>
		</SidePanel>
	);
}

export default memo(LeftSideLayout3);
