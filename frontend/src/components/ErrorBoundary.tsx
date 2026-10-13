/**
 * Error Boundary — Dell 1996 retro style
 */

import { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.setState({ error, errorInfo });
  }

  handleReset = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="min-h-screen bg-muted flex items-center justify-center px-4">
          <div className="w-full max-w-md rounded-xl border border-border bg-background p-8 shadow-card">
            <div className="flex items-center gap-3">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-destructive/10 text-destructive">
                <AlertTriangle className="w-5 h-5" />
              </span>
              <h1 className="text-xl font-medium tracking-tight text-foreground">
                Something went wrong
              </h1>
            </div>

            <p className="mt-4 text-sm text-muted-foreground mb-6">
              Something unexpected happened. Please try refreshing the page.
            </p>

            {process.env['NODE_ENV'] === 'development' && this.state.error && (
              <div className="mb-6 rounded-md border border-destructive/30 bg-destructive/5 p-3">
                <p className="text-xs font-medium text-destructive mb-1">
                  Error: {this.state.error.message}
                </p>
                {this.state.errorInfo && (
                  <pre className="text-[11px] text-destructive/80 overflow-auto max-h-32">
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            <div className="flex gap-2">
              <button
                onClick={this.handleReset}
                className="inline-flex h-10 items-center justify-center rounded-full bg-primary px-5 text-sm font-medium text-primary-foreground hover:bg-primary/85 transition-colors duration-fast"
              >
                Try Again
              </button>
              <button
                onClick={() => window.location.href = '/'}
                className="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-5 text-sm font-medium text-foreground hover:bg-muted transition-colors duration-fast"
              >
                Go Home
              </button>
            </div>

            <p className="mt-6 text-xs text-muted-foreground">
              If this problem persists, please{' '}
              <a href="mailto:support@keyles.com" className="text-foreground underline underline-offset-4 hover:text-muted-foreground">
                contact support
              </a>
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
