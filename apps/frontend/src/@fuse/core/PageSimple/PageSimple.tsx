'use client';
import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import { memo, ReactNode, RefObject, useImperativeHandle, useRef } from 'react';
import GlobalStyles from '@mui/material/GlobalStyles';
import { SystemStyleObject, Theme } from '@mui/system';
import PageSimpleHeader from './PageSimpleHeader';
import PageSimpleSidebar, { PageSimpleSidebarProps } from './PageSimpleSidebar';
import { ScrollbarsProps } from '../Scrollbars/Scrollbars';

const headerHeight = 120;
const toolbarHeight = 64;

/**
 * Props for the PageSimple component.
 */
type PageSimpleProps = SystemStyleObject<Theme> & {
	className?: string;
	header?: ReactNode;
	content?: ReactNode;
	scroll?: 'normal' | 'page' | 'content';
	leftSidebarProps?: PageSimpleSidebarProps;
	rightSidebarProps?: PageSimpleSidebarProps;
	contentScrollbarsProps?: ScrollbarsProps;
	ref?: RefObject<{ toggleLeftSidebar: (val: boolean) => void; toggleRightSidebar: (val: boolean) => void }>;
};

/**
 * The Root styled component is the top-level container for the PageSimple component.
 */
const Root = styled('div')<PageSimpleProps>(({ theme, ...props }) => ({
	display: 'flex',
	flexDirection: 'column',
	minWidth: 0,
	minHeight: '100%',
	position: 'relative',
	flex: '1 1 auto',
	width: '100%',
	height: 'auto',
	backgroundColor: theme.vars.palette.background.default,

	'&.PageSimple-scroll-content': {
		height: '100%'
	},

	'& .PageSimple-wrapper': {
		display: 'flex',
		flexDirection: 'row',
		flex: '1 1 auto',
		zIndex: 2,
		minWidth: 0,
		height: '100%',
		backgroundColor: theme.vars.palette.background.default,

		...(props.scroll === 'content' && {
			position: 'absolute',
			top: 0,
			bottom: 0,
			right: 0,
			left: 0,
			overflow: 'hidden'
		})
	},

	'& .PageSimple-header': {
		display: 'flex',
		flex: '0 0 auto',
		backgroundSize: 'cover'
	},

	'& .PageSimple-topBg': {
		position: 'absolute',
		left: 0,
		right: 0,
		top: 0,
		height: headerHeight,
		pointerEvents: 'none'
	},

	'& .PageSimple-contentWrapper': {
		display: 'flex',
		flexDirection: 'column',
		width: '100%',
		flex: '1',
		overflow: 'hidden',
		//    WebkitOverflowScrolling: 'touch',
		zIndex: 9999
	},

	'& .PageSimple-toolbar': {
		height: toolbarHeight,
		minHeight: toolbarHeight,
		display: 'flex',
		alignItems: 'center'
	},

	'& .PageSimple-content': {
		display: 'flex',
		flexDirection: 'column',
		flex: '1 1 auto',
		alignItems: 'start',
		minHeight: 0,
		overflowY: 'auto',
		'& > .container': {
			display: 'flex',
			flexDirection: 'column',
			minHeight: '100%'
		}
	},

	'& .PageSimple-sidebarWrapper': {
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

					'&.PageSimple-leftSidebar': {
						marginLeft: -props.leftSidebarProps?.width
					},
					'&.PageSimple-rightSidebar': {
						marginRight: -props.rightSidebarProps?.width
					}
				}
			}
		}
	},

	'& .PageSimple-sidebar': {
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

	'& .PageSimple-leftSidebar': {
		width: props.leftSidebarProps?.width,
		maxWidth: props.leftSidebarProps?.width,

		[theme.breakpoints.up('lg')]: {
			borderRight: `1px solid ${theme.vars.palette.divider}`,
			borderLeft: 0
		}
	},

	'& .PageSimple-rightSidebar': {
		width: props.rightSidebarProps?.width,
		maxWidth: props.rightSidebarProps?.width,

		[theme.breakpoints.up('lg')]: {
			borderLeft: `1px solid ${theme.vars.palette.divider}`,
			borderRight: 0,
			flex: '1'
		}
	},

	'& .PageSimple-backdrop': {
		position: 'absolute'
	}
}));

const sidebarPropsDefaults = { variant: 'permanent' as const };

/**
 * The PageSimple component is a layout component that provides a simple page layout with a header, left sidebar, right sidebar, and content area.
 * It is designed to be used as a top-level component for an application or as a sub-component within a larger layout.
 */
function PageSimple(props: PageSimpleProps) {
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
		rootRef,
		toggleLeftSidebar: (val: boolean) => {
			leftSidebarRef?.current?.toggleSidebar(val);
		},
		toggleRightSidebar: (val: boolean) => {
			rightSidebarRef?.current?.toggleSidebar(val);
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
				className={clsx('PageSimple-root', `PageSimple-scroll-${scroll}`, className)}
				ref={rootRef}
				scroll={scroll}
				leftSidebarProps={{ ...sidebarPropsDefaults, ...leftSidebarProps }}
				rightSidebarProps={{ ...sidebarPropsDefaults, ...rightSidebarProps }}
			>
				<div className="z-10 flex h-full flex-auto flex-col">
					<div className="PageSimple-wrapper">
						<PageSimpleSidebar
							position="left"
							ref={leftSidebarRef}
							{...sidebarPropsDefaults}
							{...leftSidebarProps}
						/>
						<div
							className="PageSimple-contentWrapper"

							// enable={scroll === 'page'}
						>
							{header && <PageSimpleHeader header={header} />}

							{content && (
								<Scrollbars
									enable={scroll === 'content'}
									className={clsx('PageSimple-content')}
									scrollToTopOnRouteChange
									{...contentScrollbarsProps}
								>
									<div className="container">{content}</div>
								</Scrollbars>
							)}
						</div>
						<PageSimpleSidebar
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

const StyledPageSimple = memo(styled(PageSimple)``);

export default StyledPageSimple;
