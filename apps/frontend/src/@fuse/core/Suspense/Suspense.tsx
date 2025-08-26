import Loading from '@fuse/core/Loading';
import { ReactNode, Suspense } from 'react';
import { LoadingProps } from '@fuse/core/Loading/Loading';

export type AppSuspenseProps = {
	loadingProps?: LoadingProps;
	children: ReactNode;
};

/**
 * The AppSuspense component is a wrapper around the React Suspense component.
 * It is used to display a loading spinner while the wrapped components are being loaded.
 * The component is memoized to prevent unnecessary re-renders.
 * React Suspense defaults
 * For to Avoid Repetition
 */
function AppSuspense(props: AppSuspenseProps) {
	const { children, loadingProps } = props;
	return <Suspense fallback={<Loading {...loadingProps} />}>{children}</Suspense>;
}

export default AppSuspense;
