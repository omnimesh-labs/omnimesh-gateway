import Scrollbars from '@fuse/core/Scrollbars';
import { ReactNode } from 'react';

/**
 * Props for the PageCardedSidebarContent component.
 */
type PageCardedSidebarContentProps = {
	innerScroll?: boolean;
	children?: ReactNode;
	content?: ReactNode;
};

/**
 * The PageCardedSidebarContent component is a content container for the PageCardedSidebar component.
 */
function PageCardedSidebarContent(props: PageCardedSidebarContentProps) {
	const { innerScroll, children, content } = props;

	if (!content && !children) {
		return null;
	}

	return (
		<Scrollbars enable={innerScroll}>
			<div className="PageCarded-sidebarContent lg:min-w-0">{content || children}</div>
		</Scrollbars>
	);
}

export default PageCardedSidebarContent;
