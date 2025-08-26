import { useEffect, useState } from 'react';

type AwaitRenderProps = {
	delay?: number;
	children: React.ReactNode;
};

function AwaitRender(props: AwaitRenderProps) {
	const { delay = 0, children } = props;
	const [awaitRender, setAwaitRender] = useState(true);

	useEffect(() => {
		setTimeout(() => {
			setAwaitRender(false);
		}, delay);
	}, [delay]);

	return awaitRender ? null : children;
}

export default AwaitRender;
