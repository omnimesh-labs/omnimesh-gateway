import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import { lighten } from '@mui/material/styles';
import MCPGatewaySignInForm from '@auth/forms/MCPGatewaySignInForm';
import SignInPageTitle from '../ui/SignInPageTitle';
import AuthPagesMessageSection from '../ui/AuthPagesMessageSection';

/**
 * The sign in page.
 */
function SignInPageView() {
	return (
		<div className="flex min-w-0 flex-auto flex-col items-center sm:flex-row sm:justify-center md:items-start md:justify-start">
			<Paper className="ltr:border-r-1 rtl:border-l-1 h-full w-full px-4 py-2 sm:h-auto sm:w-auto sm:rounded-xl sm:p-12 sm:shadow-sm md:flex md:h-full md:w-1/2 md:items-center md:justify-end md:rounded-none md:p-16 md:shadow-none">
				<div className="mx-auto flex w-full max-w-80 flex-col gap-8 sm:mx-0 sm:w-80">
					<SignInPageTitle />

					<MCPGatewaySignInForm />

					<Box
						className="rounded-lg px-4 py-2 text-md leading-[1.625]"
						sx={{
							backgroundColor: (theme) => lighten(theme.palette.primary.main, 0.8),
							color: 'primary.dark'
						}}
					>
						Welcome to <b>Omnimesh Gateway</b>. Sign in to access your dashboard and manage your MCP servers.
					</Box>
				</div>
			</Paper>

			<AuthPagesMessageSection />
		</div>
	);
}

export default SignInPageView;
