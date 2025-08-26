import Scrollbars from '@fuse/core/Scrollbars';
import IconButton from '@mui/material/IconButton';
import SvgIcon from '@fuse/core/SvgIcon';
import Typography from '@mui/material/Typography';
import ThemeSelector from '@fuse/core/ThemeSelector/ThemeSelector';
import { styled, useTheme } from '@mui/material/styles';
import Dialog from '@mui/material/Dialog';
import Slide from '@mui/material/Slide';
import { SwipeableHandlers } from 'react-swipeable';
import themeOptions from 'src/configs/themeOptions';
import { ThemeOption } from '@fuse/core/ThemeSelector/ThemePreview';
import useSettings from '@fuse/core/Settings/hooks/useSettings';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';
const StyledDialog = styled(Dialog)(({ theme }) => ({
	'& .MuiDialog-paper': {
		position: 'fixed',
		width: '100%',
		maxWidth: '40%',
		[theme.breakpoints.down('md')]: {
			maxWidth: '90%'
		},
		backgroundColor: theme.vars.palette.background.paper,
		top: 0,
		height: '100%',
		minHeight: '100%',
		bottom: 0,
		right: 0,
		margin: 0,
		zIndex: 1000,
		borderRadius: 0
	}
}));

type TransitionProps = {
	children?: React.ReactElement;
	ref?: React.RefObject<HTMLDivElement>;
};

function Transition(props: TransitionProps) {
	const { children, ref, ...other } = props;

	const theme = useTheme();

	if (!children) {
		return null;
	}

	return (
		<Slide
			direction={theme.direction === 'ltr' ? 'left' : 'right'}
			ref={ref}
			{...other}
		>
			{children}
		</Slide>
	);
}

type ThemesPanelProps = {
	schemesHandlers: SwipeableHandlers;
	onClose: () => void;
	open: boolean;
};

function ThemesPanel(props: ThemesPanelProps) {
	const { schemesHandlers, onClose, open } = props;
	const { setSettings } = useSettings();
	// const { isGuest, updateUserSettings } = useUser();
	// const { enqueueSnackbar } = useSnackbar();

	async function handleThemeSelect(_theme: ThemeOption) {
		const _newSettings = setSettings({ theme: { ..._theme?.section } } as Partial<SettingsConfigType>);

		/**
		 * Updating user settings disabled for demonstration purposes
		 * The request is made to the mock API and will not persist the changes
		 * You can enable it by removing the comment block below when using a real API
		 * */
		/* if (!isGuest) {
			const updatedUserData = await updateUserSettings(_newSettings);

			if (updatedUserData) {
				enqueueSnackbar('User settings saved.', {
					variant: 'success'
				});
			}
		}*/
	}

	return (
		<StyledDialog
			slots={{
				transition: Transition
			}}
			aria-labelledby="schemes-panel"
			aria-describedby="schemes"
			open={open}
			onClose={onClose}
			slotProps={{
				backdrop: {
					invisible: true
				}
			}}
			fullScreen
			classes={{
				paper: 'shadow-lg'
			}}
			disableRestoreFocus
			{...schemesHandlers}
		>
			<Scrollbars className="p-4 sm:p-6">
				<IconButton
					className="fixed top-0 z-10 ltr:right-0 rtl:left-0"
					onClick={onClose}
					size="large"
				>
					<SvgIcon>lucide:x</SvgIcon>
				</IconButton>

				<Typography
					className="mb-8"
					variant="h6"
				>
					Theme Color Options
				</Typography>

				<Typography
					className="text-md mb-6 text-justify italic"
					color="text.secondary"
				>
					* Selected option will be applied to all layout elements (navbar, toolbar, etc.). You can also
					create your own theme options and color schemes.
				</Typography>

				<ThemeSelector
					options={themeOptions}
					onSelect={handleThemeSelect}
				/>
			</Scrollbars>
		</StyledDialog>
	);
}

export default ThemesPanel;
