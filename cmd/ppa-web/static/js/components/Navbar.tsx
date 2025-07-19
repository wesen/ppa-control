import { useAppStore } from '../store';

export function Navbar() {
    const { sidebarVisible, setSidebarVisible, activeView, setActiveView } = useAppStore();

    return (
        <nav className="navbar navbar-expand-lg navbar-dark bg-dark">
            <div className="container-fluid">
                <button
                    className="btn btn-outline-light me-3"
                    type="button"
                    onClick={() => setSidebarVisible(!sidebarVisible)}
                    aria-label="Toggle sidebar"
                >
                    <i className="bi bi-list"></i>
                </button>
                
                <a className="navbar-brand" href="/">
                    <i className="bi bi-graph-up-arrow me-2"></i>
                    PPA Packet Analysis
                </a>
                
                <div className="navbar-nav ms-auto">
                    <div className="nav-item dropdown">
                        <button
                            className="btn btn-outline-light dropdown-toggle"
                            type="button"
                            data-bs-toggle="dropdown"
                            aria-expanded="false"
                        >
                            <i className="bi bi-grid-3x3-gap me-1"></i>
                            View
                        </button>
                        <ul className="dropdown-menu dropdown-menu-end">
                            <li>
                                <button
                                    className={`dropdown-item ${activeView === 'dashboard' ? 'active' : ''}`}
                                    onClick={() => setActiveView('dashboard')}
                                >
                                    <i className="bi bi-house me-2"></i>
                                    Dashboard
                                </button>
                            </li>
                            <li>
                                <button
                                    className={`dropdown-item ${activeView === 'analysis' ? 'active' : ''}`}
                                    onClick={() => setActiveView('analysis')}
                                >
                                    <i className="bi bi-bar-chart me-2"></i>
                                    Analysis
                                </button>
                            </li>
                            <li>
                                <button
                                    className={`dropdown-item ${activeView === 'packets' ? 'active' : ''}`}
                                    onClick={() => setActiveView('packets')}
                                >
                                    <i className="bi bi-diagram-3 me-2"></i>
                                    Packets
                                </button>
                            </li>
                            <li>
                                <button
                                    className={`dropdown-item ${activeView === 'documents' ? 'active' : ''}`}
                                    onClick={() => setActiveView('documents')}
                                >
                                    <i className="bi bi-file-text me-2"></i>
                                    Documents
                                </button>
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        </nav>
    );
}
