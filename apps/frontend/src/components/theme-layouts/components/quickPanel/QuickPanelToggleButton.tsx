import IconButton from '@mui/material/IconButton';
import SvgIcon from '@fuse/core/SvgIcon';
import { useQuickPanelContext } from './contexts/QuickPanelContext/useQuickPanelContext';

type QuickPanelToggleButtonProps = {
	className?: string;
	children?: React.ReactNode;
};

/**
 * The quick panel toggle button.
 */
function QuickPanelToggleButton(props: QuickPanelToggleButtonProps) {
	const { className = '', children = <SvgIcon>lucide:bookmark</SvgIcon> } = props;
	const { toggleQuickPanel } = useQuickPanelContext();

	return (
		<IconButton
			onClick={() => toggleQuickPanel()}
			className={className}
		>
			{children}
		</IconButton>
	);
}

export default QuickPanelToggleButton;
