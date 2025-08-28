'use client';

import { useState, useEffect, useCallback } from 'react';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Card,
	CardContent,
	Avatar,
	Box,
	Divider,
	List,
	ListItem,
	ListItemText,
	ListItemIcon,
	Chip,
	TextField,
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { useRouter } from 'next/navigation';
import useUser from '@auth/useUser';
import { authApi } from '@/lib/auth-api';

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

interface UserProfile {
	id: string;
	email: string;
	name: string;
	role: string;
	organization: {
		id: string;
		name: string;
		plan: string;
	};
	created_at: string;
	last_login: string;
	api_keys_count: number;
	active_sessions: number;
}

function ProfileView() {
	const router = useRouter();
	const { enqueueSnackbar } = useSnackbar();
	const { data: currentUser } = useUser();
	const [editDialogOpen, setEditDialogOpen] = useState(false);
	const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
	const [profile, setProfile] = useState<UserProfile | null>(null);
	const [_isLoading, setIsLoading] = useState(false);
	const [editFormData, setEditFormData] = useState({
		name: '',
		email: ''
	});
	const [passwordFormData, setPasswordFormData] = useState({
		currentPassword: '',
		newPassword: '',
		confirmPassword: ''
	});

	const fetchProfile = useCallback(async () => {
		setIsLoading(true);
		try {
			const data = await authApi.getProfile();
			// Transform the API response to match UserProfile interface
			const transformedProfile: UserProfile = {
				id: data.id || currentUser?.id || '',
				email: data.email || currentUser?.email || '',
				name: currentUser?.displayName || 'User',
				role: data.role || currentUser?.role || 'user',
				organization: {
					id: data.organization_id || '',
					name: 'Organization',
					plan: 'free'
				},
				created_at: data.created_at || new Date().toISOString(),
				last_login: new Date().toISOString(),
				api_keys_count: 0,
				active_sessions: 0
			};
			setProfile(transformedProfile);
		} catch (_error) {
			enqueueSnackbar('Failed to fetch profile data', { variant: 'error' });
			console.error('Error fetching profile:', _error);
		} finally {
			setIsLoading(false);
		}
	}, [currentUser, enqueueSnackbar]);

	// Fetch profile data on mount
	useEffect(() => {
		fetchProfile();
	}, [fetchProfile]);

	const handleEditProfile = () => {
		if (profile) {
			setEditFormData({
				name: profile.name,
				email: profile.email
			});
			setEditDialogOpen(true);
		}
	};

	const handleSaveProfile = async () => {
		try {
			await authApi.updateProfile(editFormData);
			enqueueSnackbar('Profile updated successfully', { variant: 'success' });
			setEditDialogOpen(false);
			fetchProfile();
		} catch (_error) {
			enqueueSnackbar('Failed to update profile', { variant: 'error' });
		}
	};

	const handleChangePassword = async () => {
		if (passwordFormData.newPassword !== passwordFormData.confirmPassword) {
			enqueueSnackbar('Passwords do not match', { variant: 'error' });
			return;
		}

		try {
			await authApi.updateProfile({
				current_password: passwordFormData.currentPassword,
				new_password: passwordFormData.newPassword
			});
			enqueueSnackbar('Password changed successfully', { variant: 'success' });
			setPasswordDialogOpen(false);
			setPasswordFormData({
				currentPassword: '',
				newPassword: '',
				confirmPassword: ''
			});
		} catch (error) {
			enqueueSnackbar('Failed to change password', { variant: 'error' });
		}
	};

	const quickActions = [
		{
			title: 'API Keys',
			description: `${profile?.api_keys_count || 0} active keys`,
			icon: 'lucide:key',
			url: '/profile/api-keys'
		},
		{
			title: 'Settings',
			description: 'Account preferences',
			icon: 'lucide:settings',
			url: '/profile/settings'
		},
		{
			title: 'Sessions',
			description: `${profile?.active_sessions || 0} active`,
			icon: 'lucide:monitor',
			url: '#'
		},
		{
			title: 'Activity Log',
			description: 'View your activity',
			icon: 'lucide:activity',
			url: '/logs'
		}
	];

	const accountDetails = profile
		? [
				{ label: 'User ID', value: profile.id },
				{ label: 'Email', value: profile.email },
				{ label: 'Role', value: profile.role, chip: true },
				{ label: 'Organization', value: profile.organization.name },
				{ label: 'Plan', value: profile.organization.plan, chip: true },
				{ label: 'Member Since', value: new Date(profile.created_at).toLocaleDateString() },
				{ label: 'Last Login', value: new Date(profile.last_login).toLocaleString() }
			]
		: [];

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div className="flex items-center space-x-4">
							<Avatar sx={{ width: 64, height: 64, bgcolor: 'primary.main' }}>
								{profile?.email?.split('@')[0]?.charAt(0)}
							</Avatar>
							<div>
								<Typography variant="h4">{profile?.email}</Typography>
								{/* <Typography
									variant="body1"
									color="textSecondary"
								>
									{profile.email}
								</Typography> */}
							</div>
						</div>
						<Box className="flex gap-2">
							<Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:edit</SvgIcon>}
								onClick={handleEditProfile}
							>
								Edit Profile
							</Button>
							<Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:lock</SvgIcon>}
								onClick={() => setPasswordDialogOpen(true)}
							>
								Change Password
							</Button>
						</Box>
					</div>
				</div>
			}
			content={
				<div className="space-y-6 p-6">
					{/* Quick Actions */}
					<div>
						<Typography
							variant="h6"
							className="mb-3"
						>
							Quick Actions
						</Typography>
						<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
							{quickActions.map((action) => (
								<Card
									key={action.title}
									sx={{
										cursor: 'pointer',
										'&:hover': { boxShadow: 3 }
									}}
									onClick={() => action.url !== '#' && router.push(action.url)}
								>
									<CardContent>
										<Box className="mb-2 flex items-center justify-between">
											<SvgIcon
												size={24}
												color="primary"
											>
												{action.icon}
											</SvgIcon>
											<SvgIcon size={16}>lucide:arrow-right</SvgIcon>
										</Box>
										<Typography variant="subtitle1">{action.title}</Typography>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											{action.description}
										</Typography>
									</CardContent>
								</Card>
							))}
						</div>
					</div>

					{/* Cards Section */}
					<div className="grid grid-cols-1 gap-6 md:grid-cols-2">
						{/* Account Details */}
						<Card>
							<CardContent>
								<Typography
									variant="h6"
									className="mb-3"
								>
									Account Details
								</Typography>
								<List>
									{accountDetails.map((detail, index) => (
										<div key={detail.label}>
											<ListItem>
												<ListItemText
													primary={detail.label}
													secondary={
														detail.chip ? (
															<Chip
																size="small"
																label={detail.value}
																color={detail.label === 'Role' ? 'primary' : 'default'}
															/>
														) : (
															detail.value
														)
													}
												/>
											</ListItem>
											{index < accountDetails.length - 1 && <Divider />}
										</div>
									))}
								</List>
							</CardContent>
						</Card>

						{/* Security Settings */}
						<Card>
							<CardContent>
								<Typography
									variant="h6"
									className="mb-3"
								>
									Security Settings
								</Typography>
								<List>
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:shield-check</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Two-Factor Authentication"
											secondary="Not enabled"
										/>
										<Button
											size="small"
											variant="outlined"
										>
											Enable
										</Button>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:key</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="API Keys"
											secondary={`${profile?.api_keys_count || 0} active keys`}
										/>
										<Button
											size="small"
											variant="outlined"
											onClick={() => router.push('/profile/api-keys')}
										>
											Manage
										</Button>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:monitor</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Active Sessions"
											secondary={`${profile?.active_sessions} devices`}
										/>
										<Button
											size="small"
											variant="outlined"
										>
											View All
										</Button>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:clock</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Session Timeout"
											secondary="24 hours"
										/>
									</ListItem>
								</List>
							</CardContent>
						</Card>
					</div>

					{/* Edit Profile Dialog */}
					<Dialog
						open={editDialogOpen}
						onClose={() => setEditDialogOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>Edit Profile</DialogTitle>
						<DialogContent>
							<Box className="mt-2 space-y-4">
								<TextField
									label="Name"
									value={editFormData.name}
									onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
									fullWidth
								/>
								<TextField
									label="Email"
									type="email"
									value={editFormData.email}
									onChange={(e) => setEditFormData({ ...editFormData, email: e.target.value })}
									fullWidth
								/>
							</Box>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setEditDialogOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSaveProfile}
							>
								Save Changes
							</Button>
						</DialogActions>
					</Dialog>

					{/* Change Password Dialog */}
					<Dialog
						open={passwordDialogOpen}
						onClose={() => setPasswordDialogOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>Change Password</DialogTitle>
						<DialogContent>
							<Box className="mt-2 space-y-4">
								<TextField
									label="Current Password"
									type="password"
									value={passwordFormData.currentPassword}
									onChange={(e) =>
										setPasswordFormData({
											...passwordFormData,
											currentPassword: e.target.value
										})
									}
									fullWidth
								/>
								<TextField
									label="New Password"
									type="password"
									value={passwordFormData.newPassword}
									onChange={(e) =>
										setPasswordFormData({
											...passwordFormData,
											newPassword: e.target.value
										})
									}
									fullWidth
								/>
								<TextField
									label="Confirm New Password"
									type="password"
									value={passwordFormData.confirmPassword}
									onChange={(e) =>
										setPasswordFormData({
											...passwordFormData,
											confirmPassword: e.target.value
										})
									}
									fullWidth
									error={
										passwordFormData.confirmPassword !== '' &&
										passwordFormData.newPassword !== passwordFormData.confirmPassword
									}
									helperText={
										passwordFormData.confirmPassword !== '' &&
										passwordFormData.newPassword !== passwordFormData.confirmPassword
											? 'Passwords do not match'
											: ''
									}
								/>
							</Box>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setPasswordDialogOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleChangePassword}
								disabled={
									!passwordFormData.currentPassword ||
									!passwordFormData.newPassword ||
									passwordFormData.newPassword !== passwordFormData.confirmPassword
								}
							>
								Change Password
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ProfileView;
