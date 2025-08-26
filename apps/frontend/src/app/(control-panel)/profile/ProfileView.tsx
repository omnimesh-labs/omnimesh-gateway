'use client';

import { useState, useEffect } from 'react';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import { 
	Typography, 
	Button, 
	Card,
	CardContent,
	GridLegacy as Grid,
	Avatar,
	Box,
	Divider,
	List,
	ListItem,
	ListItemText,
	ListItemIcon,
	Chip,
	IconButton,
	TextField,
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Alert
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { useRouter } from 'next/navigation';
import useUser from '@auth/useUser';

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
	const [editFormData, setEditFormData] = useState({
		name: '',
		email: ''
	});
	const [passwordFormData, setPasswordFormData] = useState({
		currentPassword: '',
		newPassword: '',
		confirmPassword: ''
	});

	// Mock user profile data
	const mockProfile: UserProfile = {
		id: currentUser?.id || 'user-1',
		email: currentUser?.email || 'admin@admin.com',
		name: currentUser?.displayName || 'Admin User',
		role: Array.isArray(currentUser?.role) ? currentUser.role[0] : (currentUser?.role || 'admin'),
		organization: {
			id: 'org-1',
			name: 'MCP Gateway Organization',
			plan: 'Enterprise'
		},
		created_at: '2024-01-01T00:00:00Z',
		last_login: new Date().toISOString(),
		api_keys_count: 3,
		active_sessions: 5
	};

	const handleEditProfile = () => {
		setEditFormData({
			name: mockProfile.name,
			email: mockProfile.email
		});
		setEditDialogOpen(true);
	};

	const handleSaveProfile = async () => {
		try {
			// TODO: Replace with actual API call
			// await authApi.updateProfile(editFormData);
			enqueueSnackbar('Profile updated successfully', { variant: 'success' });
			setEditDialogOpen(false);
		} catch (error) {
			enqueueSnackbar('Failed to update profile', { variant: 'error' });
		}
	};

	const handleChangePassword = async () => {
		if (passwordFormData.newPassword !== passwordFormData.confirmPassword) {
			enqueueSnackbar('Passwords do not match', { variant: 'error' });
			return;
		}
		
		try {
			// TODO: Replace with actual API call
			// await authApi.changePassword(passwordFormData);
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
			description: `${mockProfile.api_keys_count} active keys`,
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
			description: `${mockProfile.active_sessions} active`,
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

	const accountDetails = [
		{ label: 'User ID', value: mockProfile.id },
		{ label: 'Email', value: mockProfile.email },
		{ label: 'Role', value: mockProfile.role, chip: true },
		{ label: 'Organization', value: mockProfile.organization.name },
		{ label: 'Plan', value: mockProfile.organization.plan, chip: true },
		{ label: 'Member Since', value: new Date(mockProfile.created_at).toLocaleDateString() },
		{ label: 'Last Login', value: new Date(mockProfile.last_login).toLocaleString() }
	];

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div className="flex items-center space-x-4">
							<Avatar 
								sx={{ width: 64, height: 64, bgcolor: 'primary.main' }}
							>
								{mockProfile.name.split(' ').map(n => n[0]).join('')}
							</Avatar>
							<div>
								<Typography variant="h4">{mockProfile.name}</Typography>
								<Typography variant="body1" color="textSecondary">
									{mockProfile.email}
								</Typography>
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
				<div className="p-6">
					<Grid container spacing={3}>
						{/* Quick Actions */}
						<Grid item xs={12}>
							<Typography variant="h6" className="mb-3">Quick Actions</Typography>
							<Grid container spacing={2}>
								{quickActions.map((action) => (
									<Grid item xs={12} sm={6} md={3} key={action.title}>
										<Card 
											sx={{ 
												cursor: 'pointer',
												'&:hover': { boxShadow: 3 }
											}}
											onClick={() => action.url !== '#' && router.push(action.url)}
										>
											<CardContent>
												<Box className="flex items-center justify-between mb-2">
													<SvgIcon size={24} color="primary">
														{action.icon}
													</SvgIcon>
													<SvgIcon size={16}>
														lucide:arrow-right
													</SvgIcon>
												</Box>
												<Typography variant="subtitle1">
													{action.title}
												</Typography>
												<Typography variant="body2" color="textSecondary">
													{action.description}
												</Typography>
											</CardContent>
										</Card>
									</Grid>
								))}
							</Grid>
						</Grid>

						{/* Account Details */}
						<Grid item xs={12} md={6}>
							<Card>
								<CardContent>
									<Typography variant="h6" className="mb-3">
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
						</Grid>

						{/* Security Settings */}
						<Grid item xs={12} md={6}>
							<Card>
								<CardContent>
									<Typography variant="h6" className="mb-3">
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
											<Button size="small" variant="outlined">
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
												secondary={`${mockProfile.api_keys_count} active keys`}
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
												secondary={`${mockProfile.active_sessions} devices`}
											/>
											<Button size="small" variant="outlined">
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
						</Grid>

						{/* Danger Zone */}
						<Grid item xs={12}>
							<Card sx={{ borderColor: 'error.main', borderWidth: 1, borderStyle: 'solid' }}>
								<CardContent>
									<Typography variant="h6" color="error" className="mb-3">
										Danger Zone
									</Typography>
									<Alert severity="warning" className="mb-3">
										These actions are irreversible. Please be certain.
									</Alert>
									<Box className="flex gap-2">
										<Button
											variant="outlined"
											color="error"
											startIcon={<SvgIcon>lucide:download</SvgIcon>}
											onClick={() => enqueueSnackbar('Export functionality coming soon', { variant: 'info' })}
										>
											Export Data
										</Button>
										<Button
											variant="outlined"
											color="error"
											startIcon={<SvgIcon>lucide:trash-2</SvgIcon>}
											onClick={() => enqueueSnackbar('Account deletion requires admin approval', { variant: 'warning' })}
										>
											Delete Account
										</Button>
									</Box>
								</CardContent>
							</Card>
						</Grid>
					</Grid>

					{/* Edit Profile Dialog */}
					<Dialog 
						open={editDialogOpen} 
						onClose={() => setEditDialogOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>Edit Profile</DialogTitle>
						<DialogContent>
							<Box className="space-y-4 mt-2">
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
							<Button variant="contained" onClick={handleSaveProfile}>Save Changes</Button>
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
							<Box className="space-y-4 mt-2">
								<TextField
									label="Current Password"
									type="password"
									value={passwordFormData.currentPassword}
									onChange={(e) => setPasswordFormData({ 
										...passwordFormData, 
										currentPassword: e.target.value 
									})}
									fullWidth
								/>
								<TextField
									label="New Password"
									type="password"
									value={passwordFormData.newPassword}
									onChange={(e) => setPasswordFormData({ 
										...passwordFormData, 
										newPassword: e.target.value 
									})}
									fullWidth
								/>
								<TextField
									label="Confirm New Password"
									type="password"
									value={passwordFormData.confirmPassword}
									onChange={(e) => setPasswordFormData({ 
										...passwordFormData, 
										confirmPassword: e.target.value 
									})}
									fullWidth
									error={passwordFormData.confirmPassword !== '' && 
										   passwordFormData.newPassword !== passwordFormData.confirmPassword}
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
