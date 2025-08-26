'use client';
import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import { memo, ReactNode, useImperativeHandle, useRef, RefObject } from 'react';
import GlobalStyles from '@mui/material/GlobalStyles';
import { SystemStyleObject, Theme } from '@mui/system';
import PageCardedSidebar, { PageCardedSidebarProps } from './PageCardedSidebar';
import PageCardedHeader from './PageCardedHeader';
import { ScrollbarsProps } from '../Scrollbars/Scrollbars';

const headerHeight = 120;
const toolbarHeight = 64;

type PageCardedProps = SystemStyleObject<Theme> & {
	className?: string;
	header?: ReactNode;
	content?: ReactNode;
	leftSidebarProps?: PageCardedSidebarProps;
	rightSidebarProps?: PageCardedSidebarProps;
	scroll?: 'normal' | 'page' | 'content';
	contentScrollbarsProps?: ScrollbarsProps;
	ref?: RefObject<{ toggleLeftSidebar: (val: boolean) => void; toggleRightSidebar: (val: boolean) => void }>;
};

const Root = styled('div')<PageCardedProps>(({ theme, ...props }) => ({
	display: 'flex',
	flexDirection: 'column',
	minWidth: 0,
	minHeight: '100%',
	position: 'relative',
	flex: '1 1 auto',
	width: '100%',
	height: 'auto',
	padding: '0 16px',
	backgroundColor: theme.vars.palette.background.default,

	'& .PageCarded-scroll-content': {
		height: '100%'
	},

	'& .PageCarded-wrapper': {
		display: 'flex',
		flexDirection: 'row',
		flex: '1 1 auto',
		zIndex: 2,
		maxWidth: '100%',
		minWidth: 0,
		height: '100%',
		backgroundColor: theme.vars.palette.background.paper,
		boxShadow: theme.vars.shadows[2],
		borderRadius: '12px 12px 0 0',
		margin: '2px 2px 0 2px',
		...(props.scroll === 'content' && {
			position: 'absolute',
			top: 0,
			bottom: 0,
			right: 0,
			left: 0,
			overflow: 'hidden'
		})
	},

	'& .PageCarded-header': {
		display: 'flex',
		flex: '0 0 auto'
	},

	'& .PageCarded-contentWrapper': {
		display: 'flex',
		flexDirection: 'column',
		flex: '1 1 auto',
		overflow: 'auto',
		WebkitOverflowScrolling: 'touch',
		zIndex: 9999
	},

	'& .PageCarded-toolbar': {
		height: toolbarHeight,
		minHeight: toolbarHeight,
		display: 'flex',
		alignItems: 'center'
	},

	'& .PageCarded-content': {
		flex: '1 0 auto'
	},

	'& .PageCarded-sidebarWrapper': {
		overflow: 'hidden',
		backgroundColor: 'transparent',
		position: 'absolute',
		'&.permanent': {
			[theme.breakpoints.up('lg')]: {
				position: 'relative',
				marginLeft: 0,
				marginRight: 0,
				transition: theme.transitions.create('margin', {
					easing: theme.transitions.easing.sharp,
					duration: theme.transitions.duration.leavingScreen
				}),
				'&.closed': {
					transition: theme.transitions.create('margin', {
						easing: theme.transitions.easing.easeOut,
						duration: theme.transitions.duration.enteringScreen
					}),

					'&.PageCarded-leftSidebar': {
						marginLeft: -props.leftSidebarProps?.width
					},
					'&.PageCarded-rightSidebar': {
						marginRight: -props.rightSidebarProps?.width
					}
				}
			}
		}
	},

	'& .PageCarded-sidebar': {
		position: 'absolute',
		backgroundColor: theme.vars.palette.background.paper,
		color: theme.vars.palette.text.primary,

		'&.permanent': {
			[theme.breakpoints.up('lg')]: {
				position: 'relative'
			}
		},
		maxWidth: '100%',
		height: '100%'
	},

	'& .PageCarded-leftSidebar': {
		width: props.leftSidebarProps?.width,

		[theme.breakpoints.up('lg')]: {
			// borderRight: `1px solid ${theme.vars.palette.divider}`,
			// borderLeft: 0,
		}
	},

	'& .PageCarded-rightSidebar': {
		width: props.rightSidebarProps?.width,

		[theme.breakpoints.up('lg')]: {
			// borderLeft: `1px solid ${theme.vars.palette.divider}`,
			// borderRight: 0,
		}
	},

	'& .PageCarded-sidebarHeader': {
		height: headerHeight,
		minHeight: headerHeight,
		backgroundColor: theme.vars.palette.primary.dark,
		color: theme.vars.palette.primary.contrastText
	},

	'& .PageCarded-sidebarHeaderInnerSidebar': {
		backgroundColor: 'transparent',
		color: 'inherit',
		height: 'auto',
		minHeight: 'auto'
	},

	'& .PageCarded-sidebarContent': {
		display: 'flex',
		flexDirection: 'column',
		minHeight: '100%'
	},

	'& .PageCarded-backdrop': {
		position: 'absolute'
	}
}));

const sidebarPropsDefaults = { variant: 'permanent' as const };

function PageCarded(props: PageCardedProps) {
	const {
		scroll = 'page',
		className,
		header,
		content,
		leftSidebarProps,
		rightSidebarProps,
		contentScrollbarsProps,
		ref
	} = props;

	const leftSidebarRef = useRef<{ toggleSidebar: (T: boolean) => void }>(null);
	const rightSidebarRef = useRef<{ toggleSidebar: (T: boolean) => void }>(null);
	const rootRef = useRef(null);

	useImperativeHandle(ref, () => ({
		toggleLeftSidebar: (val: boolean) => {
			if (leftSidebarRef.current) {
				leftSidebarRef.current.toggleSidebar(val);
			}
		},
		toggleRightSidebar: (val: boolean) => {
			if (rightSidebarRef.current) {
				rightSidebarRef.current.toggleSidebar(val);
			}
		}
	}));

	return (
		<>
			<GlobalStyles
				styles={() => ({
					...(scroll !== 'page' && {
						'#fuse-toolbar': {
							position: 'static!important'
						},
						'#fuse-footer': {
							position: 'static!important'
						}
					}),
					...(scroll === 'page' && {
						'#fuse-toolbar': {
							position: 'sticky',
							top: 0
						},
						'#fuse-footer': {
							position: 'sticky',
							bottom: 0
						}
					})
				})}
			/>
			<Root
				className={clsx('PageCarded-root', `PageCarded-scroll-${scroll}`, className)}
				ref={rootRef}
				scroll={scroll}
				leftSidebarProps={{ ...sidebarPropsDefaults, ...leftSidebarProps }}
				rightSidebarProps={{ ...sidebarPropsDefaults, ...rightSidebarProps }}
			>
				{header && <PageCardedHeader header={header} />}

				<div className="relative z-10 container flex h-full flex-auto flex-col overflow-hidden">
					<div className="PageCarded-wrapper">
						<PageCardedSidebar
							position="left"
							ref={leftSidebarRef}
							{...sidebarPropsDefaults}
							{...leftSidebarProps}
						/>
						<Scrollbars
							className="PageCarded-contentWrapper"
							enable={scroll === 'content'}
							{...contentScrollbarsProps}
						>
							{content && <div className={clsx('PageCarded-content')}>{content}</div>}
						</Scrollbars>
						<PageCardedSidebar
							position="right"
							ref={rightSidebarRef}
							{...sidebarPropsDefaults}
							{...rightSidebarProps}
						/>
					</div>
				</div>
			</Root>
		</>
	);
}

const StyledPageCarded = memo(styled(PageCarded)``);

export default StyledPageCarded;
