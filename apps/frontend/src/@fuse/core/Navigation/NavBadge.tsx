import clsx from 'clsx';
import { memo } from 'react';
import { NavBadgeType } from './types/NavBadgeType';
import Chip from '@mui/material/Chip';

type NavBadgeProps = {
	className?: string;
	classes?: string;
	badge: NavBadgeType;
};

/**
 * NavBadge component.
 * This component will render a badge on a Nav element. It accepts a `NavBadgeType` as a prop,
 * which is an object containing a title and background and foreground colour.
 */
function NavBadge(props: NavBadgeProps) {
	const { className = '', badge } = props;

	return (
		<Chip
			className={clsx('item-badge truncate text-xs leading-none font-bold', className)}
			size="small"
			color="secondary"
			sx={{
				backgroundColor: badge.bg,
				color: badge.fg
			}}
			label={badge.title}
		/>
	);
}

export default memo(NavBadge);
