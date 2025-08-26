'use client';
import Search from '@fuse/core/Search';
import useNavigationItems from './hooks/useNavigationItems';

type NavigationSearchProps = {
	className?: string;
	variant?: 'basic' | 'full';
};

/**
 * The navigation search.
 */
function NavigationSearch(props: NavigationSearchProps) {
	const { variant, className } = props;
	const { flattenData: navigation } = useNavigationItems();

	return (
		<Search
			className={className}
			variant={variant}
			navigation={navigation}
		/>
	);
}

export default NavigationSearch;
