import clsx from 'clsx';
import Typography from '@mui/material/Typography';
import { ThemesType } from '@fuse/core/Settings/Settings';
import { Box } from '@mui/material';

export type ThemeOption = {
	id: string;
	section: ThemesType;
};

type ThemePreviewProps = {
	className?: string;
	onSelect?: (T: ThemeOption) => void;
	theme: ThemeOption;
};

/**
 * The ThemePreview component is responsible for rendering a preview of a theme scheme.
 * It uses various MUI components to render the preview.
 * The component is memoized to prevent unnecessary re-renders.
 */
function ThemePreview(props: ThemePreviewProps) {
	const { theme, className, onSelect = () => {} } = props;
	const { section, id } = theme;

	const { navbar, toolbar, footer, main } = section;

	return (
		<div className={clsx(className, 'min-h-full w-full')}>
			<button
				className={clsx(
					'relative flex h-40 w-full cursor-pointer items-stretch overflow-hidden rounded-sm p-0 text-left font-medium shadow-sm transition-all duration-200 ease-in-out hover:scale-105 hover:shadow-lg',
					{
						'bg-white': id === 'default',
						'bg-gray-700': id !== 'default'
					}
				)}
				style={{
					backgroundColor: main?.palette.background.default,
					color: main?.palette.text.primary
				}}
				onClick={() => {
					onSelect(theme);
				}}
				type="button"
			>
				<Box
					className="flex min-h-full w-1/3 flex-col p-1"
					sx={{
						borderRight: '1px solid',
						borderColor: navbar?.palette.divider,
						backgroundColor: navbar?.palette.background.default,
						color: navbar?.palette.text.primary
					}}
				>
					<span className="text-sm">Navbar</span>
				</Box>

				<div className="flex w-2/3 flex-col">
					<Box
						className="border-b-1 w-full border-gray-700 px-1 py-1"
						sx={{
							borderBottom: '1px solid',
							borderColor: toolbar.palette.divider,
							backgroundColor: toolbar.palette.background.default,
							color: toolbar.palette.text.primary
						}}
					>
						<span className="text-sm">Toolbar</span>
					</Box>

					<div className="flex w-full flex-1 flex-col">
						<div
							className="relative h-11 w-full px-1"
							style={{
								backgroundColor: main?.palette.primary.main,
								color: main?.palette.primary.contrastText
							}}
						>
							<span className="text-sm">Header</span>

							<Box
								className="shadow-1 absolute bottom-0 right-0 z-10 mb-2.5 mr-2 flex h-6 w-6 items-center justify-center rounded-full text-xs"
								sx={{
									backgroundColor: main?.palette.secondary.main,
									color: main?.palette.secondary.contrastText
								}}
							>
								<span className="">S</span>
							</Box>
						</div>

						<div className="-mt-6 w-full flex-1 pl-1 pr-1">
							<div
								className="shadow-1 relative h-full w-full rounded-sm p-1"
								style={{
									backgroundColor: main?.palette.background.paper,
									color: main?.palette.text.primary
								}}
							>
								<span className="text-sm">Paper</span>
							</div>
						</div>

						<div className="w-full p-1">
							<span className="text-sm">Background</span>
						</div>
					</div>

					<Box
						className="w-full px-2 py-1"
						sx={{
							borderTop: '1px solid',
							borderColor: footer?.palette.divider,
							backgroundColor: footer?.palette.background.default,
							color: footer?.palette.text.primary
						}}
					>
						<span className="text-sm">Footer</span>
					</Box>
				</div>
			</button>
			<Typography className="mt-3 w-full text-center font-semibold">{id}</Typography>
		</div>
	);
}

export default ThemePreview;
