import { useState, useMemo } from 'preact/hooks';
import { useAppStore } from '../store';
import type { Packet, PacketType } from '../types';

interface PacketTableProps {
    packets: Packet[];
}

export function PacketTable({ packets }: PacketTableProps) {
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize, setPageSize] = useState(50);
    const [sortField, setSortField] = useState<keyof Packet>('timestamp');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
    const [filterType, setFilterType] = useState<PacketType | 'all'>('all');
    const [searchTerm, setSearchTerm] = useState('');

    const selectPacket = useAppStore(state => state.selectPacket);

    const filteredAndSortedPackets = useMemo(() => {
        let filtered = packets;

        // Apply type filter
        if (filterType !== 'all') {
            filtered = filtered.filter(packet => packet.messageType === filterType);
        }

        // Apply search filter
        if (searchTerm) {
            const term = searchTerm.toLowerCase();
            filtered = filtered.filter(packet =>
                packet.source.toLowerCase().includes(term) ||
                packet.destination.toLowerCase().includes(term) ||
                packet.messageType.toLowerCase().includes(term) ||
                packet.id.toLowerCase().includes(term)
            );
        }

        // Apply sorting
        filtered.sort((a, b) => {
            let aValue = a[sortField];
            let bValue = b[sortField];

            // Handle different data types
            if (aValue instanceof Date && bValue instanceof Date) {
                aValue = aValue.getTime();
                bValue = bValue.getTime();
            } else if (typeof aValue === 'string' && typeof bValue === 'string') {
                aValue = aValue.toLowerCase();
                bValue = bValue.toLowerCase();
            }

            if (aValue !== undefined && bValue !== undefined) {
                if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1;
                if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1;
            }
            return 0;
        });

        return filtered;
    }, [packets, filterType, searchTerm, sortField, sortDirection]);

    const paginatedPackets = useMemo(() => {
        const startIndex = (currentPage - 1) * pageSize;
        return filteredAndSortedPackets.slice(startIndex, startIndex + pageSize);
    }, [filteredAndSortedPackets, currentPage, pageSize]);

    const totalPages = Math.ceil(filteredAndSortedPackets.length / pageSize);

    const handleSort = (field: keyof Packet) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
        setCurrentPage(1);
    };

    const handleRowClick = (packet: Packet) => {
        selectPacket(packet);
        // Could also switch to a packet detail view
        useAppStore.getState().setActiveView('packets');
    };

    const getSortIcon = (field: keyof Packet) => {
        if (sortField !== field) return 'bi-arrow-down-up';
        return sortDirection === 'asc' ? 'bi-arrow-up' : 'bi-arrow-down';
    };

    const getTypeIcon = (type: PacketType) => {
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
            second: '2-digit',
            fractionalSecondDigits: 3,
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
        <div>
            {/* Filters and Controls */}
            <div className="row mb-3">
                <div className="col-md-6">
                    <div className="d-flex gap-2">
                        <select
                            className="form-select form-select-sm"
                            value={filterType}
                            onChange={(e) => {
                                setFilterType((e.target as HTMLSelectElement).value as PacketType | 'all');
                                setCurrentPage(1);
                            }}
                            style={{ width: 'auto' }}
                        >
                            <option value="all">All Types</option>
                            <option value="request">Request</option>
                            <option value="response">Response</option>
                            <option value="error">Error</option>
                            <option value="other">Other</option>
                        </select>

                        <select
                            className="form-select form-select-sm"
                            value={pageSize}
                            onChange={(e) => {
                                setPageSize(parseInt((e.target as HTMLSelectElement).value));
                                setCurrentPage(1);
                            }}
                            style={{ width: 'auto' }}
                        >
                            <option value={25}>25 per page</option>
                            <option value={50}>50 per page</option>
                            <option value={100}>100 per page</option>
                            <option value={200}>200 per page</option>
                        </select>
                    </div>
                </div>
                <div className="col-md-6">
                    <div className="input-group input-group-sm">
                        <input
                            type="text"
                            className="form-control"
                            placeholder="Search packets..."
                            value={searchTerm}
                            onInput={(e) => {
                                setSearchTerm((e.target as HTMLInputElement).value);
                                setCurrentPage(1);
                            }}
                        />
                        <span className="input-group-text">
                            <i className="bi bi-search"></i>
                        </span>
                    </div>
                </div>
            </div>

            {/* Results info */}
            <div className="d-flex justify-content-between align-items-center mb-3">
                <div className="text-muted">
                    Showing {paginatedPackets.length} of {filteredAndSortedPackets.length} packets
                    {filteredAndSortedPackets.length !== packets.length && ` (filtered from ${packets.length})`}
                </div>
                
                {/* Pagination */}
                {totalPages > 1 && (
                    <nav>
                        <ul className="pagination pagination-sm mb-0">
                            <li className={`page-item ${currentPage === 1 ? 'disabled' : ''}`}>
                                <button
                                    className="page-link"
                                    onClick={() => setCurrentPage(currentPage - 1)}
                                    disabled={currentPage === 1}
                                >
                                    Previous
                                </button>
                            </li>
                            
                            {[...Array(totalPages)].map((_, i) => {
                                const page = i + 1;
                                const isNearCurrent = Math.abs(page - currentPage) <= 2;
                                const isFirst = page === 1;
                                const isLast = page === totalPages;
                                
                                if (!isNearCurrent && !isFirst && !isLast) {
                                    return null;
                                }
                                
                                return (
                                    <li key={page} className={`page-item ${currentPage === page ? 'active' : ''}`}>
                                        <button
                                            className="page-link"
                                            onClick={() => setCurrentPage(page)}
                                        >
                                            {page}
                                        </button>
                                    </li>
                                );
                            })}
                            
                            <li className={`page-item ${currentPage === totalPages ? 'disabled' : ''}`}>
                                <button
                                    className="page-link"
                                    onClick={() => setCurrentPage(currentPage + 1)}
                                    disabled={currentPage === totalPages}
                                >
                                    Next
                                </button>
                            </li>
                        </ul>
                    </nav>
                )}
            </div>

            {/* Table */}
            <div className="table-responsive">
                <table className="table table-hover table-sm">
                    <thead className="table-light">
                        <tr>
                            <th 
                                scope="col" 
                                style={{ cursor: 'pointer' }}
                                onClick={() => handleSort('timestamp')}
                            >
                                Timestamp <i className={`bi ${getSortIcon('timestamp')}`}></i>
                            </th>
                            <th scope="col">Type</th>
                            <th 
                                scope="col"
                                style={{ cursor: 'pointer' }}
                                onClick={() => handleSort('source')}
                            >
                                Source <i className={`bi ${getSortIcon('source')}`}></i>
                            </th>
                            <th 
                                scope="col"
                                style={{ cursor: 'pointer' }}
                                onClick={() => handleSort('destination')}
                            >
                                Destination <i className={`bi ${getSortIcon('destination')}`}></i>
                            </th>
                            <th 
                                scope="col"
                                style={{ cursor: 'pointer' }}
                                onClick={() => handleSort('size')}
                            >
                                Size <i className={`bi ${getSortIcon('size')}`}></i>
                            </th>
                            <th scope="col">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {paginatedPackets.map(packet => (
                            <tr 
                                key={packet.id}
                                style={{ cursor: 'pointer' }}
                                onClick={() => handleRowClick(packet)}
                            >
                                <td className="font-monospace small">
                                    {formatTimestamp(packet.timestamp)}
                                </td>
                                <td>
                                    <span className="d-flex align-items-center">
                                        <i className={`bi ${getTypeIcon(packet.messageType)} me-2`}></i>
                                        <span className="text-capitalize">{packet.messageType}</span>
                                    </span>
                                </td>
                                <td className="font-monospace small">{packet.source}</td>
                                <td className="font-monospace small">{packet.destination}</td>
                                <td>{formatBytes(packet.size)}</td>
                                <td>
                                    <button
                                        className="btn btn-outline-primary btn-sm"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            handleRowClick(packet);
                                        }}
                                    >
                                        <i className="bi bi-eye"></i>
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {paginatedPackets.length === 0 && (
                <div className="text-center py-4 text-muted">
                    <i className="bi bi-inbox d-block mb-2" style={{ fontSize: '2rem' }}></i>
                    No packets found
                </div>
            )}
        </div>
    );
}
