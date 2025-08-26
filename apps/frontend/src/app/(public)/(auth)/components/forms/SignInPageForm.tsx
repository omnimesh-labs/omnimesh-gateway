'use client';
import { Controller, useForm } from 'react-hook-form';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Link from '@fuse/core/Link';
import _ from 'lodash';
import SvgIcon from '@fuse/core/SvgIcon';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import FormLabel from '@mui/material/FormLabel';

/**
 * Form Validation Schema
 */
const schema = z.object({
	email: z.string().email('You must enter a valid email').nonempty('You must enter an email'),
	password: z
		.string()
		.min(8, 'Password is too short - must be at least 8 chars.')
		.nonempty('Please enter your password.'),
	remember: z.boolean().optional()
});

type FormType = z.infer<typeof schema>;

const defaultValues = {
	email: '',
	password: '',
	remember: true
};

function SignInPageForm() {
	const { control, formState, handleSubmit, reset } = useForm<FormType>({
		mode: 'onChange',
		defaultValues,
		resolver: zodResolver(schema)
	});

	const { isValid, dirtyFields, errors } = formState;

	function onSubmit() {
		reset(defaultValues);
	}

	return (
		<form
			name="loginForm"
			noValidate
			className="flex w-full flex-col justify-center gap-4"
			onSubmit={handleSubmit(onSubmit)}
		>
			<Controller
				name="email"
				control={control}
				render={({ field }) => (
					<FormControl>
						<FormLabel htmlFor="email">Email address</FormLabel>
						<TextField
							{...field}
							autoFocus
							type="email"
							error={!!errors.email}
							helperText={errors?.email?.message}
							required
							fullWidth
						/>
					</FormControl>
				)}
			/>

			<Controller
				name="password"
				control={control}
				render={({ field }) => (
					<FormControl>
						<FormLabel htmlFor="password">Password</FormLabel>
						<TextField
							{...field}
							type="password"
							error={!!errors.password}
							helperText={errors?.password?.message}
							required
							fullWidth
						/>
					</FormControl>
				)}
			/>

			<div className="flex flex-col items-center justify-center sm:flex-row sm:justify-between">
				<Controller
					name="remember"
					control={control}
					render={({ field }) => (
						<FormControl>
							<FormControlLabel
								label="Remember me"
								control={
									<Checkbox
										size="small"
										{...field}
									/>
								}
							/>
						</FormControl>
					)}
				/>

				<Link
					className="text-md font-medium"
					to="/pages/auth/forgot-password"
				>
					Forgot password?
				</Link>
			</div>

			<Button
				variant="contained"
				color="secondary"
				className="w-full"
				aria-label="Sign in"
				disabled={_.isEmpty(dirtyFields) || !isValid}
				type="submit"
				size="medium"
			>
				Sign in
			</Button>

			<div className="flex items-center py-4">
				<div className="mt-px flex-auto border-t" />
				<Typography
					className="mx-2"
					color="text.secondary"
				>
					Or continue with
				</Typography>
				<div className="mt-px flex-auto border-t" />
			</div>

			<div className="flex items-center gap-4">
				<Button
					variant="outlined"
					className="flex-auto"
				>
					<SvgIcon
						size={20}
						color="action"
					>
						feather:facebook
					</SvgIcon>
				</Button>
				<Button
					variant="outlined"
					className="flex-auto"
				>
					<SvgIcon
						size={20}
						color="action"
					>
						feather:twitter
					</SvgIcon>
				</Button>
				<Button
					variant="outlined"
					className="flex-auto"
				>
					<SvgIcon
						size={20}
						color="action"
					>
						feather:github
					</SvgIcon>
				</Button>
			</div>
		</form>
	);
}

export default SignInPageForm;
