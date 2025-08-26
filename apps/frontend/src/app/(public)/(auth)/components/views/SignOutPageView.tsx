import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Link from '@fuse/core/Link';
import AuthPagesMessageSection from '../ui/AuthPagesMessageSection';
import SignOutPageTitle from '../ui/SignOutPageTitle';

/**
 * The sign out page.
 */
function SignOutPageView() {
	return (
		<div className="flex min-w-0 flex-auto flex-col items-center sm:flex-row sm:justify-center md:items-start md:justify-start">
			<Paper className="flex h-full w-full items-center px-4 py-2 sm:h-auto sm:w-auto sm:rounded-xl sm:p-12 sm:shadow-sm md:flex md:h-full md:w-1/2 md:items-center md:justify-end md:rounded-none md:p-16 md:shadow-none ltr:border-r-1 rtl:border-l-1">
				<div className="mx-auto w-full max-w-80 sm:mx-0 sm:w-80">
					<SignOutPageTitle />

					<Typography
						className="text-md mt-4 text-center font-medium"
						color="text.secondary"
					>
						Go to <Link to="/sign-in">sign in</Link>
					</Typography>
				</div>
			</Paper>

			<AuthPagesMessageSection />
		</div>
	);
}

export default SignOutPageView;
