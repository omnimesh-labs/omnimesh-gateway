import { NavItemType } from '@fuse/core/Navigation/types/NavItemType';

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
		id: 'llms',
		title: 'Models',
		type: 'item',
		icon: 'lucide:bot',
		url: '/llms'
	},
	{
		id: 'mcp-servers',
		title: 'MCP Servers',
		type: 'collapse',
		icon: 'lucide:server',
		children: [
			{
				id: 'servers',
				title: 'Servers',
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
			}
		]
	},
	{
		id: 'content',
		title: 'Content',
		type: 'collapse',
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
		id: 'security',
		title: 'Security',
		type: 'collapse',
		icon: 'lucide:shield-check',
		children: [
			{
				id: 'security-policies',
				title: 'Policies',
				type: 'item',
				icon: 'lucide:scroll',
				url: '/security/policies'
			},
			{
				id: 'security-filters',
				title: 'Content Filters',
				type: 'item',
				icon: 'lucide:shield-check',
				url: '/security/filters'
			},
			{
				id: 'security-auth',
				title: 'Auth Methods',
				type: 'item',
				icon: 'lucide:key',
				url: '/security/auth'
			}
		]
	},
	// {
	// 	id: 'a2a',
	// 	title: 'A2A Agents',
	// 	type: 'item',
	// 	icon: 'lucide:bot',
	// 	url: '/a2a'
	// },
	{
		id: 'settings',
		title: 'Settings',
		type: 'collapse',
		icon: 'lucide:computer',
		children: [
			{
				id: 'policies',
				title: 'Policies',
				type: 'item',
				icon: 'lucide:shield',
				url: '/policies'
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
			}
		]
	}
];

export default navigationConfig;
