import Button, { ButtonProps } from '@mui/material/Button';
import clsx from 'clsx';
import SvgIcon from '@fuse/core/SvgIcon';

type PurchaseButtonProps = ButtonProps & {
	className?: string;
};

/**
 * The purchase button.
 */
function PurchaseButton(props: PurchaseButtonProps) {
	const { className = '', children = 'Purchase FUSE React', ...rest } = props;

	return (
		<Button
			component="a"
			href="https://1.envato.market/zDGL6"
			target="_blank"
			rel="noreferrer noopener"
			role="button"
			className={clsx('whitespace-nowrap', className)}
			variant="contained"
			color="secondary"
			startIcon={<SvgIcon>lucide:shopping-cart</SvgIcon>}
			{...rest}
		>
			{children}
		</Button>
	);
}

export default PurchaseButton;
