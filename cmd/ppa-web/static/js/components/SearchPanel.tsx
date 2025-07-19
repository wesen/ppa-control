import { useState } from 'preact/hooks';
import { useAppStore } from '../store';

export function SearchPanel() {
    const { searchQuery, searchResults, searchDocuments } = useAppStore();
    const [localQuery, setLocalQuery] = useState(searchQuery);

    const handleSearch = async (event: Event) => {
        event.preventDefault();
        await searchDocuments(localQuery);
    };

    const handleInputChange = (event: Event) => {
        const target = event.target as HTMLInputElement;
        setLocalQuery(target.value);
    };

    return (
        <div>
            <form onSubmit={handleSearch} className="mb-3">
                <div className="input-group input-group-sm">
                    <input
                        type="text"
                        className="form-control"
                        placeholder="Search documents..."
                        value={localQuery}
                        onInput={handleInputChange}
                    />
                    <button
                        className="btn btn-outline-secondary"
                        type="submit"
                        disabled={!localQuery.trim()}
                    >
                        <i className="bi bi-search"></i>
                    </button>
                </div>
            </form>

            {searchQuery && (
                <div className="mb-3">
                    <div className="small text-muted mb-2">
                        {searchResults.length} result(s) for "{searchQuery}"
                    </div>
                    
                    {searchResults.length === 0 ? (
                        <div className="text-muted text-center py-2">
                            <i className="bi bi-search d-block mb-1"></i>
                            No results found
                        </div>
                    ) : (
                        <div className="list-group list-group-flush">
                            {searchResults.map(result => (
                                <div
                                    key={result.documentId}
                                    className="list-group-item border-0 px-0 py-2"
                                    style={{ cursor: 'pointer' }}
                                    onClick={() => {
                                        // Navigate to document view
                                        useAppStore.getState().setActiveView('documents');
                                    }}
                                >
                                    <div className="fw-semibold mb-1" title={result.title}>
                                        {result.title}
                                    </div>
                                    <div className="small text-muted">
                                        {result.matches.length} match(es)
                                    </div>
                                    {result.matches.slice(0, 2).map((match, index) => (
                                        <div key={index} className="small text-truncate mt-1">
                                            <span className="text-muted">Line {match.line}:</span>{' '}
                                            <span dangerouslySetInnerHTML={{ __html: match.highlight }}></span>
                                        </div>
                                    ))}
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
