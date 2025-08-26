'use client';

import { useState } from 'react';
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
	Alert
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { authApi } from '@/lib/api';

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
		email: 'user@example.com',
		name: 'John Doe',
		current_password: '',
		new_password: '',
		confirm_password: ''
	});
	const [isLoading, setIsLoading] = useState(false);
	const { enqueueSnackbar } = useSnackbar();

	// Mock user data
	const mockUser = {
		id: 'user-123',
		email: 'user@example.com',
		name: 'John Doe',
		role: 'admin',
		organization_id: 'org-456',
		created_at: '2024-01-01T00:00:00Z',
		is_active: true
	};

	const handleProfileUpdate = async (e: React.FormEvent) => {
		e.preventDefault();

		if (profileForm.new_password && profileForm.new_password !== profileForm.confirm_password) {
			enqueueSnackbar('New passwords do not match', { variant: 'error' });
			return;
		}

		setIsLoading(true);
		try {
			// Prepare update data
			const updateData: any = {};

			if (profileForm.email) updateData.email = profileForm.email;

			if (profileForm.new_password) {
				updateData.current_password = profileForm.current_password;
				updateData.new_password = profileForm.new_password;
			}

			await authApi.updateProfile(updateData);
			enqueueSnackbar('Profile updated successfully', { variant: 'success' });
			setProfileForm((prev) => ({
				...prev,
				current_password: '',
				new_password: '',
				confirm_password: ''
			}));
		} catch (error) {
			enqueueSnackbar('Failed to update profile', { variant: 'error' });
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
										gutterBottom
										className="flex items-center gap-2"
									>
										<SvgIcon>lucide:user</SvgIcon>
										Account Information
									</Typography>

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
												{mockUser.id}
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
													{mockUser.role}
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
												{mockUser.organization_id}
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
												{new Date(mockUser.created_at).toLocaleDateString()}
											</Paper>
										</Box>
									</Stack>
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
											gutterBottom
											className="flex items-center gap-2"
										>
											<SvgIcon>lucide:edit</SvgIcon>
											Update Profile
										</Typography>

										<Stack spacing={3}>
											<TextField
												label="Name"
												value={profileForm.name}
												onChange={(e) =>
													setProfileForm((prev) => ({ ...prev, name: e.target.value }))
												}
												fullWidth
												required
											/>

											<TextField
												label="Email Address"
												type="email"
												value={profileForm.email}
												onChange={(e) =>
													setProfileForm((prev) => ({ ...prev, email: e.target.value }))
												}
												fullWidth
												required
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
												placeholder="Enter current password to make changes"
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
											/>
										</Stack>
									</CardContent>
									<CardActions>
										<Button
											type="submit"
											variant="contained"
											disabled={isLoading}
											startIcon={<SvgIcon>lucide:save</SvgIcon>}
										>
											{isLoading ? 'Updating...' : 'Update Profile'}
										</Button>
									</CardActions>
								</form>
							</Card>
						</Grid>
					</Grid>
				</div>
			}
		/>
	);
}

export default ProfileSettingsView;
