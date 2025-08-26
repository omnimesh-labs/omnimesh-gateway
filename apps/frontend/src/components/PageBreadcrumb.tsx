'use client';

import Breadcrumbs, { BreadcrumbsProps } from '@mui/material/Breadcrumbs';
import { NavItemType } from '@fuse/core/Navigation/types/NavItemType';
import usePathname from '@fuse/hooks/usePathname';

import Typography from '@mui/material/Typography';
import clsx from 'clsx';
import Link from '@fuse/core/Link';
import useNavigationItems from './theme-layouts/components/navigation/hooks/useNavigationItems';

type PageBreadcrumbProps = BreadcrumbsProps & {
	className?: string;
	skipHome?: boolean;
	borderColor?: string;
};

// Function to get the navigation item based on URL
function getNavigationItem(url: string, navigationItems: NavItemType[]): NavItemType {
	for (const item of navigationItems) {
		if (item.url === url) {
			return item;
		}

		if (item.children) {
			const childItem = getNavigationItem(url, item.children);

			if (childItem) {
				return childItem;
			}
		}
	}
	return null;
}

function PageBreadcrumb(props: PageBreadcrumbProps) {
	const {
		className,
		skipHome = false,
		color = 'action.active',
		borderColor = 'divider',
		maxItems = 4,
		...rest
	} = props;
	const pathname = usePathname();
	const { data: navigation } = useNavigationItems();

	const crumbs = pathname
		.split('/')
		.filter(Boolean)
		.reduce(
			(acc: { title: string; url: string }[], part, index, array) => {
				const url = `/${array.slice(0, index + 1).join('/')}`;
				const navItem = getNavigationItem(url, navigation);
				const title = navItem?.title || part;

				acc.push({ title, url });
				return acc;
			},
			skipHome ? [] : [{ title: 'Home', url: '/' }]
		);

	return (
		<Breadcrumbs
			classes={{ ol: 'list-none m-0 p-0' }}
			className={clsx('border-1 flex w-fit rounded-sm px-2', className)}
			sx={{ borderColor: borderColor + '!important' }}
			aria-label="breadcrumb"
			color={color}
			maxItems={maxItems}
			{...rest}
		>
			{crumbs.map((item, index) => (
				<Typography
					component={item.url ? Link : 'span'}
					to={item.url}
					key={index}
					className="block max-w-32 truncate text-md font-medium capitalize tracking-tight"
					role="button"
					color="inherit"
				>
					{item.title}
				</Typography>
			))}
		</Breadcrumbs>
	);
}

export default PageBreadcrumb;
