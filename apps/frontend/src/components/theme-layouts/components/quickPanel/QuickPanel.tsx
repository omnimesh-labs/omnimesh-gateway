import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import Divider from '@mui/material/Divider';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import ListSubheader from '@mui/material/ListSubheader';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import Switch from '@mui/material/Switch';
import Typography from '@mui/material/Typography';
import { format } from 'date-fns/format';
import { useState } from 'react';
import SvgIcon from '@fuse/core/SvgIcon';
import { useQuickPanelContext } from './contexts/QuickPanelContext/useQuickPanelContext';

const StyledSwipeableDrawer = styled(SwipeableDrawer)(() => ({
	'& .MuiDrawer-paper': {
		width: 280
	}
}));

/**
 * The quick panel.
 */
function QuickPanel() {
	const { data, open, toggleQuickPanel } = useQuickPanelContext();

	const [checked, setChecked] = useState<string[]>(['notifications']);

	const handleToggle = (value: string) => () => {
		const currentIndex = checked.indexOf(value);
		const newChecked = [...checked];

		if (currentIndex === -1) {
			newChecked.push(value);
		} else {
			newChecked.splice(currentIndex, 1);
		}

		setChecked(newChecked);
	};

	return (
		<StyledSwipeableDrawer
			open={open}
			anchor="right"
			onOpen={() => {}}
			onClose={() => toggleQuickPanel()}
			disableSwipeToOpen
		>
			<Scrollbars>
				<ListSubheader component="div">Today</ListSubheader>

				<div className="mb-0 px-6 py-4">
					<Typography
						className="mb-3 text-5xl"
						color="text.secondary"
					>
						{format(new Date(), 'eeee')}
					</Typography>
					<div className="flex">
						<Typography
							className="text-5xl leading-none"
							color="text.secondary"
						>
							{format(new Date(), 'dd')}
						</Typography>
						<Typography
							className="text-lg leading-none"
							color="text.secondary"
						>
							th
						</Typography>
						<Typography
							className="text-5xl leading-none"
							color="text.secondary"
						>
							{format(new Date(), 'MMMM')}
						</Typography>
					</div>
				</div>
				<Divider />
				<List>
					<ListSubheader component="div">Events</ListSubheader>
					{data &&
						data.events.map((event) => (
							<ListItem key={event.id}>
								<ListItemText
									primary={event.title}
									secondary={event.detail}
								/>
							</ListItem>
						))}
				</List>
				<Divider />
				<List>
					<ListSubheader component="div">Notes</ListSubheader>
					{data &&
						data.notes.map((note) => (
							<ListItem key={note.id}>
								<ListItemText
									primary={note.title}
									secondary={note.detail}
								/>
							</ListItem>
						))}
				</List>
				<Divider />
				<List>
					<ListSubheader component="div">Quick Settings</ListSubheader>
					<ListItem
						secondaryAction={
							<Switch
								color="primary"
								onChange={handleToggle('notifications')}
								checked={checked.indexOf('notifications') !== -1}
							/>
						}
					>
						<ListItemIcon className="min-w-9">
							<SvgIcon>lucide:bell</SvgIcon>
						</ListItemIcon>
						<ListItemText primary="Notifications" />
					</ListItem>
					<ListItem
						secondaryAction={
							<Switch
								color="secondary"
								onChange={handleToggle('cloudSync')}
								checked={checked.indexOf('cloudSync') !== -1}
							/>
						}
					>
						<ListItemIcon className="min-w-9">
							<SvgIcon>lucide:cloud</SvgIcon>
						</ListItemIcon>
						<ListItemText primary="Cloud Sync" />
					</ListItem>
					<ListItem
						secondaryAction={
							<Switch
								color="primary"
								onChange={handleToggle('retroThrusters')}
								checked={checked.indexOf('retroThrusters') !== -1}
							/>
						}
					>
						<ListItemIcon className="min-w-9">
							<SvgIcon>lucide:sun</SvgIcon>
						</ListItemIcon>
						<ListItemText primary="Retro Thrusters" />
					</ListItem>
				</List>
			</Scrollbars>
		</StyledSwipeableDrawer>
	);
}

export default QuickPanel;
