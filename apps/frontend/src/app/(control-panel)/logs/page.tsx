import dynamic from 'next/dynamic';

const LogsView = dynamic(() => import('./LogsView'));

export default LogsView;
