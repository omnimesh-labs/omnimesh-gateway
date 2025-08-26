import React from 'react';
import { MenuItem, Select, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { useTheme } from '@mui/material/styles';

type ProjectOption = {
	value: string;
	logo: string;
	darkLogo: string;
	name: string;
	url: string;
};

const projectOptions: ProjectOption[] = [
	{
		value: 'Janex',
		logo: '/assets/images/logo/mcp-gateway.svg',
		darkLogo: '/assets/images/logo/mcp-gateway-dark.svg',
		name: 'Janex',
		url: typeof window !== 'undefined' ? window.location.origin : ''
	}
];

function MainProjectSelection() {
	const [selectedProjectValue, setSelectedProject] = React.useState<string>(projectOptions[0].value);
	const selectedProject = projectOptions.find((project) => project.value === selectedProjectValue);
	const theme = useTheme();
	const handleMenuItemClick = (projectValue: string) => {
		setSelectedProject(projectValue);

		const selectedProjectUrl = projectOptions.find((project) => project.value === projectValue)?.url;

		if (typeof window !== 'undefined' && selectedProjectUrl) {
			const currentUrl = new URL(window.location.href);
			const newUrl = selectedProjectUrl + currentUrl.pathname;

			window.location.href = newUrl;
		}
	};

	return (
		<Select
			value={selectedProjectValue}
			onChange={(event) => handleMenuItemClick(event.target.value)}
			displayEmpty
			renderValue={(_selectedValue) => (
				<div style={{ display: 'flex', alignItems: 'center' }}>
					<img
						src={theme.palette.mode === 'dark' ? selectedProject.darkLogo : selectedProject.logo}
						alt={`${selectedProject.name} Logo`}
						width={16}
						height={16}
						style={{ marginRight: 8 }}
					/>
					<Typography className="text-md font-semibold">{selectedProject.name}</Typography>
				</div>
			)}
			sx={{
				backgroundColor: 'transparent',
				'& .MuiInputBase-input': {
					padding: '0 22px 0 8px!important'
				},
				'& .MuiSelect-icon': {
					width: 20,
					right: 1
				}
			}}
			size="small"
		>
			{projectOptions.map((project) => (
				<MenuItem
					key={project.value}
					value={project.value}
				>
					<ListItemIcon>
						<img
							src={theme.palette.mode === 'dark' ? project.darkLogo : project.logo}
							alt={`${project.name} Logo`}
							width={16}
							height={16}
						/>
					</ListItemIcon>
					<ListItemText
						primary={project.name}
						classes={{ primary: 'text-md font-semibold' }}
					/>
				</MenuItem>
			))}
		</Select>
	);
}

export default MainProjectSelection;
