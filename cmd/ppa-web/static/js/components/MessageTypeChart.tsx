import type { PacketType } from '../types';

interface MessageTypeChartProps {
    distribution: Record<PacketType, number>;
}

export function MessageTypeChart({ distribution }: MessageTypeChartProps) {
    const total = Object.values(distribution).reduce((sum, count) => sum + count, 0);
    
    const data = Object.entries(distribution).map(([type, count]) => ({
        type: type as PacketType,
        count,
        percentage: total > 0 ? (count / total) * 100 : 0
    }));

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

    const getTypeIcon = (type: PacketType) => {
        switch (type) {
            case 'request':
                return 'bi-arrow-right-circle';
            case 'response':
                return 'bi-arrow-left-circle';
            case 'error':
                return 'bi-exclamation-triangle';
            default:
                return 'bi-circle';
        }
    };

    if (total === 0) {
        return (
            <div className="text-center py-4 text-muted">
                <i className="bi bi-pie-chart d-block mb-2" style={{ fontSize: '2rem' }}></i>
                No packets to display
            </div>
        );
    }

    return (
        <div>
            {/* Simple bar chart */}
            <div className="mb-3">
                {data.map(({ type, count, percentage }) => (
                    <div key={type} className="mb-3">
                        <div className="d-flex justify-content-between align-items-center mb-1">
                            <div className="d-flex align-items-center">
                                <i className={`bi ${getTypeIcon(type)} me-2`} style={{ color: getTypeColor(type) }}></i>
                                <span className="text-capitalize fw-semibold">{type}</span>
                            </div>
                            <div className="text-end">
                                <div className="fw-bold">{count.toLocaleString()}</div>
                                <div className="small text-muted">{percentage.toFixed(1)}%</div>
                            </div>
                        </div>
                        <div className="progress" style={{ height: '8px' }}>
                            <div
                                className="progress-bar"
                                style={{ 
                                    width: `${percentage}%`,
                                    backgroundColor: getTypeColor(type)
                                }}
                            />
                        </div>
                    </div>
                ))}
            </div>

            {/* Summary */}
            <div className="border-top pt-3">
                <div className="d-flex justify-content-between">
                    <span className="text-muted">Total Packets:</span>
                    <span className="fw-bold">{total.toLocaleString()}</span>
                </div>
            </div>
        </div>
    );
}
