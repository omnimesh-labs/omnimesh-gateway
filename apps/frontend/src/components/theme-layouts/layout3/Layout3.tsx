import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import { memo, ReactNode } from 'react';
import Configurator from 'src/components/theme-layouts/components/configurator/Configurator';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import FooterLayout3 from './components/FooterLayout3';
import LeftSideLayout3 from './components/LeftSideLayout3';
import NavbarWrapperLayout3 from './components/NavbarWrapperLayout3';
import RightSideLayout3 from './components/RightSideLayout3';
import ToolbarLayout3 from './components/ToolbarLayout3';
import { Layout3ConfigDefaultsType } from './Layout3Config';

const Root = styled('div')(({ config }: { config: Layout3ConfigDefaultsType }) => ({
	...(config.mode === 'boxed' && {
		clipPath: 'inset(0)',
		maxWidth: `${config.containerWidth}px`,
		margin: '0 auto',
		boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)'
	}),
	...(config.mode === 'container' && {
		'& .container': {
			maxWidth: `${config.containerWidth}px`,
			width: '100%',
			margin: '0 auto',
			'@media (min-width: 96rem)': {
				maxWidth: `${config.containerWidth}px!important`
			}
		}
	}),
	...(config.mode === 'fullwidth' && {
		'& .container': {
			maxWidth: '100%!important',
			width: '100%!important'
		}
	})
}));

type Layout3Props = {
	children?: ReactNode;
};

/**
 * The layout 3.
 */
function Layout3(props: Layout3Props) {
	const { children } = props;

	const settings = useLayoutSettings();
	const config = settings.config as Layout3ConfigDefaultsType;

	return (
		<Root
			id="fuse-layout"
			className="flex w-full flex-auto"
			config={config}
		>
			{config.leftSidePanel.display && <LeftSideLayout3 />}

			<div className="flex min-w-0 flex-auto flex-col">
				<main
					id="fuse-main"
					className="relative flex min-h-svh min-w-0 flex-auto flex-col"
				>
					{config.navbar.display && (
						<NavbarWrapperLayout3
							className={clsx(config?.navbar?.style === 'fixed' ? 'sticky top-0 z-50' : '')}
						/>
					)}

					{config.toolbar.display && (
						<ToolbarLayout3
							className={clsx(
								config.toolbar.style === 'fixed' && 'sticky top-0',
								config.toolbar.position === 'above' && 'z-40 order-first'
							)}
						/>
					)}

					<div className="sticky top-0 z-99">
						<Configurator />
					</div>

					<div className="relative z-10 flex min-h-0 flex-auto flex-col">{children}</div>

					{config.footer.display && (
						<FooterLayout3 className={config.footer.style === 'fixed' ? 'sticky bottom-0' : ''} />
					)}
				</main>
			</div>

			{config.rightSidePanel.display && <RightSideLayout3 />}
		</Root>
	);
}

export default memo(Layout3);
