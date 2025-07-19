import { useAppStore } from '../store';

export function AnalysisStatusPanel() {
    const analysisStatus = useAppStore(state => state.analysisStatus);
    const pcapFiles = useAppStore(state => state.pcapFiles);

    const activeAnalyses = Object.values(analysisStatus).filter(status => 
        status.status === 'analyzing'
    );

    if (activeAnalyses.length === 0) {
        return (
            <div className="text-muted text-center py-3">
                <i className="bi bi-clipboard-check d-block mb-2"></i>
                No active analyses
            </div>
        );
    }

    return (
        <div className="list-group list-group-flush">
            {activeAnalyses.map(status => {
                const file = pcapFiles.find(f => f.id === status.pcapFileId);
                return (
                    <div key={status.pcapFileId} className="list-group-item border-0 px-0">
                        <div className="d-flex align-items-center mb-2">
                            <div className="spinner-border spinner-border-sm text-primary me-2" role="status">
                                <span className="visually-hidden">Analyzing...</span>
                            </div>
                            <div className="flex-grow-1 min-w-0">
                                <div className="fw-semibold text-truncate" title={file?.name}>
                                    {file?.name || 'Unknown file'}
                                </div>
                            </div>
                        </div>
                        
                        <div className="progress mb-2" style={{ height: '6px' }}>
                            <div
                                className="progress-bar progress-bar-striped progress-bar-animated"
                                style={{ width: `${status.progress}%` }}
                            ></div>
                        </div>
                        
                        <div className="small text-muted">
                            {status.message}
                            {status.estimatedTimeRemaining && (
                                <div className="mt-1">
                                    <i className="bi bi-clock me-1"></i>
                                    ~{Math.ceil(status.estimatedTimeRemaining / 1000)}s remaining
                                </div>
                            )}
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
