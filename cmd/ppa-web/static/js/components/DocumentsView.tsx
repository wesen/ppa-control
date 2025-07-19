import { useState, useEffect } from 'preact/hooks';
import { useAppStore } from '../store';
import { MarkdownRenderer } from './MarkdownRenderer';

export function DocumentsView() {
    const { documents, searchResults, searchQuery } = useAppStore();
    const [selectedDocument, setSelectedDocument] = useState<string | null>(null);
    const [viewMode, setViewMode] = useState<'list' | 'preview'>('list');

    const currentDoc = selectedDocument ? documents.find(d => d.id === selectedDocument) : null;

    useEffect(() => {
        // Load documents on mount
        const loadDocuments = async () => {
            try {
                const response = await fetch('/api/docs/list');
                if (response.ok) {
                    const backendDocs = await response.json();
                    // Transform backend response to match frontend Document interface
                    const docs = backendDocs.map((doc: any) => ({
                        id: doc.relativePath,
                        title: doc.title || doc.name,
                        content: doc.summary || '',
                        type: doc.type === '.md' ? 'markdown' : 'text',
                        createdAt: new Date(doc.modTime),
                        updatedAt: new Date(doc.modTime),
                        tags: []
                    }));
                    useAppStore.getState().setDocuments(docs);
                } else {
                    console.error('Failed to load documents:', response.statusText);
                }
            } catch (error) {
                console.error('Failed to load documents:', error);
            }
        };
        
        loadDocuments();
    }, []);

    const filteredDocuments = searchQuery && searchResults.length > 0
        ? documents.filter(doc => searchResults.some(result => result.document?.relativePath === doc.id))
        : documents;

    const handleDocumentSelect = async (documentId: string) => {
        setSelectedDocument(documentId);
        setViewMode('preview');
        
        // Load full document content
        try {
            const response = await fetch(`/api/docs/view?path=${encodeURIComponent(documentId)}`);
            if (response.ok) {
                const docData = await response.json();
                // Update the document in the store with full content
                const currentDocs = useAppStore.getState().documents;
                const updatedDocs = currentDocs.map(doc => 
                    doc.id === documentId 
                        ? { ...doc, content: docData.content }
                        : doc
                );
                useAppStore.getState().setDocuments(updatedDocs);
            }
        } catch (error) {
            console.error('Failed to load document content:', error);
        }
    };

    const getDocumentIcon = (type: string) => {
        switch (type) {
            case 'markdown':
                return 'bi-markdown text-primary';
            case 'analysis':
                return 'bi-bar-chart text-success';
            default:
                return 'bi-file-text text-muted';
        }
    };

    const formatDate = (date: Date) => {
        return date.toLocaleDateString([], {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    if (viewMode === 'preview' && currentDoc) {
        return (
            <div className="container-fluid p-4">
                <div className="row mb-4">
                    <div className="col">
                        <div className="d-flex align-items-center">
                            <button
                                className="btn btn-outline-secondary me-3"
                                onClick={() => setViewMode('list')}
                            >
                                <i className="bi bi-arrow-left"></i>
                            </button>
                            <div className="flex-grow-1">
                                <h1 className="h3 mb-0">
                                    <i className={`bi ${getDocumentIcon(currentDoc.type)} me-2`}></i>
                                    {currentDoc.title}
                                </h1>
                                <div className="text-muted">
                                    <span className="badge bg-secondary me-2">{currentDoc.type}</span>
                                    Last updated: {formatDate(currentDoc.updatedAt)}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Document Tags */}
                {currentDoc.tags.length > 0 && (
                    <div className="row mb-3">
                        <div className="col">
                            <div className="d-flex flex-wrap gap-2">
                                {currentDoc.tags.map(tag => (
                                    <span key={tag} className="badge bg-light text-dark">
                                        <i className="bi bi-tag me-1"></i>
                                        {tag}
                                    </span>
                                ))}
                            </div>
                        </div>
                    </div>
                )}

                {/* Document Content */}
                <div className="row">
                    <div className="col">
                        <div className="card">
                            <div className="card-body">
                                {currentDoc.type === 'markdown' ? (
                                    <MarkdownRenderer content={currentDoc.content} />
                                ) : (
                                    <pre className="pre-wrap">{currentDoc.content}</pre>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="container-fluid p-4">
            <div className="row mb-4">
                <div className="col">
                    <h1 className="h3 mb-0">
                        <i className="bi bi-file-text me-2"></i>
                        Documents
                    </h1>
                    <p className="text-muted">
                        Analysis reports, documentation, and markdown files
                    </p>
                </div>
            </div>

            {/* Search Results Summary */}
            {searchQuery && (
                <div className="row mb-4">
                    <div className="col">
                        <div className="alert alert-info">
                            <i className="bi bi-search me-2"></i>
                            <strong>{searchResults.length}</strong> document(s) found for "{searchQuery}"
                            <button
                                className="btn btn-sm btn-outline-info ms-3"
                                onClick={() => {
                                    useAppStore.getState().setSearchQuery('');
                                    useAppStore.getState().setSearchResults([]);
                                }}
                            >
                                Clear search
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Document List */}
            <div className="row">
                <div className="col">
                    {filteredDocuments.length === 0 ? (
                        <div className="text-center py-5">
                            <i className="bi bi-folder-x display-1 text-muted mb-3"></i>
                            <h4 className="text-muted">
                                {searchQuery ? 'No documents match your search' : 'No documents available'}
                            </h4>
                            <p className="text-muted">
                                {searchQuery 
                                    ? 'Try adjusting your search terms or clear the search to see all documents.'
                                    : 'Documents will appear here as you perform analyses and generate reports.'
                                }
                            </p>
                        </div>
                    ) : (
                        <div className="row">
                            {filteredDocuments.map(document => {
                                const searchResult = searchResults.find(r => r.document?.relativePath === document.id);
                                return (
                                    <div key={document.id} className="col-md-6 col-lg-4 mb-4">
                                        <div 
                                            className="card h-100"
                                            style={{ cursor: 'pointer' }}
                                            onClick={() => handleDocumentSelect(document.id)}
                                        >
                                            <div className="card-body">
                                                <div className="d-flex align-items-start mb-3">
                                                    <i className={`bi ${getDocumentIcon(document.type)} me-3 mt-1`}></i>
                                                    <div className="flex-grow-1 min-w-0">
                                                        <h5 className="card-title text-truncate" title={document.title}>
                                                            {document.title}
                                                        </h5>
                                                        <div className="small text-muted mb-2">
                                                            <span className="badge bg-secondary me-2">
                                                                {document.type}
                                                            </span>
                                                            {formatDate(document.updatedAt)}
                                                        </div>
                                                    </div>
                                                </div>

                                                {/* Search matches */}
                                                {searchResult && searchResult.matches.length > 0 && (
                                                    <div className="mb-3">
                                                        <div className="small fw-bold text-primary mb-1">
                                                            {searchResult.matches.length} match(es):
                                                        </div>
                                                        {searchResult.matches.slice(0, 2).map((match, index) => (
                                                            <div key={index} className="small text-muted mb-1">
                                                                <span className="text-decoration-underline">Line {match.line}:</span>{' '}
                                                                <span 
                                                                    dangerouslySetInnerHTML={{ __html: match.highlight }}
                                                                    className="text-truncate d-inline-block"
                                                                    style={{ maxWidth: '200px' }}
                                                                />
                                                            </div>
                                                        ))}
                                                    </div>
                                                )}

                                                {/* Document preview */}
                                                <p className="card-text small text-muted">
                                                    {document.content.substring(0, 150)}
                                                    {document.content.length > 150 && '...'}
                                                </p>

                                                {/* Tags */}
                                                {document.tags.length > 0 && (
                                                    <div className="d-flex flex-wrap gap-1 mt-2">
                                                        {document.tags.slice(0, 3).map(tag => (
                                                            <span key={tag} className="badge bg-light text-dark small">
                                                                {tag}
                                                            </span>
                                                        ))}
                                                        {document.tags.length > 3 && (
                                                            <span className="badge bg-light text-muted small">
                                                                +{document.tags.length - 3}
                                                            </span>
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                            
                                            <div className="card-footer bg-transparent">
                                                <div className="d-flex justify-content-between align-items-center">
                                                    <small className="text-muted">
                                                        {document.content.length} characters
                                                    </small>
                                                    <button className="btn btn-sm btn-outline-primary">
                                                        <i className="bi bi-eye me-1"></i>
                                                        View
                                                    </button>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
