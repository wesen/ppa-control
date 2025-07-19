import { Component, JSX } from 'preact';

interface ErrorBoundaryProps {
    children: JSX.Element;
}

interface ErrorBoundaryState {
    hasError: boolean;
    error?: Error;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
    constructor(props: ErrorBoundaryProps) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: any) {
        console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return (
                <div className="container-fluid p-4">
                    <div className="alert alert-danger" role="alert">
                        <h4 className="alert-heading">
                            <i className="bi bi-exclamation-triangle-fill me-2"></i>
                            Something went wrong
                        </h4>
                        <p className="mb-3">
                            An unexpected error occurred while rendering this component.
                        </p>
                        {this.state.error && (
                            <details className="mt-3">
                                <summary className="btn btn-outline-danger btn-sm">
                                    Show error details
                                </summary>
                                <pre className="mt-2 p-3 bg-light border rounded">
                                    <code>{this.state.error.stack}</code>
                                </pre>
                            </details>
                        )}
                        <hr />
                        <button
                            className="btn btn-primary"
                            onClick={() => window.location.reload()}
                        >
                            <i className="bi bi-arrow-clockwise me-2"></i>
                            Reload Page
                        </button>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}
