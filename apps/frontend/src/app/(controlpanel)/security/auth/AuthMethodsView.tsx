'use client';

import { useState, useEffect } from 'react';
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
	InputLabel,
	CircularProgress,
	Skeleton
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { authConfigApi } from '@/lib/client-api';
import type { AuthConfiguration, AuthConfigurationRequest } from '@/lib/types';

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

function AuthMethodsView() {
	const [tabValue, setTabValue] = useState(0);
	const [isSaving, setIsSaving] = useState(false);
	const [isLoading, setIsLoading] = useState(true);
	const [authConfig, setAuthConfig] = useState<AuthConfiguration | null>(null);
	const { enqueueSnackbar } = useSnackbar();

	// Load auth configuration on component mount
	useEffect(() => {
		const loadAuthConfig = async () => {
			try {
				setIsLoading(true);
				const config = await authConfigApi.getAuthConfig();
				setAuthConfig(config);
			} catch (error) {
				enqueueSnackbar(
					'Failed to load authentication configuration: ' + (error instanceof Error ? error.message : 'Unknown error'),
					{ variant: 'error' }
				);
			} finally {
				setIsLoading(false);
			}
		};

		loadAuthConfig();
	}, [enqueueSnackbar]);

	const handleSaveAuthConfig = async () => {
		if (!authConfig) return;

		setIsSaving(true);
		try {
			// Transform the nested security structure to the flat structure expected by backend
			const flattenedSecurity = {
				// Password Requirements
				password_min_length: authConfig.security.password_requirements.min_length,
				password_require_uppercase: authConfig.security.password_requirements.require_uppercase,
				password_require_lowercase: authConfig.security.password_requirements.require_lowercase,
				password_require_numbers: authConfig.security.password_requirements.require_numbers,
				password_require_special: authConfig.security.password_requirements.require_special,
				password_max_age_days: authConfig.security.password_requirements.max_age_days,
				password_history_count: authConfig.security.password_requirements.history_count,

				// Account Security
				account_lockout_enabled: authConfig.security.account_security.lockout_enabled,
				account_lockout_threshold: authConfig.security.account_security.lockout_threshold,
				account_lockout_duration_minutes: authConfig.security.account_security.lockout_duration_minutes,

				// Email Verification
				email_verification_required: authConfig.security.email_verification.required,
				email_verification_expiry_hours: authConfig.security.email_verification.expiry_hours,

				// IP Restrictions
				ip_whitelist: authConfig.security.ip_restrictions.whitelist,
				geo_blocking_enabled: authConfig.security.ip_restrictions.geo_blocking_enabled,
				allowed_countries: authConfig.security.ip_restrictions.allowed_countries,

				// Compliance
				password_change_required: authConfig.security.password_change_required,
				compliance_mode: authConfig.security.compliance_mode
			};

			const request: AuthConfigurationRequest = {
				methods: authConfig.methods,
				session: authConfig.session,
				security: flattenedSecurity as any // Cast to bypass type checking since we're transforming the structure
			};

			const updatedConfig = await authConfigApi.updateAuthConfig(request);

			// Only update the state if we got a valid response
			if (updatedConfig) {
				setAuthConfig(updatedConfig);
			}
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

	const handleAuthMethodChange = (field: keyof AuthConfiguration['methods'], value: boolean | number | string) => {
		if (!authConfig) return;

		setAuthConfig(prev => prev ? {
			...prev,
			methods: {
				...prev.methods,
				[field]: value
			}
		} : prev);
	};

	const handleSessionChange = (field: keyof AuthConfiguration['session'], value: boolean | number | string) => {
		if (!authConfig) return;

		setAuthConfig(prev => prev ? {
			...prev,
			session: {
				...prev.session,
				[field]: value
			}
		} : prev);
	};

	const handleSecurityChange = (section: keyof AuthConfiguration['security'], field: string, value: boolean | number | string) => {
		if (!authConfig) return;

		setAuthConfig(prev => prev ? {
			...prev,
			security: {
				...prev.security,
				[section]: {
					...prev.security[section],
					[field]: value
				}
			}
		} : prev);
	};

	// Show loading skeleton while data is being fetched
	if (isLoading || !authConfig) {
		return (
			<Root
				header={
					<div className="p-6">
						<Skeleton variant="text" width={200} height={32} />
						<Skeleton variant="text" width={400} height={24} />
					</div>
				}
				content={
					<div className="p-6">
						<Stack spacing={3}>
							<Skeleton variant="rectangular" height={60} />
							<Skeleton variant="rectangular" height={300} />
						</Stack>
					</div>
				}
			/>
		);
	}

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
								Configure which authentication methods are available for your Omnimesh Gateway instance.
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
															checked={authConfig.methods.jwt_enabled}
															onChange={(e) =>
																handleAuthMethodChange('jwt_enabled', e.target.checked)
															}
														/>
													}
													label="JWT Token Authentication"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.methods.api_keys_enabled}
															onChange={(e) =>
																handleAuthMethodChange('api_keys_enabled', e.target.checked)
															}
														/>
													}
													label="API Key Authentication"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.methods.oauth2_enabled}
															onChange={(e) =>
																handleAuthMethodChange('oauth2_enabled', e.target.checked)
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
															checked={authConfig.methods.mfa_required}
															onChange={(e) =>
																handleAuthMethodChange('mfa_required', e.target.checked)
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
										startIcon={isSaving ? <CircularProgress size={16} /> : <SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={
											isSaving ||
											(!authConfig.methods.jwt_enabled &&
												!authConfig.methods.api_keys_enabled &&
												!authConfig.methods.oauth2_enabled)
										}
									>
										{isSaving ? 'Saving Configuration...' : 'Save Configuration'}
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
									{authConfig.methods.jwt_enabled && (
										<Chip
											label="JWT"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.methods.api_keys_enabled && (
										<Chip
											label="API Keys"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.methods.oauth2_enabled && (
										<Chip
											label="OAuth 2.0"
											color="primary"
											size="small"
										/>
									)}
									{authConfig.methods.mfa_required && (
										<Chip
											label="MFA Required"
											color="warning"
											size="small"
										/>
									)}
									{!authConfig.methods.jwt_enabled &&
										!authConfig.methods.api_keys_enabled &&
										!authConfig.methods.oauth2_enabled && (
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
											value={authConfig.session.session_timeout_seconds}
											onChange={(e) =>
												handleSessionChange('session_timeout_seconds', parseInt(e.target.value))
											}
											fullWidth
											helperText="How long before idle sessions expire"
										/>

										<FormControl fullWidth>
											<InputLabel>Token Refresh Strategy</InputLabel>
											<Select
												value={authConfig.session.refresh_strategy}
												label="Token Refresh Strategy"
												onChange={(e) =>
													handleSessionChange('refresh_strategy', e.target.value)
												}
											>
												<MenuItem value="sliding">Sliding Window</MenuItem>
												<MenuItem value="fixed">Fixed Expiration</MenuItem>
												<MenuItem value="none">No Refresh</MenuItem>
											</Select>
										</FormControl>

										<TextField
											label="Max Concurrent Sessions"
											type="number"
											value={authConfig.session.max_concurrent_sessions}
											onChange={(e) =>
												handleSessionChange('max_concurrent_sessions', parseInt(e.target.value))
											}
											fullWidth
											helperText="Maximum number of concurrent sessions per user"
										/>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={isSaving ? <CircularProgress size={16} /> : <SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={isSaving}
									>
										{isSaving ? 'Saving Configuration...' : 'Save Configuration'}
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
												<TextField
													label="Minimum Password Length"
													type="number"
													value={authConfig.security.password_requirements.min_length}
													onChange={(e) => {
														const value = parseInt(e.target.value);
														if (value >= 6 && value <= 128) {
															handleSecurityChange('password_requirements', 'min_length', value);
														}
													}}
													size="small"
													sx={{ width: 200 }}
													inputProps={{ min: 6, max: 128 }}
													helperText="Must be between 6 and 128 characters"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.security.password_requirements.require_uppercase}
															onChange={(e) =>
																handleSecurityChange('password_requirements', 'require_uppercase', e.target.checked)
															}
														/>
													}
													label="Require uppercase letters"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.security.password_requirements.require_lowercase}
															onChange={(e) =>
																handleSecurityChange('password_requirements', 'require_lowercase', e.target.checked)
															}
														/>
													}
													label="Require lowercase letters"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.security.password_requirements.require_numbers}
															onChange={(e) =>
																handleSecurityChange('password_requirements', 'require_numbers', e.target.checked)
															}
														/>
													}
													label="Require numbers"
												/>
												<FormControlLabel
													control={
														<Checkbox
															checked={authConfig.security.password_requirements.require_special}
															onChange={(e) =>
																handleSecurityChange('password_requirements', 'require_special', e.target.checked)
															}
														/>
													}
													label="Require special characters"
												/>
												<TextField
													label="Password Max Age (days)"
													type="number"
													value={authConfig.security.password_requirements.max_age_days || ''}
													onChange={(e) =>
														handleSecurityChange('password_requirements', 'max_age_days', e.target.value ? parseInt(e.target.value) : null)
													}
													size="small"
													sx={{ width: 200 }}
													helperText="Leave empty for no expiration"
												/>
												<TextField
													label="Password History Count"
													type="number"
													value={authConfig.security.password_requirements.history_count}
													onChange={(e) =>
														handleSecurityChange('password_requirements', 'history_count', parseInt(e.target.value) || 0)
													}
													size="small"
													sx={{ width: 200 }}
													helperText="Number of previous passwords to remember"
												/>
											</FormGroup>
										</FormControl>

										<FormControl component="fieldset">
											<FormLabel component="legend">Account Security</FormLabel>
											<FormGroup>
												<FormControlLabel
													control={
														<Switch
															checked={authConfig.security.account_security.lockout_enabled}
															onChange={(e) =>
																handleSecurityChange('account_security', 'lockout_enabled', e.target.checked)
															}
														/>
													}
													label="Enable account lockout"
												/>
												<TextField
													label="Lockout Threshold (failed attempts)"
													type="number"
													value={authConfig.security.account_security.lockout_threshold}
													onChange={(e) => {
														const value = parseInt(e.target.value) || 5;
														if (value >= 1 && value <= 100) {
															handleSecurityChange('account_security', 'lockout_threshold', value);
														}
													}}
													size="small"
													sx={{ width: 250 }}
													disabled={!authConfig.security.account_security.lockout_enabled}
													inputProps={{ min: 1, max: 100 }}
													helperText="Number of failed attempts before lockout (1-100)"
												/>
												<TextField
													label="Lockout Duration (minutes)"
													type="number"
													value={authConfig.security.account_security.lockout_duration_minutes}
													onChange={(e) => {
														const value = parseInt(e.target.value) || 30;
														if (value >= 1 && value <= 1440) { // Max 24 hours
															handleSecurityChange('account_security', 'lockout_duration_minutes', value);
														}
													}}
													size="small"
													sx={{ width: 200 }}
													disabled={!authConfig.security.account_security.lockout_enabled}
													inputProps={{ min: 1, max: 1440 }}
													helperText="Duration in minutes (1-1440)"
												/>
												<FormControlLabel
													control={
														<Switch
															checked={authConfig.security.email_verification.required}
															onChange={(e) =>
																handleSecurityChange('email_verification', 'required', e.target.checked)
															}
														/>
													}
													label="Require email verification"
												/>
												<TextField
													label="Email Verification Expiry (hours)"
													type="number"
													value={authConfig.security.email_verification.expiry_hours}
													onChange={(e) => {
														const value = parseInt(e.target.value) || 24;
														if (value >= 1 && value <= 168) { // Max 7 days
															handleSecurityChange('email_verification', 'expiry_hours', value);
														}
													}}
													size="small"
													sx={{ width: 250 }}
													disabled={!authConfig.security.email_verification.required}
													inputProps={{ min: 1, max: 168 }}
													helperText="Expiry time in hours (1-168)"
												/>
												<FormControlLabel
													control={
														<Switch
															checked={authConfig.security.password_change_required}
															onChange={(e) =>
																handleSecurityChange('security', 'password_change_required', e.target.checked)
															}
														/>
													}
													label="Force password reset on next login"
												/>
											</FormGroup>
										</FormControl>

										<TextField
											label="IP Whitelist (comma-separated)"
											fullWidth
											multiline
											rows={3}
											value={authConfig.security.ip_restrictions.whitelist?.join(', ') || ''}
											onChange={(e) =>
												handleSecurityChange('ip_restrictions', 'whitelist',
													e.target.value.split(',').map(ip => ip.trim()).filter(ip => ip)
												)
											}
											placeholder="192.168.1.0/24, 10.0.0.0/8"
											helperText="Leave empty to allow all IPs"
										/>

										<FormControlLabel
											control={
												<Switch
													checked={authConfig.security.ip_restrictions.geo_blocking_enabled}
													onChange={(e) =>
														handleSecurityChange('ip_restrictions', 'geo_blocking_enabled', e.target.checked)
													}
												/>
											}
											label="Enable geographic blocking"
										/>

										<TextField
											label="Allowed Countries (comma-separated ISO codes)"
											fullWidth
											multiline
											rows={2}
											value={authConfig.security.ip_restrictions.allowed_countries?.join(', ') || ''}
											onChange={(e) =>
												handleSecurityChange('ip_restrictions', 'allowed_countries',
													e.target.value.split(',').map(country => country.trim().toUpperCase()).filter(country => country)
												)
											}
											placeholder="US, CA, GB, DE"
											helperText="Two-letter ISO country codes. Leave empty to allow all countries."
											disabled={!authConfig.security.ip_restrictions.geo_blocking_enabled}
										/>

										<FormControl fullWidth>
											<InputLabel>Compliance Mode</InputLabel>
											<Select
												value={authConfig.security.compliance_mode}
												label="Compliance Mode"
												onChange={(e) =>
													handleSecurityChange('security', 'compliance_mode', e.target.value)
												}
											>
												<MenuItem value="standard">Standard</MenuItem>
												<MenuItem value="strict">Strict</MenuItem>
												<MenuItem value="pci">PCI DSS</MenuItem>
												<MenuItem value="hipaa">HIPAA</MenuItem>
											</Select>
										</FormControl>
									</Stack>
								</CardContent>
								<CardActions>
									<Button
										variant="contained"
										startIcon={isSaving ? <CircularProgress size={16} /> : <SvgIcon>lucide:save</SvgIcon>}
										onClick={handleSaveAuthConfig}
										disabled={isSaving}
									>
										{isSaving ? 'Applying Security Settings...' : 'Apply Security Settings'}
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
