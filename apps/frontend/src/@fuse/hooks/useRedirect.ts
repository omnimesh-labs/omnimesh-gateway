import { useRouter } from 'next/navigation';

function useRedirect() {
	const router = useRouter();

	return router.push;
}

export default useRedirect;
