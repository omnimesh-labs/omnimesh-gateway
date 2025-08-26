import { useMemo } from 'react';
import i18n from '@i18n';
import useUser from '@auth/useUser';
import useI18n from '@i18n/useI18n';
import Utils from '@fuse/utils';
import NavigationHelper from '@fuse/utils/NavigationHelper';
import { NavItemType } from '@fuse/core/Navigation/types/NavItemType';
import { useNavigationContext } from '../contexts/useNavigationContext';

function useNavigationItems() {
	const { navigationItems: navigationData } = useNavigationContext();

	const { data: user } = useUser();
	const userRole = user?.role;
	const { languageId } = useI18n();

	const data = useMemo(() => {
		const _navigation = NavigationHelper.unflattenNavigation(navigationData);

		function setAdditionalData(data: NavItemType[]): NavItemType[] {
			return data?.map((item) => ({
				hasPermission: Boolean(Utils.hasPermission(item?.auth, userRole)),
				...item,
				...(item?.translate && item?.title ? { title: i18n.t(`navigation:${item?.translate}`) } : {}),
				...(item?.children ? { children: setAdditionalData(item?.children) } : {})
			}));
		}

		const translatedValues = setAdditionalData(_navigation);

		return translatedValues;
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [navigationData, userRole, languageId]);

	const flattenData = useMemo(() => {
		return NavigationHelper.flattenNavigation(data);
	}, [data]);

	return { data, flattenData };
}

export default useNavigationItems;
