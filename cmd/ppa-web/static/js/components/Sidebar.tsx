import { useAppStore } from '../store';
import { PCAPFileList } from './PCAPFileList';
import { AnalysisStatusPanel } from './AnalysisStatusPanel';
import { SearchPanel } from './SearchPanel';

export function Sidebar() {
    const activeView = useAppStore(state => state.activeView);

    return (
        <div className="p-3">
            <div className="mb-4">
                <h6 className="text-muted text-uppercase fw-bold mb-3">
                    <i className="bi bi-files me-2"></i>
                    PCAP Files
                </h6>
                <PCAPFileList />
            </div>

            {activeView === 'analysis' && (
                <div className="mb-4">
                    <h6 className="text-muted text-uppercase fw-bold mb-3">
                        <i className="bi bi-activity me-2"></i>
                        Analysis Status
                    </h6>
                    <AnalysisStatusPanel />
                </div>
            )}

            {activeView === 'documents' && (
                <div className="mb-4">
                    <h6 className="text-muted text-uppercase fw-bold mb-3">
                        <i className="bi bi-search me-2"></i>
                        Search
                    </h6>
                    <SearchPanel />
                </div>
            )}

            <div className="mt-auto pt-4">
                <div className="small text-muted">
                    <div className="d-flex align-items-center mb-2">
                        <span className="status-indicator status-connected"></span>
                        Backend Connected
                    </div>
                    <div>
                        <i className="bi bi-clock me-1"></i>
                        {new Date().toLocaleTimeString()}
                    </div>
                </div>
            </div>
        </div>
    );
}
