'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';
import { useAuth } from '@auth/AuthContext';
import SvgIcon from '@fuse/core/SvgIcon';
import { LoginRequest } from '@/lib/api';

const schema = z.object({
    email: z.string().email('Please enter a valid email address'),
    password: z.string().min(1, 'Please enter your password')
});

type FormType = z.infer<typeof schema>;

function MCPGatewaySignInForm() {
    const [showPassword, setShowPassword] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const { login } = useAuth();
    const router = useRouter();

    const {
        control,
        formState: { errors, isValid },
        handleSubmit,
    } = useForm<FormType>({
        mode: 'onChange',
        defaultValues: {
            email: '',
            password: '',
        },
        resolver: zodResolver(schema),
    });

    async function onSubmit(formData: FormType) {
        setIsLoading(true);
        setErrorMessage('');

        try {
            await login(formData as LoginRequest);
            
            // Check for redirect URL and navigate there, otherwise go to dashboard
            const redirectUrl = localStorage.getItem('redirectUrl');
            if (redirectUrl) {
                localStorage.removeItem('redirectUrl');
                router.replace(redirectUrl);
            } else {
                router.replace('/');
            }
        } catch (error) {
            console.error('Sign in error:', error);
            setErrorMessage(
                error instanceof Error 
                    ? error.message 
                    : 'Sign in failed. Please check your credentials and try again.'
            );
        } finally {
            setIsLoading(false);
        }
    }

    return (
        <div className="w-full">
            <form
                name="loginForm"
                noValidate
                className="flex flex-col justify-center w-full mt-8"
                onSubmit={handleSubmit(onSubmit)}
            >
                <Controller
                    name="email"
                    control={control}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            className="mb-6"
                            label="Email"
                            autoFocus
                            type="email"
                            error={!!errors.email}
                            helperText={errors?.email?.message}
                            variant="outlined"
                            required
                            fullWidth
                        />
                    )}
                />

                <Controller
                    name="password"
                    control={control}
                    render={({ field }) => (
                        <TextField
                            {...field}
                            className="mb-6"
                            label="Password"
                            type={showPassword ? 'text' : 'password'}
                            error={!!errors.password}
                            helperText={errors?.password?.message}
                            variant="outlined"
                            required
                            fullWidth
                            InputProps={{
                                endAdornment: (
                                    <InputAdornment position="end">
                                        <IconButton
                                            onClick={() => setShowPassword(!showPassword)}
                                            size="small"
                                        >
                                            <SvgIcon size={20}>
                                                {showPassword ? 'heroicons-solid:eye-slash' : 'heroicons-solid:eye'}
                                            </SvgIcon>
                                        </IconButton>
                                    </InputAdornment>
                                )
                            }}
                        />
                    )}
                />

                {errorMessage && (
                    <FormControl error className="mb-6">
                        <FormHelperText>{errorMessage}</FormHelperText>
                    </FormControl>
                )}

                <Button
                    variant="contained"
                    color="secondary"
                    className="w-full mt-4"
                    aria-label="Sign in"
                    disabled={!isValid || isLoading}
                    type="submit"
                    size="large"
                >
                    {isLoading ? 'Signing in...' : 'Sign in'}
                </Button>
            </form>
        </div>
    );
}

export default MCPGatewaySignInForm;