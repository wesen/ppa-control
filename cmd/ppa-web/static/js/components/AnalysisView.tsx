import { useAppStore } from '../store';
import { PacketTimeline } from './PacketTimeline';
import { MessageTypeChart } from './MessageTypeChart';
import { PacketTable } from './PacketTable';

export function AnalysisView() {
    const { currentAnalysis, selectedPcapFile } = useAppStore();

    if (!selectedPcapFile) {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <i className="bi bi-file-earmark display-1 text-muted mb-3"></i>
                    <h3 className="text-muted">No PCAP file selected</h3>
                    <p className="text-muted">Select a PCAP file from the sidebar to view its analysis.</p>
                </div>
            </div>
        );
    }

    if (selectedPcapFile.status === 'uploaded') {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <i className="bi bi-play-circle display-1 text-primary mb-3"></i>
                    <h3>Ready to analyze</h3>
                    <p className="text-muted mb-4">
                        Start the analysis for <strong>{selectedPcapFile.name}</strong>
                    </p>
                    <button
                        className="btn btn-primary btn-lg"
                        onClick={() => useAppStore.getState().startAnalysis(selectedPcapFile.id)}
                    >
                        <i className="bi bi-play-fill me-2"></i>
                        Start Analysis
                    </button>
                </div>
            </div>
        );
    }

    if (selectedPcapFile.status === 'analyzing') {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <div className="spinner-border text-primary mb-3" style={{ width: '3rem', height: '3rem' }}>
                        <span className="visually-hidden">Analyzing...</span>
                    </div>
                    <h3>Analysis in progress</h3>
                    <p className="text-muted mb-4">
                        Analyzing <strong>{selectedPcapFile.name}</strong>...
                    </p>
                    {selectedPcapFile.analysisProgress !== undefined && (
                        <div className="progress mx-auto" style={{ maxWidth: '400px', height: '8px' }}>
                            <div
                                className="progress-bar progress-bar-striped progress-bar-animated"
                                style={{ width: `${selectedPcapFile.analysisProgress}%` }}
                            ></div>
                        </div>
                    )}
                </div>
            </div>
        );
    }

    if (selectedPcapFile.status === 'error') {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <i className="bi bi-exclamation-triangle display-1 text-danger mb-3"></i>
                    <h3 className="text-danger">Analysis failed</h3>
                    <p className="text-muted mb-4">
                        Failed to analyze <strong>{selectedPcapFile.name}</strong>
                    </p>
                    {selectedPcapFile.errorMessage && (
                        <div className="alert alert-danger mx-auto" style={{ maxWidth: '600px' }}>
                            {selectedPcapFile.errorMessage}
                        </div>
                    )}
                    <button
                        className="btn btn-primary"
                        onClick={() => useAppStore.getState().startAnalysis(selectedPcapFile.id)}
                    >
                        <i className="bi bi-arrow-clockwise me-2"></i>
                        Retry Analysis
                    </button>
                </div>
            </div>
        );
    }

    if (!currentAnalysis) {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <div className="spinner-border text-primary mb-3">
                        <span className="visually-hidden">Loading...</span>
                    </div>
                    <h3>Loading analysis results...</h3>
                </div>
            </div>
        );
    }

    const duration = currentAnalysis.timeRange.end.getTime() - currentAnalysis.timeRange.start.getTime();

    return (
        <div className="container-fluid p-4">
            <div className="row mb-4">
                <div className="col">
                    <h1 className="h3 mb-0">
                        <i className="bi bi-bar-chart me-2"></i>
                        Analysis Results
                    </h1>
                    <p className="text-muted">{selectedPcapFile.name}</p>
                </div>
            </div>

            {/* Summary Cards */}
            <div className="row mb-4">
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h4 text-primary mb-1">{currentAnalysis.totalPackets.toLocaleString()}</div>
                            <div className="text-muted">Total Packets</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h4 text-info mb-1">{formatDuration(duration)}</div>
                            <div className="text-muted">Duration</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h4 text-success mb-1">
                                {(currentAnalysis.totalPackets / (duration / 1000)).toFixed(1)}
                            </div>
                            <div className="text-muted">Packets/sec</div>
                        </div>
                    </div>
                </div>
                <div className="col-sm-6 col-lg-3 mb-3">
                    <div className="card text-center">
                        <div className="card-body">
                            <div className="h4 text-warning mb-1">
                                {Object.keys(currentAnalysis.messageTypeDistribution).length}
                            </div>
                            <div className="text-muted">Message Types</div>
                        </div>
                    </div>
                </div>
            </div>

            <div className="row mb-4">
                {/* Packet Timeline */}
                <div className="col-lg-8 mb-4">
                    <div className="card">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-clock me-2"></i>
                                Packet Timeline
                            </h5>
                            <PacketTimeline analysis={currentAnalysis} />
                        </div>
                    </div>
                </div>

                {/* Message Type Distribution */}
                <div className="col-lg-4 mb-4">
                    <div className="card">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-pie-chart me-2"></i>
                                Message Types
                            </h5>
                            <MessageTypeChart distribution={currentAnalysis.messageTypeDistribution} />
                        </div>
                    </div>
                </div>
            </div>

            {/* Packet Table */}
            <div className="row">
                <div className="col-12">
                    <div className="card">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-table me-2"></i>
                                Packet Details
                            </h5>
                            <PacketTable packets={currentAnalysis.packets} />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
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
