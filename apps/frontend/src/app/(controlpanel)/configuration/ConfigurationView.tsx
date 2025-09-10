'use client';

import { useState } from 'react';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Paper,
	Box,
	Tabs,
	Tab,
	Card,
	CardContent,
	CardActions,
	LinearProgress,
	Alert,
	Stack,
	FormControl,
	FormLabel,
	FormGroup,
	FormControlLabel,
	Checkbox,
	TextField,
	Chip
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { configApi } from '@/lib/config-api';

const Root = styled(PageSimple)(({ theme }) => ({
	'& .PageSimple-header': {
		backgroundColor: theme.vars.palette.background.paper,
		borderBottomWidth: 1,
		borderStyle: 'solid',
		borderColor: theme.vars.palette.divider
	},
	'& .PageSimple-content': {
		backgroundColor: theme.vars.palette.background.default
	}
}));

interface ExportOptions {
	entityTypes: string[];
	includeInactive: boolean;
	includeDependencies: boolean;
}

function ConfigurationView() {
	const [tabValue, setTabValue] = useState(0);
	const [isExporting, setIsExporting] = useState(false);
	const [isImporting, setIsImporting] = useState(false);
	const [exportProgress, setExportProgress] = useState(0);
	const [exportOptions, setExportOptions] = useState<ExportOptions>({
		entityTypes: ['servers', 'namespaces'],
		includeInactive: false,
		includeDependencies: true
	});
	const { enqueueSnackbar } = useSnackbar();

	const handleExport = async () => {
		setIsExporting(true);
		setExportProgress(0);

		try {
			// Call the actual export API
			const exportData = await configApi.exportConfiguration({
				entityTypes: exportOptions.entityTypes,
				includeInactive: exportOptions.includeInactive,
				includeDependencies: exportOptions.includeDependencies
			});

			// Update progress
			setExportProgress(100);

			// Create download blob
			const blob = new Blob([JSON.stringify(exportData, null, 2)], {
				type: 'application/json'
			});
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `omnimesh-gateway-config-${Date.now()}.json`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);

			enqueueSnackbar('Configuration exported successfully', { variant: 'success' });
		} catch (error) {
			enqueueSnackbar('Export failed: ' + (error instanceof Error ? error.message : 'Unknown error'), {
				variant: 'error'
			});
		} finally {
			setIsExporting(false);
			setExportProgress(0);
		}
	};

	const handleImport = async (file: File) => {
		setIsImporting(true);
		try {
			const text = await file.text();
			const importData = JSON.parse(text);
			await configApi.importConfiguration(importData);
			enqueueSnackbar('Configuration imported successfully', { variant: 'success' });
		} catch (error) {
			enqueueSnackbar('Import failed: ' + (error instanceof Error ? error.message : 'Invalid file format'), {
				variant: 'error'
			});
		} finally {
			setIsImporting(false);
		}
	};

	const handleEntityTypeChange = (entityType: string, checked: boolean) => {
		setExportOptions((prev) => ({
			...prev,
			entityTypes: checked
				? [...prev.entityTypes, entityType]
				: prev.entityTypes.filter((type) => type !== entityType)
		}));
	};

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Configuration</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Export and import your Omnimesh AI Gateway configuration
							</Typography>
						</div>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<Box className="mb-6">
						<Tabs
							value={tabValue}
							onChange={(_, newValue) => setTabValue(newValue)}
							variant="scrollable"
							scrollButtons="auto"
							sx={{
								'& .MuiTab-root': {
									gap: '8px',
									paddingLeft: '16px',
									paddingRight: '16px'
								}
							}}
						>
							<Tab
								label="Export Configuration"
								icon={<SvgIcon size={20}>lucide:download</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label="Import Configuration"
								icon={<SvgIcon size={20}>lucide:upload</SvgIcon>}
								iconPosition="start"
							/>
						</Tabs>
					</Box>

					{tabValue === 0 && (
						<Stack spacing={3}>
							<Alert severity="info">
								Export your Omnimesh AI Gateway configuration including servers, namespaces, and content.
							</Alert>

							<Card>
								<CardContent>
									<Typography
										variant="h6"
										gutterBottom
									>
										Export Options
									</Typography>

									<Stack spacing={3}>
										<FormControl component="fieldset">
											<FormLabel component="legend">Entity Types</FormLabel>
											<FormGroup>
												{[
													{ id: 'servers', label: 'MCP Servers' },
													{ id: 'namespaces', label: 'Namespaces' },
													{ id: 'tools', label: 'Tools' },
													{ id: 'prompts', label: 'Prompts' },
													{ id: 'resources', label: 'Resources' }
												].map((entityType) => (
													<FormControlLabel
														key={entityType.id}
														control={
															<Checkbox
																checked={exportOptions.entityTypes.includes(
																	entityType.id
																)}
																onChange={(e) =>
																	handleEntityTypeChange(
																		entityType.id,
																		e.target.checked
																	)
																}
															/>
														}
														label={entityType.label}
													/>
												))}
											</FormGroup>
										</FormControl>

										<FormControl component="fieldset">
											<FormLabel component="legend">Additional Options</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={
														<Checkbox
															checked={exportOptions.includeInactive}
															onChange={(e) =>
																setExportOptions((prev) => ({
																	...prev,
																	includeInactive: e.target.checked
																}))
															}
														/>
													}
													label="Include inactive items"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={exportOptions.includeDependencies}
															onChange={(e) =>
																setExportOptions((prev) => ({
																	...prev,
																	includeDependencies: e.target.checked
																}))
															}
														/>
													}
													label="Include dependencies"
												/>
											</FormGroup>
										</FormControl>
									</Stack>

									{isExporting && (
										<Box sx={{ mt: 2 }}>
											<Typography
												variant="body2"
												color="textSecondary"
												gutterBottom
											>
												Exporting configuration... {exportProgress}%
											</Typography>
											<LinearProgress
												variant="determinate"
												value={exportProgress}
											/>
										</Box>
									)}
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={<SvgIcon>lucide:download</SvgIcon>}
										onClick={handleExport}
										disabled={isExporting || exportOptions.entityTypes.length === 0}
									>
										{isExporting ? 'Exporting...' : 'Export Configuration'}
									</Button>
								</CardActions>
							</Card>

							<Paper sx={{ p: 2 }}>
								<Typography
									variant="subtitle2"
									gutterBottom
								>
									Selected for Export:
								</Typography>
								<Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
									{exportOptions.entityTypes.map((type) => (
										<Chip
											key={type}
											label={type}
											size="small"
										/>
									))}
									{exportOptions.entityTypes.length === 0 && (
										<Typography
											variant="body2"
											color="textSecondary"
										>
											No entity types selected
										</Typography>
									)}
								</Box>
							</Paper>
						</Stack>
					)}

					{tabValue === 1 && (
						<Stack spacing={3}>
							<Alert severity="warning">
								Import functionality will overwrite existing configuration. Please backup your current
								setup first.
							</Alert>

							<Card>
								<CardContent>
									<Typography
										variant="h6"
										gutterBottom
									>
										Import Configuration
									</Typography>

									<Stack spacing={2}>
										<TextField
											type="file"
											label="Configuration File"
											slotProps={{
												inputLabel: { shrink: true },
												htmlInput: { accept: '.json' }
											}}
											helperText="Select a JSON configuration file to import"
										/>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={<SvgIcon>lucide:upload</SvgIcon>}
										onClick={handleImport}
										disabled={isImporting}
									>
										{isImporting ? 'Importing...' : 'Import Configuration'}
									</Button>
								</CardActions>
							</Card>

							<Alert severity="info">
								<Typography
									variant="subtitle2"
									gutterBottom
								>
									Supported Configuration Format:
								</Typography>
								<Typography
									variant="body2"
									component="div"
								>
									<ul>
										<li>JSON format with metadata section</li>
										<li>Entity sections: servers, namespaces, tools, prompts, resources</li>
										<li>Version compatibility checks</li>
										<li>Conflict resolution options</li>
									</ul>
								</Typography>
							</Alert>
						</Stack>
					)}
				</div>
			}
		/>
	);
}

export default ConfigurationView;
