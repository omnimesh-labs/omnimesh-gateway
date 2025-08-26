import * as React from 'react';
import { Dialog as MuiDialog, DialogActions, IconButton, Typography, Box } from '@mui/material';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DialogContextValue {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

const DialogContext = React.createContext<DialogContextValue | undefined>(undefined);

interface DialogProps {
	open?: boolean;
	defaultOpen?: boolean;
	onOpenChange?: (open: boolean) => void;
	children?: React.ReactNode;
}

const Dialog = ({ open: controlledOpen, defaultOpen = false, onOpenChange, children }: DialogProps) => {
	const [uncontrolledOpen, setUncontrolledOpen] = React.useState(defaultOpen);
	const open = controlledOpen !== undefined ? controlledOpen : uncontrolledOpen;

	const handleOpenChange = React.useCallback(
		(newOpen: boolean) => {
			if (controlledOpen === undefined) {
				setUncontrolledOpen(newOpen);
			}

			onOpenChange?.(newOpen);
		},
		[controlledOpen, onOpenChange]
	);

	return <DialogContext.Provider value={{ open, onOpenChange: handleOpenChange }}>{children}</DialogContext.Provider>;
};

const DialogTrigger = React.forwardRef<HTMLButtonElement, React.ButtonHTMLAttributes<HTMLButtonElement>>(
	({ children, onClick, ...props }, ref) => {
		const context = React.useContext(DialogContext);

		return (
			<button
				ref={ref}
				onClick={(e) => {
					onClick?.(e);
					context?.onOpenChange(true);
				}}
				{...props}
			>
				{children}
			</button>
		);
	}
);
DialogTrigger.displayName = 'DialogTrigger';

const DialogPortal = ({ children }: { children?: React.ReactNode }) => {
	return <>{children}</>;
};

const DialogClose = React.forwardRef<HTMLButtonElement, React.ButtonHTMLAttributes<HTMLButtonElement>>(
	({ children, onClick, ...props }, ref) => {
		const context = React.useContext(DialogContext);

		return (
			<button
				ref={ref}
				onClick={(e) => {
					onClick?.(e);
					context?.onOpenChange(false);
				}}
				{...props}
			>
				{children}
			</button>
		);
	}
);
DialogClose.displayName = 'DialogClose';

const DialogOverlay = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
	({ className, ...props }, ref) => {
		const context = React.useContext(DialogContext);

		if (!context?.open) return null;

		return (
			<div
				ref={ref}
				className={cn('fixed inset-0 z-50 bg-black/80', className)}
				{...props}
			/>
		);
	}
);
DialogOverlay.displayName = 'DialogOverlay';

const DialogContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
	({ className, children, ...props }, ref) => {
		const context = React.useContext(DialogContext);

		if (!context) {
			throw new Error('DialogContent must be used within a Dialog');
		}

		return (
			<MuiDialog
				open={context.open}
				onClose={() => context.onOpenChange(false)}
				maxWidth="md"
				fullWidth
			>
				<Box
					ref={ref}
					className={cn('relative', className)}
					{...props}
				>
					{children}
					<IconButton
						aria-label="close"
						onClick={() => context.onOpenChange(false)}
						sx={{
							position: 'absolute',
							right: 8,
							top: 8
						}}
						size="small"
					>
						<X size={18} />
					</IconButton>
				</Box>
			</MuiDialog>
		);
	}
);
DialogContent.displayName = 'DialogContent';

const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
	<div
		className={cn('flex flex-col space-y-1.5 p-6 pb-0 text-center sm:text-left', className)}
		{...props}
	/>
);
DialogHeader.displayName = 'DialogHeader';

const DialogFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
	<DialogActions
		className={cn('p-6 pt-0', className)}
		{...props}
	/>
);
DialogFooter.displayName = 'DialogFooter';

const DialogTitle = React.forwardRef<HTMLHeadingElement, React.HTMLAttributes<HTMLHeadingElement>>(
	({ className, children, ...props }, ref) => (
		<Typography
			ref={ref}
			component="h2"
			variant="h6"
			className={cn('font-semibold leading-none tracking-tight', className)}
			{...props}
		>
			{children}
		</Typography>
	)
);
DialogTitle.displayName = 'DialogTitle';

const DialogDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
	({ className, children, ...props }, ref) => (
		<Typography
			ref={ref}
			variant="body2"
			color="text.secondary"
			className={cn('mt-1', className)}
			{...props}
		>
			{children}
		</Typography>
	)
);
DialogDescription.displayName = 'DialogDescription';

export {
	Dialog,
	DialogPortal,
	DialogOverlay,
	DialogTrigger,
	DialogClose,
	DialogContent,
	DialogHeader,
	DialogFooter,
	DialogTitle,
	DialogDescription
};
