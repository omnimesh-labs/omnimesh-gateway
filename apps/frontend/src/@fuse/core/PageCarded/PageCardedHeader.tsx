import clsx from 'clsx';
import { ReactNode } from 'react';

/**
 * Props for the PageCardedHeader component.
 */
type PageCardedHeaderProps = {
	header?: ReactNode;
};

/**
 * The PageCardedHeader component is a header for the PageCarded component.
 */
function PageCardedHeader(props: PageCardedHeaderProps) {
	const { header = null } = props;

	return <div className={clsx('PageCarded-header', 'container')}>{header}</div>;
}

export default PageCardedHeader;
