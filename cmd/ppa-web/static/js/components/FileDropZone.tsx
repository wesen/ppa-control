import { useState } from 'preact/hooks';
import { useAppStore } from '../store';

export function FileDropZone() {
    const [isDragOver, setIsDragOver] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const uploadPcapFile = useAppStore(state => state.uploadPcapFile);

    const handleDragOver = (event: DragEvent) => {
        event.preventDefault();
        setIsDragOver(true);
    };

    const handleDragLeave = (event: DragEvent) => {
        event.preventDefault();
        setIsDragOver(false);
    };

    const handleDrop = async (event: DragEvent) => {
        event.preventDefault();
        setIsDragOver(false);

        const files = event.dataTransfer?.files;
        if (files && files.length > 0) {
            const file = files[0];
            if (isValidPcapFile(file)) {
                await handleFileUpload(file);
            } else {
                alert('Please upload a valid PCAP file (.pcap, .pcapng, .cap)');
            }
        }
    };

    const handleClick = () => {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.pcap,.pcapng,.cap';
        input.onchange = async (event) => {
            const target = event.target as HTMLInputElement;
            const file = target.files?.[0];
            if (file) {
                await handleFileUpload(file);
            }
        };
        input.click();
    };

    const handleFileUpload = async (file: File) => {
        setIsUploading(true);
        try {
            await uploadPcapFile(file);
        } finally {
            setIsUploading(false);
        }
    };

    const isValidPcapFile = (file: File): boolean => {
        const validExtensions = ['.pcap', '.pcapng', '.cap'];
        return validExtensions.some(ext => file.name.toLowerCase().endsWith(ext));
    };

    return (
        <div
            className={`file-drop-zone ${isDragOver ? 'dragover' : ''}`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={handleClick}
        >
            {isUploading ? (
                <div className="d-flex flex-column align-items-center">
                    <div className="spinner-border text-primary mb-3" role="status">
                        <span className="visually-hidden">Uploading...</span>
                    </div>
                    <div className="h5 mb-2">Uploading file...</div>
                    <div className="text-muted">Please wait while your file is being uploaded.</div>
                </div>
            ) : (
                <div className="d-flex flex-column align-items-center">
                    <i className="bi bi-cloud-upload display-4 text-muted mb-3"></i>
                    <div className="h5 mb-2">Drop PCAP file here</div>
                    <div className="text-muted mb-3">or click to browse</div>
                    <div className="small text-muted">
                        Supported formats: .pcap, .pcapng, .cap
                    </div>
                </div>
            )}
        </div>
    );
}
