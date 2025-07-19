import { useAppStore } from '../store';
import { Layout } from './Layout';
import { Dashboard } from './Dashboard';
import { AnalysisView } from './AnalysisView';
import { PacketsView } from './PacketsView';
import { DocumentsView } from './DocumentsView';

export function App() {
    const activeView = useAppStore(state => state.activeView);

    const renderCurrentView = () => {
        switch (activeView) {
            case 'dashboard':
                return <Dashboard />;
            case 'analysis':
                return <AnalysisView />;
            case 'packets':
                return <PacketsView />;
            case 'documents':
                return <DocumentsView />;
            default:
                return <Dashboard />;
        }
    };

    return (
        <Layout>
            {renderCurrentView()}
        </Layout>
    );
}
