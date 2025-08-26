import { Box } from '@mui/material';
import Typography from '@mui/material/Typography';
import clsx from 'clsx';
import Link from '@fuse/core/Link';
import SvgIcon from '@fuse/core/SvgIcon';

type GoToDocBoxProps = {
	className?: string;
};

function GoToDocBox(props: GoToDocBoxProps) {
	const { className } = props;
	return (
		<Box
			className={clsx('documentation-hero border-1 flex flex-col gap-2 rounded-sm px-3 py-2', className)}
			sx={{ backgroundColor: 'background.paper', borderColor: 'divider' }}
		>
			<Typography className="truncate">Need assistance to get started?</Typography>
			<Typography
				className="flex items-center gap-1 truncate"
				component={Link}
				to="/documentation"
				color="secondary"
			>
				View documentation <SvgIcon>lucide:arrow-right</SvgIcon>
			</Typography>
		</Box>
	);
}

export default GoToDocBox;
