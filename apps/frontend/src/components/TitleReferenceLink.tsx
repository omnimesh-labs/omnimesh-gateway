'use client';

import { Tooltip, Typography } from '@mui/material';
import Link from '@fuse/core/Link';
import usePathname from '@fuse/hooks/usePathname';
import { useState } from 'react';
import Box from '@mui/material/Box';

type TitleReferenceLinkProps = {
	id: string;
	children?: React.ReactNode;
};

function TitleReferenceLink(props: TitleReferenceLinkProps) {
	const { children = '#', id = '' } = props;
	const pathname = usePathname();
	const href = `${typeof window !== 'undefined' ? window.location.origin : ''}${pathname}#${id}`;
	const [open, setOpen] = useState(false);

	function handleCopy() {
		navigator.clipboard.writeText(href);
		setOpen(true);
		setTimeout(() => {
			setOpen(false);
		}, 800);
	}

	return (
		<Tooltip
			title="Copied!"
			open={open}
			slotProps={{ popper: { placement: 'top' } }}
			arrow
		>
			<Box
				component="span"
				id={id}
			>
				<Typography
					component={Link}
					to={href}
					className="italic opacity-30"
					sx={{ color: 'inherit!important', textDecoration: 'none!important' }}
					onClick={handleCopy}
				>
					{children}
				</Typography>
			</Box>
		</Tooltip>
	);
}

export default TitleReferenceLink;
