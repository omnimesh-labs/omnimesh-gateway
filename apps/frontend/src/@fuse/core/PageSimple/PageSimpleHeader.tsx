import clsx from 'clsx';
import { ReactNode } from 'react';

/**
 * Props for the PageSimpleHeader component.
 */
type PageSimpleHeaderProps = {
	className?: string;
	header?: ReactNode;
};

/**
 * The PageSimpleHeader component is a sub-component of the PageSimple layout component.
 * It provides a header area for the layout.
 */
function PageSimpleHeader(props: PageSimpleHeaderProps) {
	const { header = null, className } = props;
	return (
		<div className={clsx('PageSimple-header', className)}>
			<div className="container">{header}</div>
		</div>
	);
}

export default PageSimpleHeader;
