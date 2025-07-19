import { useAppStore } from '../store';
import { FileDropZone } from './FileDropZone';

export function Dashboard() {
    const { pcapFiles, analysisResults } = useAppStore();

    const recentFiles = pcapFiles.slice(-5);
    const recentAnalyses = analysisResults.slice(-3);

    const stats = {
        totalFiles: pcapFiles.length,
        analyzedFiles: pcapFiles.filter(f => f.status === 'analyzed').length,
        totalAnalyses: analysisResults.length,
        totalPackets: analysisResults.reduce((sum, a) => sum + a.totalPackets, 0)
    };

    return (
        <div className="container-fluid p-4">
            <div className="row mb-4">
                <div className="col">
                    <h1 className="h3 mb-0">
                        <i className="bi bi-speedometer2 me-2"></i>
                        Dashboard
                    </h1>
                    <p className="text-muted">Overview of your packet analysis activities</p>
                </div>
            </div>

            {/* Statistics Cards */}
            <div className="row mb-4">
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h2 text-primary mb-1">{stats.totalFiles}</div>
                            <div className="text-muted">PCAP Files</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h2 text-success mb-1">{stats.analyzedFiles}</div>
                            <div className="text-muted">Analyzed</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h2 text-info mb-1">{stats.totalAnalyses}</div>
                            <div className="text-muted">Analysis Reports</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h2 text-warning mb-1">{stats.totalPackets.toLocaleString()}</div>
                            <div className="text-muted">Total Packets</div>
                        </div>
                    </div>
                </div>
            </div>

            <div className="row">
                {/* File Upload */}
                <div className="col-lg-6 mb-4">
                    <div className="card h-100">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-cloud-upload me-2"></i>
                                Upload PCAP File
                            </h5>
                            <p className="card-text text-muted">
                                Drag and drop a PCAP file here or click to browse
                            </p>
                            <FileDropZone />
                        </div>
                    </div>
                </div>

                {/* Recent Files */}
                <div className="col-lg-6 mb-4">
                    <div className="card h-100">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-clock-history me-2"></i>
                                Recent Files
                            </h5>
                            {recentFiles.length === 0 ? (
                                <p className="text-muted">No files uploaded yet</p>
                            ) : (
                                <div className="list-group list-group-flush">
                                    {recentFiles.map(file => (
                                        <div key={file.id} className="list-group-item px-0">
                                            <div className="d-flex align-items-center">
                                                <i className={`bi ${getStatusIcon(file.status)} me-3`}></i>
                                                <div className="flex-grow-1">
                                                    <div className="fw-semibold">{file.name}</div>
                                                    <div className="small text-muted">
                                                        {file.uploadDate.toLocaleString()}
                                                    </div>
                                                </div>
                                                <span className={`badge ${getStatusBadgeClass(file.status)}`}>
                                                    {file.status}
                                                </span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Recent Analyses */}
            {recentAnalyses.length > 0 && (
                <div className="row">
                    <div className="col-12">
                        <div className="card">
                            <div className="card-body">
                                <h5 className="card-title">
                                    <i className="bi bi-bar-chart me-2"></i>
                                    Recent Analyses
                                </h5>
                                <div className="table-responsive">
                                    <table className="table table-hover">
                                        <thead>
                                            <tr>
                                                <th>Analysis</th>
                                                <th>Total Packets</th>
                                                <th>Time Range</th>
                                                <th>Created</th>
                                                <th>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {recentAnalyses.map(analysis => {
                                                const file = pcapFiles.find(f => f.id === analysis.pcapFileId);
                                                return (
                                                    <tr key={analysis.id}>
                                                        <td>
                                                            <div className="fw-semibold">{file?.name || 'Unknown'}</div>
                                                            <div className="small text-muted">{analysis.id}</div>
                                                        </td>
                                                        <td>{analysis.totalPackets.toLocaleString()}</td>
                                                        <td>
                                                            <div className="small">
                                                                {formatDuration(analysis.timeRange.end.getTime() - analysis.timeRange.start.getTime())}
                                                            </div>
                                                        </td>
                                                        <td>{analysis.createdAt.toLocaleDateString()}</td>
                                                        <td>
                                                            <button
                                                                className="btn btn-sm btn-outline-primary"
                                                                onClick={() => {
                                                                    useAppStore.getState().setCurrentAnalysis(analysis);
                                                                    useAppStore.getState().setActiveView('analysis');
                                                                }}
                                                            >
                                                                View
                                                            </button>
                                                        </td>
                                                    </tr>
                                                );
                                            })}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

function getStatusIcon(status: string) {
    switch (status) {
        case 'uploaded':
            return 'bi-file-earmark text-primary';
        case 'analyzing':
            return 'bi-hourglass-split text-warning';
        case 'analyzed':
            return 'bi-check-circle-fill text-success';
        case 'error':
            return 'bi-exclamation-triangle-fill text-danger';
        default:
            return 'bi-file-earmark';
    }
}

function getStatusBadgeClass(status: string) {
    switch (status) {
        case 'uploaded':
            return 'bg-primary';
        case 'analyzing':
            return 'bg-warning';
        case 'analyzed':
            return 'bg-success';
        case 'error':
            return 'bg-danger';
        default:
            return 'bg-secondary';
    }
}

function formatDuration(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes % 60}m`;
    } else if (minutes > 0) {
        return `${minutes}m ${seconds % 60}s`;
    } else {
        return `${seconds}s`;
    }
}
