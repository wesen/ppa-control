import { JSX } from 'preact';
import { useAppStore } from '../store';
import { Navbar } from './Navbar';
import { Sidebar } from './Sidebar';
import { ErrorBoundary } from './ErrorBoundary';
import { LoadingOverlay } from './LoadingOverlay';

interface LayoutProps {
    children: JSX.Element;
}

export function Layout({ children }: LayoutProps) {
    const sidebarVisible = useAppStore(state => state.sidebarVisible);
    const loading = useAppStore(state => state.loading);
    const error = useAppStore(state => state.error);

    return (
        <div className="d-flex flex-column min-vh-100">
            <Navbar />
            
            {error && (
                <div className="alert alert-danger alert-dismissible m-3" role="alert">
                    <i className="bi bi-exclamation-triangle-fill me-2"></i>
                    <strong>Error:</strong> {error.message}
                    {error.code && <small className="d-block text-muted mt-1">Code: {error.code}</small>}
                    <button 
                        type="button" 
                        className="btn-close" 
                        onClick={() => useAppStore.getState().setError(undefined)}
                        aria-label="Close"
                    ></button>
                </div>
            )}
            
            <div className="flex-grow-1 d-flex">
                {sidebarVisible && (
                    <div className="col-md-3 col-lg-2 sidebar">
                        <Sidebar />
                    </div>
                )}
                
                <div className={`${sidebarVisible ? 'col-md-9 col-lg-10' : 'col-12'} main-content position-relative`}>
                    <ErrorBoundary>
                        {children}
                    </ErrorBoundary>
                    {loading && <LoadingOverlay />}
                </div>
            </div>
        </div>
    );
}
