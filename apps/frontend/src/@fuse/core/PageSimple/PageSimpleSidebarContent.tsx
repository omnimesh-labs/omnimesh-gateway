import Scrollbars from '@fuse/core/Scrollbars';
import { ReactNode } from 'react';

/**
 * Props for the PageSimpleSidebarContent component.
 */
type PageSimpleSidebarContentProps = {
	innerScroll?: boolean;
	children?: ReactNode;
	content?: ReactNode;
};

/**
 * The PageSimpleSidebarContent component is a content container for the PageSimpleSidebar component.
 */
function PageSimpleSidebarContent(props: PageSimpleSidebarContentProps) {
	const { innerScroll, children, content } = props;

	if (!children && !content) {
		return null;
	}

	return (
		<Scrollbars enable={innerScroll}>
			<div className="PageSimple-sidebarContent flex min-h-full flex-col lg:min-w-0">
				{content || children}
			</div>
		</Scrollbars>
	);
}

export default PageSimpleSidebarContent;
