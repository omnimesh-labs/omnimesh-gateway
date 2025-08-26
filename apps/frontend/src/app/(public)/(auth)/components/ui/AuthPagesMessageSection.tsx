import AvatarGroup from '@mui/material/AvatarGroup';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';

function AuthPagesMessageSection() {
	return (
		<Box
			className="relative hidden h-full flex-auto items-center justify-center overflow-hidden p-16 md:flex lg:px-28"
			sx={{
				backgroundColor: 'primary.dark',
				color: 'primary.contrastText'
			}}
		>
			<svg
				className="pointer-events-none absolute inset-0"
				viewBox="0 0 960 540"
				width="100%"
				height="100%"
				preserveAspectRatio="xMidYMax slice"
				xmlns="http://www.w3.org/2000/svg"
			>
				<Box
					component="g"
					className="opacity-5"
					fill="none"
					stroke="currentColor"
					strokeWidth="100"
				>
					<circle
						r="234"
						cx="196"
						cy="23"
					/>
					<circle
						r="234"
						cx="790"
						cy="491"
					/>
				</Box>
			</svg>
			<Box
				component="svg"
				className="absolute -top-16 -right-16 opacity-20"
				sx={{ color: 'primary.light' }}
				viewBox="0 0 220 192"
				width="220px"
				height="192px"
				fill="none"
			>
				<defs>
					<pattern
						id="837c3e70-6c3a-44e6-8854-cc48c737b659"
						x="0"
						y="0"
						width="20"
						height="20"
						patternUnits="userSpaceOnUse"
					>
						<rect
							x="0"
							y="0"
							width="4"
							height="4"
							fill="currentColor"
						/>
					</pattern>
				</defs>
				<rect
					width="220"
					height="192"
					fill="url(#837c3e70-6c3a-44e6-8854-cc48c737b659)"
				/>
			</Box>

			<div className="relative z-10 w-full max-w-4xl">
				<div className="text-7xl leading-none font-bold text-gray-100">
					<div>Welcome to</div>
					<div>our community</div>
				</div>
				<div className="mt-6 text-lg leading-6 tracking-tight text-gray-400">
					MCP Gateway provides enterprise-grade infrastructure for Model Context Protocol servers with 
					authentication, logging, rate limiting, and multi-protocol transport support.
				</div>
				<div className="mt-8 flex items-center">
					<AvatarGroup
						sx={{
							'& .MuiAvatar-root': {
								borderColor: 'primary.main'
							}
						}}
					>
						<Avatar src="/assets/images/avatars/female-18.jpg" />
						<Avatar src="/assets/images/avatars/female-11.jpg" />
						<Avatar src="/assets/images/avatars/male-09.jpg" />
						<Avatar src="/assets/images/avatars/male-16.jpg" />
					</AvatarGroup>

					<div className="ml-4 font-medium tracking-tight text-gray-400">
						Join developers building the future of AI integration
					</div>
				</div>
			</div>
		</Box>
	);
}

export default AuthPagesMessageSection;
