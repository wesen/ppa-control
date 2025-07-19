import { useState, useMemo } from 'preact/hooks';

interface HexDumpProps {
    data: Uint8Array;
    bytesPerLine?: number;
    showAscii?: boolean;
    highlightBytes?: number[];
}

export function HexDump({ 
    data, 
    bytesPerLine = 16, 
    showAscii = true,
    highlightBytes = []
}: HexDumpProps) {
    const [searchTerm, setSearchTerm] = useState('');
    const [searchMode, setSearchMode] = useState<'hex' | 'ascii'>('hex');

    const lines = useMemo(() => {
        const result = [];
        for (let i = 0; i < data.length; i += bytesPerLine) {
            const lineData = data.slice(i, i + bytesPerLine);
            const offset = i;
            
            // Convert to hex
            const hexBytes = Array.from(lineData).map(byte => 
                byte.toString(16).padStart(2, '0').toUpperCase()
            );
            
            // Convert to ASCII
            const asciiChars = Array.from(lineData).map(byte => 
                (byte >= 32 && byte <= 126) ? String.fromCharCode(byte) : '.'
            );
            
            result.push({
                offset,
                hexBytes,
                asciiChars,
                rawBytes: lineData
            });
        }
        return result;
    }, [data, bytesPerLine]);

    const searchResults = useMemo(() => {
        if (!searchTerm) return [];
        
        const results = [];
        const term = searchTerm.toLowerCase();
        
        if (searchMode === 'hex') {
            // Search in hex representation
            const hexString = Array.from(data).map(b => b.toString(16).padStart(2, '0')).join('');
            const cleanTerm = term.replace(/[^0-9a-f]/g, '');
            let index = 0;
            
            while ((index = hexString.indexOf(cleanTerm, index)) !== -1) {
                const byteIndex = Math.floor(index / 2);
                const length = Math.ceil(cleanTerm.length / 2);
                results.push({ start: byteIndex, length });
                index += 2;
            }
        } else {
            // Search in ASCII representation
            const asciiString = Array.from(data).map(byte => 
                (byte >= 32 && byte <= 126) ? String.fromCharCode(byte) : '.'
            ).join('');
            let index = 0;
            
            while ((index = asciiString.indexOf(term, index)) !== -1) {
                results.push({ start: index, length: term.length });
                index++;
            }
        }
        
        return results;
    }, [data, searchTerm, searchMode]);

    const isHighlighted = (byteIndex: number) => {
        return highlightBytes.includes(byteIndex) || 
               searchResults.some(result => 
                   byteIndex >= result.start && byteIndex < result.start + result.length
               );
    };

    const formatOffset = (offset: number) => {
        return offset.toString(16).padStart(8, '0').toUpperCase();
    };

    const handleExport = () => {
        const content = lines.map(line => {
            const offsetStr = formatOffset(line.offset);
            const hexStr = line.hexBytes.join(' ').padEnd(bytesPerLine * 3 - 1, ' ');
            const asciiStr = line.asciiChars.join('');
            return `${offsetStr}  ${hexStr}  |${asciiStr}|`;
        }).join('\n');
        
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'hexdump.txt';
        a.click();
        URL.revokeObjectURL(url);
    };

    return (
        <div>
            {/* Controls */}
            <div className="row mb-3">
                <div className="col-md-8">
                    <div className="input-group input-group-sm">
                        <select
                            className="form-select"
                            value={searchMode}
                            onChange={(e) => setSearchMode((e.target as HTMLSelectElement).value as 'hex' | 'ascii')}
                            style={{ maxWidth: '100px' }}
                        >
                            <option value="hex">Hex</option>
                            <option value="ascii">ASCII</option>
                        </select>
                        <input
                            type="text"
                            className="form-control"
                            placeholder={`Search ${searchMode}...`}
                            value={searchTerm}
                            onInput={(e) => setSearchTerm((e.target as HTMLInputElement).value)}
                        />
                        <button 
                            className="btn btn-outline-secondary"
                            onClick={() => setSearchTerm('')}
                            disabled={!searchTerm}
                        >
                            <i className="bi bi-x"></i>
                        </button>
                    </div>
                    {searchResults.length > 0 && (
                        <div className="small text-muted mt-1">
                            Found {searchResults.length} match(es)
                        </div>
                    )}
                </div>
                <div className="col-md-4 text-end">
                    <button
                        className="btn btn-outline-secondary btn-sm me-2"
                        onClick={handleExport}
                    >
                        <i className="bi bi-download me-1"></i>
                        Export
                    </button>
                    <span className="small text-muted">
                        {data.length} bytes
                    </span>
                </div>
            </div>

            {/* Hex Dump */}
            <div className="hex-dump">
                {lines.map((line, lineIndex) => (
                    <div key={lineIndex} className="d-flex font-monospace">
                        {/* Offset */}
                        <span className="hex-offset">
                            {formatOffset(line.offset)}
                        </span>
                        
                        {/* Hex bytes */}
                        <span className="hex-bytes flex-shrink-0">
                            {line.hexBytes.map((byte, byteIndex) => {
                                const globalIndex = line.offset + byteIndex;
                                const highlighted = isHighlighted(globalIndex);
                                return (
                                    <span
                                        key={byteIndex}
                                        className={highlighted ? 'search-highlight' : ''}
                                        style={{ marginRight: '4px' }}
                                    >
                                        {byte}
                                    </span>
                                );
                            })}
                            {/* Pad remaining space if line is short */}
                            {Array(bytesPerLine - line.hexBytes.length).fill(0).map((_, i) => (
                                <span key={`pad-${i}`} style={{ marginRight: '4px' }}>{'   '}</span>
                            ))}
                        </span>
                        
                        {/* ASCII representation */}
                        {showAscii && (
                            <>
                                <span className="mx-2">|</span>
                                <span className="hex-ascii">
                                    {line.asciiChars.map((char, charIndex) => {
                                        const globalIndex = line.offset + charIndex;
                                        const highlighted = isHighlighted(globalIndex);
                                        return (
                                            <span
                                                key={charIndex}
                                                className={highlighted ? 'search-highlight' : ''}
                                            >
                                                {char}
                                            </span>
                                        );
                                    })}
                                    {/* Pad remaining space if line is short */}
                                    {Array(bytesPerLine - line.asciiChars.length).fill('.').map((char, i) => (
                                        <span key={`ascii-pad-${i}`}>{char}</span>
                                    ))}
                                </span>
                                <span>|</span>
                            </>
                        )}
                    </div>
                ))}
            </div>

            {data.length === 0 && (
                <div className="text-center py-4 text-muted">
                    <i className="bi bi-file-x d-block mb-2" style={{ fontSize: '2rem' }}></i>
                    No data to display
                </div>
            )}
        </div>
    );
}
