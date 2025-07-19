import { useRef } from 'preact/hooks';
import { useAppStore } from '../store';
import type { PCAPFile } from '../types';

export function PCAPFileList() {
    const { pcapFiles, selectedPcapFile, selectPcapFile, uploadPcapFile, startAnalysis } = useAppStore();
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleFileSelect = () => {
        fileInputRef.current?.click();
    };

    const handleFileChange = async (event: Event) => {
        const target = event.target as HTMLInputElement;
        const file = target.files?.[0];
        if (file) {
            await uploadPcapFile(file);
            target.value = ''; // Reset input
        }
    };

    const handleFileClick = (file: PCAPFile) => {
        selectPcapFile(file);
        useAppStore.getState().setActiveView('analysis');
    };

    const handleAnalyzeClick = async (event: Event, file: PCAPFile) => {
        event.stopPropagation();
        await startAnalysis(file.id);
    };

    const getStatusIcon = (status: PCAPFile['status']) => {
        switch (status) {
            case 'uploaded':
                return 'bi-file-earmark';
            case 'analyzing':
                return 'bi-hourglass-split';
            case 'analyzed':
                return 'bi-check-circle-fill text-success';
            case 'error':
                return 'bi-exclamation-triangle-fill text-danger';
            default:
                return 'bi-file-earmark';
        }
    };

    const formatFileSize = (bytes: number) => {
        const units = ['B', 'KB', 'MB', 'GB'];
        let size = bytes;
        let unitIndex = 0;
        
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        
        return `${size.toFixed(1)} ${units[unitIndex]}`;
    };

    return (
        <div>
            <button
                className="btn btn-primary btn-sm w-100 mb-3"
                onClick={handleFileSelect}
            >
                <i className="bi bi-upload me-2"></i>
                Upload PCAP
            </button>
            
            <input
                ref={fileInputRef}
                type="file"
                accept=".pcap,.pcapng,.cap"
                style={{ display: 'none' }}
                onChange={handleFileChange}
            />

            <div className="list-group list-group-flush">
                {pcapFiles.length === 0 ? (
                    <div className="text-muted text-center py-3">
                        <i className="bi bi-inbox d-block mb-2" style={{ fontSize: '2rem' }}></i>
                        No PCAP files uploaded
                    </div>
                ) : (
                    pcapFiles.map(file => (
                        <div
                            key={file.id}
                            className={`list-group-item list-group-item-action ${
                                selectedPcapFile?.id === file.id ? 'active' : ''
                            }`}
                            style={{ cursor: 'pointer', border: 'none', padding: '0.75rem 0' }}
                            onClick={() => handleFileClick(file)}
                        >
                            <div className="d-flex align-items-start">
                                <i className={`bi ${getStatusIcon(file.status)} me-2 mt-1`}></i>
                                <div className="flex-grow-1 min-w-0">
                                    <div className="fw-semibold text-truncate" title={file.name}>
                                        {file.name}
                                    </div>
                                    <div className="small text-muted">
                                        {formatFileSize(file.size)}
                                        <span className="mx-1">•</span>
                                        {file.uploadDate.toLocaleDateString()}
                                    </div>
                                    {file.status === 'analyzing' && file.analysisProgress !== undefined && (
                                        <div className="progress mt-2" style={{ height: '4px' }}>
                                            <div
                                                className="progress-bar progress-bar-striped progress-bar-animated"
                                                style={{ width: `${file.analysisProgress}%` }}
                                            ></div>
                                        </div>
                                    )}
                                </div>
                                
                                {file.status === 'uploaded' && (
                                    <button
                                        className="btn btn-outline-primary btn-sm ms-2"
                                        onClick={(e) => handleAnalyzeClick(e, file)}
                                        title="Start Analysis"
                                    >
                                        <i className="bi bi-play-fill"></i>
                                    </button>
                                )}
                                
                                {file.status === 'analyzing' && (
                                    <div className="spinner-border spinner-border-sm text-primary ms-2" role="status">
                                        <span className="visually-hidden">Analyzing...</span>
                                    </div>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
