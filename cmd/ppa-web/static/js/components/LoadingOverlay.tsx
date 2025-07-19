export function LoadingOverlay() {
    return (
        <div className="loading-overlay">
            <div className="d-flex flex-column align-items-center">
                <div className="spinner-border text-primary mb-3" role="status">
                    <span className="visually-hidden">Loading...</span>
                </div>
                <div className="text-muted">Loading...</div>
            </div>
        </div>
    );
}
