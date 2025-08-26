import { Box, Button, lighten, Typography } from '@mui/material';
import { signIn } from 'next-auth/react';
import { authJsProviderMap } from '@auth/authJs';

const providerLogoPath = 'https://authjs.dev/img/providers';

function AuthJsProviderSelect() {
	function handleSignIn(providerId: string) {
		try {
			signIn(providerId);
		} catch (error) {
			console.error(error);
		}
	}

	if (authJsProviderMap?.length === 0) {
		return null;
	}

	return (
		<div className="w-full">
			<div className="mb-8 flex items-center">
				<div className="mt-px flex-auto border-t" />
				<Typography
					className="mx-2"
					color="text.secondary"
				>
					Or continue with
				</Typography>
				<div className="mt-px flex-auto border-t" />
			</div>
			<div className="flex flex-col gap-3">
				{Object.values(authJsProviderMap)
					.filter((provider) => provider.id !== 'credentials')
					.map((provider) => (
						<Button
							key={provider.id}
							className="items-between text-md flex"
							onClick={() => handleSignIn(provider.id)}
							sx={(theme) => ({
								backgroundColor: theme.palette.background.default,
								color: theme.palette.text.primary,
								'&:hover': {
									color: provider?.style?.text || theme.palette.secondary.contrastText,
									backgroundColor: provider?.style?.bg || theme.palette.secondary.main,
									'& .provider-icon': {
										backgroundColor: provider?.style?.bg
											? lighten(provider?.style?.bg, 0.7)
											: theme.palette.secondary.light
									}
								}
							})}
							endIcon={
								<Box className="provider-icon flex h-6 w-6 items-center justify-center rounded-full">
									<img
										className="flex h-auto w-4"
										src={`${providerLogoPath}/${provider.id}.svg`}
										alt={provider.name}
									/>
								</Box>
							}
						>
							<span className="flex flex-1">Sign in with {provider.name}</span>
						</Button>
					))}
				<Button
					className="text-md"
					href="https://authjs.dev/getting-started#official-providers"
					target="_blank"
				>
					+ more auth providers
				</Button>
			</div>
		</div>
	);
}

export default AuthJsProviderSelect;
