import { createContext } from 'react';

// Use the same example data from the slice
const exampleData = {
	notes: [
		{
			id: 1,
			title: 'Best songs to listen while working',
			detail: 'Last edit: May 8th, 2015'
		},
		{
			id: 2,
			title: 'Useful subreddits',
			detail: 'Last edit: January 12th, 2015'
		}
	],
	events: [
		{
			id: 1,
			title: 'Group Meeting',
			detail: 'In 32 Minutes, Room 1B'
		},
		{
			id: 2,
			title: 'Public Beta Release',
			detail: '11:00 PM'
		},
		{
			id: 3,
			title: 'Dinner with David',
			detail: '17:30 PM'
		},
		{
			id: 4,
			title: 'Q&A Session',
			detail: '20:30 PM'
		}
	]
};

export type QuickPanelData = typeof exampleData;

export interface QuickPanelContextType {
	data: QuickPanelData;
	open: boolean;
	removeEvents: () => void;
	toggleQuickPanel: () => void;
	openQuickPanel: () => void;
	closeQuickPanel: () => void;
}

export const QuickPanelContext = createContext<QuickPanelContextType>({
	data: exampleData,
	open: false,
	removeEvents: () => {},
	toggleQuickPanel: () => {},
	openQuickPanel: () => {},
	closeQuickPanel: () => {}
});
