'use client';

import { useState, useEffect } from 'react';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Paper,
	Box,
	Grid,
	TextField,
	Stack,
	Card,
	CardContent,
	CardActions,
	Divider,
	Alert,
	CircularProgress
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { authApi, type User } from '@/lib/api';

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

function ProfileSettingsView() {
	const [profileForm, setProfileForm] = useState({
		email: '',
		current_password: '',
		new_password: '',
		confirm_password: ''
	});
	const [user, setUser] = useState<User | null>(null);
	const [isLoading, setIsLoading] = useState(false);
	const [isLoadingProfile, setIsLoadingProfile] = useState(true);
	const { enqueueSnackbar } = useSnackbar();

	// Fetch user profile data on mount
	useEffect(() => {
		const fetchProfile = async () => {
			try {
				const profileData = await authApi.getProfile();
				setUser(profileData);
				setProfileForm(prev => ({
					...prev,
					email: profileData.email
				}));
			} catch (_error) {
				enqueueSnackbar('Failed to load profile data', { variant: 'error' });
			} finally {
				setIsLoadingProfile(false);
			}
		};

		fetchProfile();
	}, [enqueueSnackbar]);

	const handleProfileUpdate = async (e: React.FormEvent) => {
		e.preventDefault();

		if (profileForm.new_password && profileForm.new_password !== profileForm.confirm_password) {
			enqueueSnackbar('New passwords do not match', { variant: 'error' });
			return;
		}

		setIsLoading(true);
		try {
			// Prepare update data
			const updateData: { email?: string; current_password?: string; new_password?: string } = {};

			if (profileForm.email !== user?.email) {
				updateData.email = profileForm.email;
			}

			if (profileForm.new_password) {
				if (!profileForm.current_password) {
					enqueueSnackbar('Current password is required to change password', { variant: 'error' });
					setIsLoading(false);
					return;
				}
				updateData.current_password = profileForm.current_password;
				updateData.new_password = profileForm.new_password;
			}

			// Only make API call if there are changes
			if (Object.keys(updateData).length === 0) {
				enqueueSnackbar('No changes to save', { variant: 'info' });
				setIsLoading(false);
				return;
			}

			const updatedUser = await authApi.updateProfile(updateData);
			setUser(updatedUser);
			enqueueSnackbar('Profile updated successfully', { variant: 'success' });
			setProfileForm((prev) => ({
				...prev,
				current_password: '',
				new_password: '',
				confirm_password: ''
			}));
		} catch (error: unknown) {
			enqueueSnackbar(error instanceof Error ? error.message : 'Failed to update profile', { variant: 'error' });
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Profile</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your account settings and preferences
							</Typography>
						</div>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					{isLoadingProfile ? (
						<div className="flex justify-center items-center h-64">
							<CircularProgress />
						</div>
					) : (
						<Grid
							container
							spacing={4}
							maxWidth="4xl"
							className="mx-auto"
						>
							{/* Account Information */}
							<Grid size={{ xs: 12, md: 6 }}>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											className="flex items-center gap-2"
											sx={{ marginBottom: '24px !important' }}
										>
											<SvgIcon>lucide:user</SvgIcon>
											Account Information
										</Typography>

										{user && (
											<Stack spacing={2}>
												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														User ID
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2 font-mono text-sm"
													>
														{user.id}
													</Paper>
												</Box>

												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														Email
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2"
													>
														{user.email}
													</Paper>
												</Box>

												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														Role
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2"
													>
														<Typography
															variant="body2"
															className="inline-block rounded bg-blue-100 px-2 py-1 text-xs font-medium capitalize text-blue-800"
														>
															{user.role}
														</Typography>
													</Paper>
												</Box>

												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														Organization ID
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2 font-mono text-sm"
													>
														{user.organization_id}
													</Paper>
												</Box>

												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														Account Status
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2"
													>
														<Typography
															variant="body2"
															className={`inline-block rounded px-2 py-1 text-xs font-medium ${
																user.is_active
																	? 'bg-green-100 text-green-800'
																	: 'bg-red-100 text-red-800'
															}`}
														>
															{user.is_active ? 'Active' : 'Inactive'}
														</Typography>
													</Paper>
												</Box>

												<Box>
													<Typography
														variant="body2"
														color="textSecondary"
													>
														Account Created
													</Typography>
													<Paper
														variant="outlined"
														className="bg-gray-50 p-2"
													>
														{new Date(user.created_at).toLocaleDateString()}
													</Paper>
												</Box>
											</Stack>
										)}
									</CardContent>
								</Card>
							</Grid>

							{/* Update Profile */}
							<Grid size={{ xs: 12, md: 6 }}>
								<Card>
									<form onSubmit={handleProfileUpdate}>
										<CardContent>
											<Typography
												variant="h6"
												className="flex items-center gap-2"
												sx={{ marginBottom: '24px !important' }}
											>
												<SvgIcon>lucide:edit</SvgIcon>
												Update Profile
											</Typography>

											<Stack spacing={3}>
												<TextField
													label="Email Address"
													type="email"
													value={profileForm.email}
													onChange={(e) =>
														setProfileForm((prev) => ({ ...prev, email: e.target.value }))
													}
													fullWidth
													required
													disabled={isLoadingProfile}
												/>

												<Divider />

												<Alert
													severity="info"
													sx={{ fontSize: '0.875rem' }}
												>
													Leave password fields empty to keep your current password
												</Alert>

												<TextField
													label="Current Password"
													type="password"
													value={profileForm.current_password}
													onChange={(e) =>
														setProfileForm((prev) => ({
															...prev,
															current_password: e.target.value
														}))
													}
													fullWidth
													placeholder="Required only when changing password"
													helperText="Enter current password to change email or password"
												/>

												<TextField
													label="New Password"
													type="password"
													value={profileForm.new_password}
													onChange={(e) =>
														setProfileForm((prev) => ({
															...prev,
															new_password: e.target.value
														}))
													}
													fullWidth
													placeholder="Leave blank to keep current password"
													helperText="Minimum 8 characters"
												/>

												<TextField
													label="Confirm New Password"
													type="password"
													value={profileForm.confirm_password}
													onChange={(e) =>
														setProfileForm((prev) => ({
															...prev,
															confirm_password: e.target.value
														}))
													}
													fullWidth
													placeholder="Confirm new password"
													error={profileForm.new_password !== '' && profileForm.new_password !== profileForm.confirm_password}
													helperText={
														profileForm.new_password !== '' && profileForm.new_password !== profileForm.confirm_password
															? 'Passwords do not match'
															: ''
													}
												/>
											</Stack>
										</CardContent>
										<CardActions>
											<Button
												type="submit"
												variant="contained"
												disabled={isLoading || isLoadingProfile}
												startIcon={<SvgIcon>lucide:save</SvgIcon>}
											>
												{isLoading ? 'Updating...' : 'Update Profile'}
											</Button>
										</CardActions>
									</form>
								</Card>
							</Grid>
						</Grid>
					)}
				</div>
			}
		/>
	);
}

export default ProfileSettingsView;
