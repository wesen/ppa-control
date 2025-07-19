import { useMemo, useState } from 'preact/hooks';
import { useAppStore } from '../store';
import type { AnalysisResult, PacketType } from '../types';

interface PacketTimelineProps {
    analysis: AnalysisResult;
}

export function PacketTimeline({ analysis }: PacketTimelineProps) {
    const [selectedTypes, setSelectedTypes] = useState<Set<PacketType>>(new Set(['request', 'response', 'error', 'other']));
    const [zoom, setZoom] = useState(1);

    const timelineData = useMemo(() => {
        const startTime = analysis.timeRange.start.getTime();
        const endTime = analysis.timeRange.end.getTime();
        const totalDuration = endTime - startTime;

        return analysis.packets
            .filter(packet => selectedTypes.has(packet.messageType))
            .map(packet => {
                const relativeTime = packet.timestamp.getTime() - startTime;
                const position = (relativeTime / totalDuration) * 100;
                return { packet, position };
            });
    }, [analysis, selectedTypes]);

    const handleTypeToggle = (type: PacketType) => {
        const newSelected = new Set(selectedTypes);
        if (newSelected.has(type)) {
            newSelected.delete(type);
        } else {
            newSelected.add(type);
        }
        setSelectedTypes(newSelected);
    };

    const getPacketTypeClass = (type: PacketType) => {
        switch (type) {
            case 'request':
                return 'packet-type-request';
            case 'response':
                return 'packet-type-response';
            case 'error':
                return 'packet-type-error';
            default:
                return 'packet-type-other';
        }
    };

    const getTypeColor = (type: PacketType) => {
        switch (type) {
            case 'request':
                return 'var(--ppa-accent)';
            case 'response':
                return 'var(--ppa-success)';
            case 'error':
                return 'var(--ppa-danger)';
            default:
                return '#6c757d';
        }
    };

    const formatTime = (timestamp: Date) => {
        return timestamp.toLocaleTimeString([], { 
            hour12: false, 
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            fractionalSecondDigits: 3,
        });
    };

    return (
        <div>
            {/* Controls */}
            <div className="d-flex justify-content-between align-items-center mb-3">
                <div className="btn-group btn-group-sm" role="group">
                    {(['request', 'response', 'error', 'other'] as PacketType[]).map(type => (
                        <button
                            key={type}
                            type="button"
                            className={`btn ${selectedTypes.has(type) ? 'btn-primary' : 'btn-outline-secondary'}`}
                            onClick={() => handleTypeToggle(type)}
                        >
                            <span 
                                className="d-inline-block me-2"
                                style={{ 
                                    width: '12px', 
                                    height: '12px', 
                                    backgroundColor: getTypeColor(type),
                                    borderRadius: '2px'
                                }}
                            ></span>
                            {type}
                            <span className="badge bg-light text-dark ms-2">
                                {analysis.messageTypeDistribution[type] || 0}
                            </span>
                        </button>
                    ))}
                </div>

                <div className="d-flex align-items-center">
                    <label className="form-label me-2 mb-0">Zoom:</label>
                    <input
                        type="range"
                        className="form-range"
                        min="1"
                        max="10"
                        step="0.5"
                        value={zoom}
                        onChange={(e) => setZoom(parseFloat((e.target as HTMLInputElement).value))}
                        style={{ width: '100px' }}
                    />
                    <span className="ms-2">{zoom}x</span>
                </div>
            </div>

            {/* Timeline */}
            <div 
                className="packet-timeline position-relative overflow-auto"
                style={{ height: '300px' }}
            >
                {/* Time axis */}
                <div className="position-absolute top-0 start-0 w-100 border-bottom bg-light" style={{ height: '30px', zIndex: 1 }}>
                    <div className="d-flex align-items-center h-100 px-3">
                        <div className="flex-grow-1 d-flex justify-content-between small text-muted">
                            <span>{formatTime(analysis.timeRange.start)}</span>
                            <span>{formatTime(analysis.timeRange.end)}</span>
                        </div>
                    </div>
                </div>

                {/* Packet bars */}
                <div 
                    className="position-relative"
                    style={{ 
                        marginTop: '30px', 
                        height: 'calc(100% - 30px)',
                        width: `${100 * zoom}%`,
                        minWidth: '100%'
                    }}
                >
                    {timelineData.map(({ packet, position }, index) => (
                        <div
                            key={packet.id}
                            className={`timeline-packet ${getPacketTypeClass(packet.messageType)}`}
                            style={{
                                position: 'absolute',
                                left: `${position}%`,
                                top: `${(index % 10) * 25 + 10}px`,
                                width: `${Math.max(2, 100 / (analysis.totalPackets * zoom))}px`,
                                zIndex: 2
                            }}
                            title={`${packet.messageType} - ${formatTime(packet.timestamp)} - ${packet.source} → ${packet.destination}`}
                            onClick={() => {
                                // Select packet for detailed view
                                useAppStore.getState().selectPacket(packet);
                            }}
                        />
                    ))}
                </div>
            </div>

            {/* Summary */}
            <div className="mt-3 small text-muted">
                Showing {timelineData.length} of {analysis.totalPackets} packets
                {timelineData.length !== analysis.totalPackets && ' (filtered)'}
            </div>
        </div>
    );
}
