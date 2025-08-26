import dynamic from 'next/dynamic';

const PoliciesView = dynamic(() => import('./PoliciesView'));

export default PoliciesView;
