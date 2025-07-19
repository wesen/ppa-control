import { useAppStore } from '../store';
import { HexDump } from './HexDump';

export function PacketsView() {
    const { selectedPacket, currentAnalysis } = useAppStore();

    if (!currentAnalysis) {
        return (
            <div className="container-fluid p-4">
                <div className="text-center py-5">
                    <i className="bi bi-diagram-3 display-1 text-muted mb-3"></i>
                    <h3 className="text-muted">No analysis selected</h3>
                    <p className="text-muted">Select an analysis to view packet details.</p>
                </div>
            </div>
        );
    }

    if (!selectedPacket) {
        return (
            <div className="container-fluid p-4">
                <div className="row mb-4">
                    <div className="col">
                        <h1 className="h3 mb-0">
                            <i className="bi bi-diagram-3 me-2"></i>
                            Packet Details
                        </h1>
                        <p className="text-muted">Select a packet from the analysis view to see detailed information</p>
                    </div>
                </div>

                <div className="text-center py-5">
                    <i className="bi bi-cursor display-1 text-muted mb-3"></i>
                    <h4 className="text-muted">No packet selected</h4>
                    <p className="text-muted">Click on a packet in the timeline or table to view its details.</p>
                    <button
                        className="btn btn-primary"
                        onClick={() => useAppStore.getState().setActiveView('analysis')}
                    >
                        <i className="bi bi-arrow-left me-2"></i>
                        Back to Analysis
                    </button>
                </div>
            </div>
        );
    }

    const getTypeIcon = (type: string) => {
        switch (type) {
            case 'request':
                return 'bi-arrow-right-circle text-primary';
            case 'response':
                return 'bi-arrow-left-circle text-success';
            case 'error':
                return 'bi-exclamation-triangle text-danger';
            default:
                return 'bi-circle text-secondary';
        }
    };

    const formatTimestamp = (timestamp: Date) => {
        return timestamp.toLocaleString([], {
            hour12: false,
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    };

    const formatBytes = (bytes: number) => {
        const units = ['B', 'KB', 'MB'];
        let size = bytes;
        let unitIndex = 0;
        
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        
        return `${size.toFixed(1)} ${units[unitIndex]}`;
    };

    return (
        <div className="container-fluid p-4">
            <div className="row mb-4">
                <div className="col">
                    <div className="d-flex align-items-center">
                        <button
                            className="btn btn-outline-secondary me-3"
                            onClick={() => useAppStore.getState().setActiveView('analysis')}
                        >
                            <i className="bi bi-arrow-left"></i>
                        </button>
                        <div>
                            <h1 className="h3 mb-0">
                                <i className="bi bi-diagram-3 me-2"></i>
                                Packet Details
                            </h1>
                            <p className="text-muted mb-0">ID: {selectedPacket.id}</p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="row">
                {/* Packet Information */}
                <div className="col-lg-6 mb-4">
                    <div className="card h-100">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-info-circle me-2"></i>
                                Packet Information
                            </h5>
                            
                            <div className="row g-3">
                                <div className="col-12">
                                    <label className="form-label fw-bold">Message Type</label>
                                    <div className="d-flex align-items-center">
                                        <i className={`bi ${getTypeIcon(selectedPacket.messageType)} me-2`}></i>
                                        <span className="text-capitalize badge bg-primary">
                                            {selectedPacket.messageType}
                                        </span>
                                    </div>
                                </div>
                                
                                <div className="col-md-6">
                                    <label className="form-label fw-bold">Timestamp</label>
                                    <div className="font-monospace small">
                                        {formatTimestamp(selectedPacket.timestamp)}
                                    </div>
                                </div>
                                
                                <div className="col-md-6">
                                    <label className="form-label fw-bold">Size</label>
                                    <div>
                                        {formatBytes(selectedPacket.size)}
                                        <span className="text-muted ms-2">({selectedPacket.size} bytes)</span>
                                    </div>
                                </div>
                                
                                <div className="col-md-6">
                                    <label className="form-label fw-bold">Source</label>
                                    <div className="font-monospace small">
                                        {selectedPacket.source}
                                    </div>
                                </div>
                                
                                <div className="col-md-6">
                                    <label className="form-label fw-bold">Destination</label>
                                    <div className="font-monospace small">
                                        {selectedPacket.destination}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Metadata */}
                <div className="col-lg-6 mb-4">
                    <div className="card h-100">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-tags me-2"></i>
                                Metadata
                            </h5>
                            
                            {selectedPacket.metadata && Object.keys(selectedPacket.metadata).length > 0 ? (
                                <div className="table-responsive">
                                    <table className="table table-sm">
                                        <tbody>
                                            {Object.entries(selectedPacket.metadata).map(([key, value]) => (
                                                <tr key={key}>
                                                    <td className="fw-semibold">{key}</td>
                                                    <td className="font-monospace small text-break">
                                                        {typeof value === 'object' 
                                                            ? JSON.stringify(value, null, 2)
                                                            : String(value)
                                                        }
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            ) : (
                                <div className="text-muted text-center py-3">
                                    <i className="bi bi-inbox d-block mb-2"></i>
                                    No metadata available
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Payload Hex Dump */}
            <div className="row">
                <div className="col-12">
                    <div className="card">
                        <div className="card-body">
                            <h5 className="card-title">
                                <i className="bi bi-file-binary me-2"></i>
                                Payload Hex Dump
                            </h5>
                            
                            {selectedPacket.payload && selectedPacket.payload.length > 0 ? (
                                <HexDump data={selectedPacket.payload} />
                            ) : (
                                <div className="text-muted text-center py-3">
                                    <i className="bi bi-file-x d-block mb-2"></i>
                                    No payload data available
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Navigation */}
            <div className="row mt-4">
                <div className="col-12">
                    <div className="d-flex justify-content-between">
                        <button
                            className="btn btn-outline-primary"
                            onClick={() => {
                                // Navigate to previous packet
                                const packets = currentAnalysis.packets;
                                const currentIndex = packets.findIndex(p => p.id === selectedPacket.id);
                                if (currentIndex > 0) {
                                    useAppStore.getState().selectPacket(packets[currentIndex - 1]);
                                }
                            }}
                            disabled={currentAnalysis.packets.findIndex(p => p.id === selectedPacket.id) === 0}
                        >
                            <i className="bi bi-arrow-left me-2"></i>
                            Previous Packet
                        </button>
                        
                        <button
                            className="btn btn-outline-primary"
                            onClick={() => {
                                // Navigate to next packet
                                const packets = currentAnalysis.packets;
                                const currentIndex = packets.findIndex(p => p.id === selectedPacket.id);
                                if (currentIndex < packets.length - 1) {
                                    useAppStore.getState().selectPacket(packets[currentIndex + 1]);
                                }
                            }}
                            disabled={currentAnalysis.packets.findIndex(p => p.id === selectedPacket.id) === currentAnalysis.packets.length - 1}
                        >
                            Next Packet
                            <i className="bi bi-arrow-right ms-2"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
