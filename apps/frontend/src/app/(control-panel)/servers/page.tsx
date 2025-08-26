import dynamic from 'next/dynamic';

const ServersView = dynamic(() => import('./ServersView'));

export default ServersView;