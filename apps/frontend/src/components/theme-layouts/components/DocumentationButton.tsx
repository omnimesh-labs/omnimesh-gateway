import Button from '@mui/material/Button';
import Link from '@fuse/core/Link';
import SvgIcon from '@fuse/core/SvgIcon';

type DocumentationButtonProps = {
	className?: string;
};

/**
 * The documentation button.
 */
function DocumentationButton(props: DocumentationButtonProps) {
	const { className = '' } = props;

	return (
		<Button
			component={Link}
			to="/documentation"
			role="button"
			className={className}
			variant="contained"
			color="primary"
			startIcon={<SvgIcon>lucide:book-open</SvgIcon>}
		>
			Documentation
		</Button>
	);
}

export default DocumentationButton;
