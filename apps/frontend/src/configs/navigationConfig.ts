import i18n from '@i18n';
import { NavItemType } from '@fuse/core/Navigation/types/NavItemType';
import ar from './navigation-i18n/ar';
import en from './navigation-i18n/en';
import tr from './navigation-i18n/tr';

i18n.addResourceBundle('en', 'navigation', en);
i18n.addResourceBundle('tr', 'navigation', tr);
i18n.addResourceBundle('ar', 'navigation', ar);

/**
 * The navigationConfig object is an array of navigation items for the MCP Gateway application.
 */
const navigationConfig: NavItemType[] = [
	{
		id: 'dashboard',
		title: 'Dashboard',
		type: 'item',
		icon: 'lucide:house',
		url: '/'
	},
	{
		id: 'servers',
		title: 'Server Management',
		type: 'item',
		icon: 'lucide:server',
		url: '/servers'
	},
	{
		id: 'namespaces',
		title: 'Namespaces',
		type: 'item',
		icon: 'lucide:folder',
		url: '/namespaces'
	},
	{
		id: 'endpoints',
		title: 'Endpoints',
		type: 'item',
		icon: 'lucide:globe',
		url: '/endpoints'
	},
	{
		id: 'content',
		title: 'Content Management',
		type: 'collapsible',
		icon: 'lucide:layers',
		children: [
			{
				id: 'content-tools',
				title: 'Tools',
				type: 'item',
				icon: 'lucide:wrench',
				url: '/content/tools'
			},
			{
				id: 'content-prompts',
				title: 'Prompts',
				type: 'item',
				icon: 'lucide:message-square',
				url: '/content/prompts'
			},
			{
				id: 'content-resources',
				title: 'Resources',
				type: 'item',
				icon: 'lucide:database',
				url: '/content/resources'
			}
		]
	},
	{
		id: 'policies',
		title: 'Policy Management',
		type: 'item',
		icon: 'lucide:shield',
		url: '/policies'
	},
	{
		id: 'a2a',
		title: 'A2A Agents',
		type: 'item',
		icon: 'lucide:bot',
		url: '/a2a'
	},
	{
		id: 'configuration',
		title: 'Configuration',
		type: 'item',
		icon: 'lucide:settings',
		url: '/configuration'
	},
	{
		id: 'logs',
		title: 'Logging & Audit',
		type: 'item',
		icon: 'lucide:scroll-text',
		url: '/logs'
	},
	{
		id: 'profile',
		title: 'Profile',
		type: 'collapsible',
		icon: 'lucide:user',
		children: [
			{
				id: 'profile-overview',
				title: 'Overview',
				type: 'item',
				icon: 'lucide:user',
				url: '/profile'
			},
			{
				id: 'profile-settings',
				title: 'Settings',
				type: 'item',
				icon: 'lucide:user-cog',
				url: '/profile/settings'
			},
			{
				id: 'profile-api-keys',
				title: 'API Keys',
				type: 'item',
				icon: 'lucide:key',
				url: '/profile/api-keys'
			}
		]
	}
];

export default navigationConfig;
