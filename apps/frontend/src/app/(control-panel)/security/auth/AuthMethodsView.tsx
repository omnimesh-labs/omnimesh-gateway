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
	Alert,
	Stack,
	FormControl,
	FormLabel,
	FormGroup,
	FormControlLabel,
	Checkbox,
	TextField,
	Chip,
	Switch,
	Select,
	MenuItem,
	InputLabel
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';

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

interface AuthConfig {
	methods: string[];
	jwtEnabled: boolean;
	apiKeysEnabled: boolean;
	oauth2Enabled: boolean;
	sessionTimeout: number;
	mfaRequired: boolean;
}

function AuthMethodsView() {
	const [tabValue, setTabValue] = useState(0);
	const [isSaving, setIsSaving] = useState(false);
	const [authConfig, setAuthConfig] = useState<AuthConfig>({
		methods: ['jwt', 'api_key'],
		jwtEnabled: true,
		apiKeysEnabled: true,
		oauth2Enabled: false,
		sessionTimeout: 3600,
		mfaRequired: false
	});
	const { enqueueSnackbar } = useSnackbar();

	const handleSaveAuthConfig = async () => {
		setIsSaving(true);
		try {
			// Simulate save operation
			await new Promise((resolve) => setTimeout(resolve, 1500));
			enqueueSnackbar('Authentication configuration saved successfully', { variant: 'success' });
		} catch (error) {
			enqueueSnackbar(
				'Failed to save configuration: ' + (error instanceof Error ? error.message : 'Unknown error'),
				{
					variant: 'error'
				}
			);
		} finally {
			setIsSaving(false);
		}
	};

	const _handleMethodToggle = (method: string, checked: boolean) => {
		setAuthConfig((prev) => ({
			...prev,
			methods: checked ? [...prev.methods, method] : prev.methods.filter((m) => m !== method)
		}));
	};

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Auth Methods</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Configure authentication methods and security settings
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
						>
							<Tab
								label="Authentication Methods"
								icon={<SvgIcon size={20}>lucide:key</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label="Session Management"
								icon={<SvgIcon size={20}>lucide:clock</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label="Security Settings"
								icon={<SvgIcon size={20}>lucide:shield</SvgIcon>}
								iconPosition="start"
							/>
						</Tabs>
					</Box>

					{tabValue === 0 && (
						<Stack spacing={3}>
							<Alert severity="info">
								Configure which authentication methods are available for your MCP Gateway instance.
							</Alert>

							<Card>
								<CardContent>
									<Typography
										variant="h6"
										gutterBottom
									>
										Available Methods
									</Typography>

									<Stack spacing={3}>
										<FormControl component="fieldset">
											<FormLabel component="legend">Primary Authentication</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.jwtEnabled}
															onChange={(e) =>
																setAuthConfig((prev) => ({
																	...prev,
																	jwtEnabled: e.target.checked
																}))
															}
														/>
													}
													label="JWT Token Authentication"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.apiKeysEnabled}
															onChange={(e) =>
																setAuthConfig((prev) => ({
																	...prev,
																	apiKeysEnabled: e.target.checked
																}))
															}
														/>
													}
													label="API Key Authentication"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.oauth2Enabled}
															onChange={(e) =>
																setAuthConfig((prev) => ({
																	...prev,
																	oauth2Enabled: e.target.checked
																}))
															}
														/>
													}
													label="OAuth 2.0 / OpenID Connect"
												/>
											</FormGroup>
										</FormControl>

										<FormControl component="fieldset">
											<FormLabel component="legend">Multi-Factor Authentication</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={
														<Switch
															checked={authConfig.mfaRequired}
															onChange={(e) =>
																setAuthConfig((prev) => ({
																	...prev,
																	mfaRequired: e.target.checked
																}))
															}
														/>
													}
													label="Require MFA for all users"
												/>
											</FormGroup>
										</FormControl>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={<SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={
											isSaving ||
											(!authConfig.jwtEnabled &&
												!authConfig.apiKeysEnabled &&
												!authConfig.oauth2Enabled)
										}
									>
										{isSaving ? 'Saving...' : 'Save Configuration'}
									</Button>
								</CardActions>
							</Card>

							<Paper sx={{ p: 2 }}>
								<Typography
									variant="subtitle2"
									gutterBottom
								>
									Enabled Methods:
								</Typography>
								<Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
									{authConfig.jwtEnabled && (
										<Chip
											label="JWT"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.apiKeysEnabled && (
										<Chip
											label="API Keys"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.oauth2Enabled && (
										<Chip
											label="OAuth 2.0"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.mfaRequired && (
										<Chip
											label="MFA Required"
											color="warning"
											size="small"
										/>
									)}
									{!authConfig.jwtEnabled &&
										!authConfig.apiKeysEnabled &&
										!authConfig.oauth2Enabled && (
											<Typography
												variant="body2"
												color="textSecondary"
											>
												No authentication methods enabled
											</Typography>
										)}
								</Box>
							</Paper>
						</Stack>
					)}

					{tabValue === 1 && (
						<Stack spacing={3}>
							<Alert severity="info">Configure session timeout and token expiration settings.</Alert>

							<Card>
								<CardContent>
									<Typography
										variant="h6"
										gutterBottom
									>
										Session Configuration
									</Typography>

									<Stack spacing={3}>
										<TextField
											label="Session Timeout (seconds)"
											type="number"
											value={authConfig.sessionTimeout}
											onChange={(e) =>
												setAuthConfig((prev) => ({
													...prev,
													sessionTimeout: parseInt(e.target.value)
												}))
											}
											fullWidth
											helperText="How long before idle sessions expire"
										/>

										<FormControl fullWidth>
											<InputLabel>Token Refresh Strategy</InputLabel>
											<Select
												value="sliding"
												label="Token Refresh Strategy"
											>
												<MenuItem value="sliding">Sliding Window</MenuItem>
												<MenuItem value="fixed">Fixed Expiration</MenuItem>
												<MenuItem value="none">No Refresh</MenuItem>
											</Select>
										</FormControl>

										<TextField
											label="Max Concurrent Sessions"
											type="number"
											defaultValue={5}
											fullWidth
											helperText="Maximum number of concurrent sessions per user"
										/>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={<SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={isSaving}
									>
										{isSaving ? 'Saving...' : 'Save Configuration'}
									</Button>
								</CardActions>
							</Card>
						</Stack>
					)}

					{tabValue === 2 && (
						<Stack spacing={3}>
							<Alert severity="warning">
								Security settings affect all authentication methods. Change with caution.
							</Alert>

							<Card>
								<CardContent>
									<Typography
										variant="h6"
										gutterBottom
									>
										Security Policies
									</Typography>

									<Stack spacing={3}>
										<FormControl component="fieldset">
											<FormLabel component="legend">Password Requirements</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={<Checkbox defaultChecked />}
													label="Minimum 8 characters"
												/>
												<FormControlLabel
													control={<Checkbox defaultChecked />}
													label="Require uppercase letters"
												/>
												<FormControlLabel
													control={<Checkbox defaultChecked />}
													label="Require lowercase letters"
												/>
												<FormControlLabel
													control={<Checkbox defaultChecked />}
													label="Require numbers"
												/>
												<FormControlLabel
													control={<Checkbox />}
													label="Require special characters"
												/>
											</FormGroup>
										</FormControl>

										<FormControl component="fieldset">
											<FormLabel component="legend">Account Security</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={<Switch defaultChecked />}
													label="Lock account after 5 failed attempts"
												/>
												<FormControlLabel
													control={<Switch />}
													label="Require email verification"
												/>
												<FormControlLabel
													control={<Switch />}
													label="Force password reset after 90 days"
												/>
											</FormGroup>
										</FormControl>

										<TextField
											label="IP Whitelist (comma-separated)"
											fullWidth
											multiline
											rows={3}
											placeholder="192.168.1.0/24, 10.0.0.0/8"
											helperText="Leave empty to allow all IPs"
										/>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={<SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={isSaving}
									>
										{isSaving ? 'Saving...' : 'Apply Security Settings'}
									</Button>
								</CardActions>
							</Card>
						</Stack>
					)}
				</div>
			}
		/>
	);
}

export default AuthMethodsView;
