import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import Fab from '@mui/material/Fab';
import IconButton from '@mui/material/IconButton';
import Paper from '@mui/material/Paper';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import Tooltip from '@mui/material/Tooltip';
import clsx from 'clsx';
import { memo, ReactNode, useState } from 'react';
import SvgIcon from '../SvgIcon';
import useThemeMediaQuery from '../../hooks/useThemeMediaQuery';

const Root = styled('div')(({ theme }) => ({
	'& .SidePanel-paper': {
		display: 'flex',
		width: 56,
		transition: theme.transitions.create(['transform', 'width', 'min-width'], {
			easing: theme.transitions.easing.sharp,
			duration: theme.transitions.duration.shorter
		}),
		paddingBottom: 64,
		height: '100%',
		maxHeight: '100vh',
		position: 'sticky',
		top: 0,
		zIndex: 999,
		'&.left': {
			'& .SidePanel-buttonWrapper': {
				left: 0,
				right: 'auto'
			},
			'& .SidePanel-buttonIcon': {
				transform: 'rotate(0deg)'
			}
		},
		'&.right': {
			'& .SidePanel-buttonWrapper': {
				right: 0,
				left: 'auto'
			},
			'& .SidePanel-buttonIcon': {
				transform: 'rotate(-180deg)'
			}
		},
		'&.closed': {
			[theme.breakpoints.up('lg')]: {
				width: 0
			},
			'&.left': {
				'& .SidePanel-buttonWrapper': {
					justifyContent: 'start'
				},
				'& .SidePanel-button': {
					borderBottomLeftRadius: 0,
					borderTopLeftRadius: 0,
					paddingLeft: 4
				},
				'& .SidePanel-buttonIcon': {
					transform: 'rotate(-180deg)'
				}
			},
			'&.right': {
				'& .SidePanel-buttonWrapper': {
					justifyContent: 'flex-end'
				},
				'& .SidePanel-button': {
					borderBottomRightRadius: 0,
					borderTopRightRadius: 0,
					paddingRight: 4
				},
				'& .SidePanel-buttonIcon': {
					transform: 'rotate(0deg)'
				}
			},
			'& .SidePanel-buttonWrapper': {
				width: 'auto'
			},
			'& .SidePanel-button': {
				backgroundColor: theme.vars.palette.background.paper,
				borderRadius: 38,
				transition: theme.transitions.create(
					['background-color', 'border-radius', 'width', 'min-width', 'padding'],
					{
						easing: theme.transitions.easing.easeInOut,
						duration: theme.transitions.duration.shorter
					}
				),
				width: 24,
				'&:hover': {
					width: 52,
					paddingLeft: 8,
					paddingRight: 8
				}
			},
			'& .SidePanel-content': {
				opacity: 0
			}
		}
	},
	'& .SidePanel-content': {
		overflow: 'hidden',
		opacity: 1,
		transition: theme.transitions.create(['opacity'], {
			easing: theme.transitions.easing.easeInOut,
			duration: theme.transitions.duration.short
		})
	},
	'& .SidePanel-buttonWrapper': {
		position: 'absolute',
		bottom: 0,
		left: 0,
		display: 'flex',
		alignItems: 'center',
		justifyContent: 'center',
		padding: '12px 0',
		width: '100%',
		minWidth: 56
	},
	'& .SidePanel-button': {
		padding: 8,
		width: 40,
		height: 40
	},
	'& .SidePanel-buttonIcon': {
		transition: theme.transitions.create(['transform'], {
			easing: theme.transitions.easing.easeInOut,
			duration: theme.transitions.duration.short
		})
	},
	'& .SidePanel-mobileButton': {
		height: 40,
		position: 'fixed',
		zIndex: 99,
		bottom: 12,
		width: 24,
		borderRadius: 38,
		padding: 8,
		backgroundColor: theme.vars.palette.background.paper,
		transition: theme.transitions.create(['background-color', 'border-radius', 'width', 'min-width', 'padding'], {
			easing: theme.transitions.easing.easeInOut,
			duration: theme.transitions.duration.shorter
		}),
		'&:hover': {
			width: 52,
			paddingLeft: 8,
			paddingRight: 8
		},
		'&.left': {
			borderBottomLeftRadius: 0,
			borderTopLeftRadius: 0,
			paddingLeft: 4,
			left: 0
		},
		'&.right': {
			borderBottomRightRadius: 0,
			borderTopRightRadius: 0,
			paddingRight: 4,
			right: 0,
			'& .SidePanel-buttonIcon': {
				transform: 'rotate(-180deg)'
			}
		}
	}
}));

type SidePanelProps = {
	position?: 'left';
	opened?: true;
	className?: string;
	children?: ReactNode;
};

/**
 * The SidePanel component is responsible for rendering a side panel that can be opened and closed.
 * It uses various MUI components to render the panel and its contents.
 * The component is memoized to prevent unnecessary re-renders.
 */
function SidePanel(props: SidePanelProps) {
	const { position = 'left', opened = true, className, children } = props;
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	const [panelOpened, setPanelOpened] = useState(Boolean(opened));
	const [mobileOpen, setMobileOpen] = useState(false);

	function toggleOpened() {
		setPanelOpened(!panelOpened);
	}

	function toggleMobileDrawer() {
		setMobileOpen(!mobileOpen);
	}

	return (
		<Root>
			{!isMobile && (
				<Paper
					className={clsx(
						'SidePanel-paper',
						className,
						panelOpened ? 'opened' : 'closed',
						position,
						'shadow-lg'
					)}
					square
				>
					<Scrollbars className={clsx('content', 'SidePanel-content')}>{children}</Scrollbars>

					<div className="SidePanel-buttonWrapper">
						<Tooltip
							title="Toggle side panel"
							placement={position === 'left' ? 'right' : 'right'}
						>
							<IconButton
								className="SidePanel-button"
								onClick={toggleOpened}
								disableRipple
								size="large"
							>
								<SvgIcon className="SidePanel-buttonIcon">lucide:chevron-left</SvgIcon>
							</IconButton>
						</Tooltip>
					</div>
				</Paper>
			)}

			{isMobile && (
				<>
					<SwipeableDrawer
						classes={{
							paper: clsx('SidePanel-paper', className)
						}}
						anchor={position}
						open={mobileOpen}
						onOpen={() => {}}
						onClose={toggleMobileDrawer}
						disableSwipeToOpen
					>
						<Scrollbars className={clsx('content', 'SidePanel-content')}>{children}</Scrollbars>
					</SwipeableDrawer>

					<Tooltip
						title="Hide side panel"
						placement={position === 'left' ? 'right' : 'right'}
					>
						<Fab
							className={clsx('SidePanel-mobileButton', position)}
							onClick={toggleMobileDrawer}
							disableRipple
						>
							<SvgIcon className="SidePanel-buttonIcon">lucide:chevron-right</SvgIcon>
						</Fab>
					</Tooltip>
				</>
			)}
		</Root>
	);
}

export default memo(SidePanel);
