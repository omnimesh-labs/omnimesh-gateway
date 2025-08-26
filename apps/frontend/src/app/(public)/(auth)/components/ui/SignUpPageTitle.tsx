import Typography from '@mui/material/Typography';
import Link from '@fuse/core/Link';

function SignUpPageTitle() {
	return (
		<div className="w-full">
			<img
				className="w-12"
				src="/assets/images/logo/logo.svg"
				alt="logo"
			/>

			<Typography className="mt-8 text-4xl leading-[1.25] font-extrabold tracking-tight">Sign up</Typography>
			<div className="mt-0.5 flex items-baseline font-medium">
				<Typography>Already have an account?</Typography>
				<Link
					className="ml-1"
					to="/sign-in"
				>
					Sign in
				</Link>
			</div>
		</div>
	);
}

export default SignUpPageTitle;
