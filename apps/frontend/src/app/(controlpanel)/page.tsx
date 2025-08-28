import { redirect } from 'next/navigation';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function DashboardPage() {
	redirect('/dashboard');
}
