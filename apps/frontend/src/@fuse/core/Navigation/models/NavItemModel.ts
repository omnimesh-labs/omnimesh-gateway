import _ from 'lodash';
import { PartialDeep } from 'type-fest';
import { NavItemType } from '../types/NavItemType';

/**
 *  NavItemModel
 *  Constructs a navigation item based on NavItemType
 */
function NavItemModel(data?: PartialDeep<NavItemType>) {
	data = data || {};

	return _.defaults(data, {
		id: _.uniqueId(),
		title: '',
		translate: '',
		auth: null,
		subtitle: '',
		icon: '',
		iconClass: '',
		url: '',
		target: '',
		type: 'item',
		sx: {},
		disabled: false,
		active: false,
		exact: false,
		end: false,
		badge: null,
		children: []
	});
}

export default NavItemModel;
