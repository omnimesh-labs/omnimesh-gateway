'use client';
import DemoContent from '@fuse/core/DemoContent';
import PageSimple from '@fuse/core/PageSimple';
import { useTranslation } from 'react-i18next';
import { styled } from '@mui/material/styles';
import '../../i18n';

const Root = styled(PageSimple)(({ theme }) => ({
	'& .PageSimple-header': {
		backgroundColor: theme.vars.palette.background.paper,
		borderBottomWidth: 1,
		borderStyle: 'solid',
		borderColor: theme.vars.palette.divider
	},
	'& .PageSimple-content': {},
	'& .PageSimple-sidebarHeader': {},
	'& .PageSimple-sidebarContent': {}
}));

function ExampleView() {
	const { t } = useTranslation('examplePage');

	return (
		<Root
			header={
				<div className="p-6">
					<h4>{t('TITLE')}</h4>
				</div>
			}
			content={
				<div className="p-6">
					<h4>Content</h4>
					<br />
					<DemoContent />
				</div>
			}
		/>
	);
}

export default ExampleView;
