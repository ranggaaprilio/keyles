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
        <div className="min-h-screen bg-white flex items-center justify-center px-4">
          <div className="w-full max-w-md">
            {/* Section eyebrow — salmon */}
            <div className="bg-[#d77a7a] px-4 py-4">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-5 h-5 text-black" />
                <h1 className="font-['Arial_Black','Helvetica',system-ui,sans-serif] text-[24px] font-black uppercase leading-[1.0] text-black">
                  SOMETHING WENT WRONG
                </h1>
              </div>
            </div>

            <div className="border-x border-b border-black bg-white p-4">
              <p className="font-['Times_New_Roman',Times,serif] text-sm text-black mb-4">
                Something unexpected happened. Please try refreshing the page.
              </p>

              {process.env['NODE_ENV'] === 'development' && this.state.error && (
                <div className="mb-4 p-3 border border-red-700 bg-red-50">
                  <p className="font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold text-red-800 mb-1">
                    Error: {this.state.error.message}
                  </p>
                  {this.state.errorInfo && (
                    <pre className="font-['Times_New_Roman',Times,serif] text-[11px] text-red-700 overflow-auto max-h-32">
                      {this.state.errorInfo.componentStack}
                    </pre>
                  )}
                </div>
              )}

              <div className="flex gap-2">
                <button
                  onClick={this.handleReset}
                  className="px-4 py-2 border border-black bg-black text-white font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] hover:bg-gray-800 transition-colors"
                >
                  Try Again
                </button>
                <button
                  onClick={() => window.location.href = '/'}
                  className="px-4 py-2 border border-black bg-white text-black font-[Helvetica,Arial,system-ui,sans-serif] text-[12px] font-bold uppercase tracking-[1px] hover:bg-gray-100 transition-colors"
                >
                  Go Home
                </button>
              </div>

              <p className="mt-4 font-['Times_New_Roman',Times,serif] text-[11px] text-gray-500">
                If this problem persists, please{' '}
                <a href="mailto:support@keyles.com" className="text-[#0000ee] underline">
                  contact support
                </a>
              </p>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
