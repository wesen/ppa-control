/**
 * Comprehensive logging utility for debugging
 * Provides different log levels and context-aware logging
 */

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

export interface LogEntry {
  timestamp: Date;
  level: LogLevel;
  context: string;
  message: string;
  data?: any;
  error?: Error;
}

class Logger {
  private logs: LogEntry[] = [];
  private maxLogs = 1000;
  private currentLevel = LogLevel.DEBUG;
  private listeners: ((entry: LogEntry) => void)[] = [];

  setLevel(level: LogLevel): void {
    this.currentLevel = level;
    this.info('Logger', `Log level set to ${LogLevel[level]}`);
  }

  addListener(listener: (entry: LogEntry) => void): void {
    this.listeners.push(listener);
  }

  removeListener(listener: (entry: LogEntry) => void): void {
    this.listeners = this.listeners.filter(l => l !== listener);
  }

  private log(level: LogLevel, context: string, message: string, data?: any, error?: Error): void {
    if (level < this.currentLevel) {
      return;
    }

    const entry: LogEntry = {
      timestamp: new Date(),
      level,
      context,
      message,
      data,
      error,
    };

    // Add to internal log storage
    this.logs.push(entry);
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }

    // Console output with proper formatting
    const timestamp = entry.timestamp.toISOString().substr(11, 12);
    const levelStr = LogLevel[level].padEnd(5);
    const contextStr = context.padEnd(15);
    
    let consoleMessage = `[${timestamp}] ${levelStr} ${contextStr} ${message}`;
    
    if (data) {
      consoleMessage += ` | Data: ${JSON.stringify(data, null, 2)}`;
    }
    
    if (error) {
      consoleMessage += ` | Error: ${error.message}\nStack: ${error.stack}`;
    }

    // Use appropriate console method
    switch (level) {
      case LogLevel.DEBUG:
        console.log(`🐛 ${consoleMessage}`);
        break;
      case LogLevel.INFO:
        console.info(`ℹ️  ${consoleMessage}`);
        break;
      case LogLevel.WARN:
        console.warn(`⚠️  ${consoleMessage}`);
        break;
      case LogLevel.ERROR:
        console.error(`❌ ${consoleMessage}`);
        break;
    }

    // Notify listeners
    this.listeners.forEach(listener => {
      try {
        listener(entry);
      } catch (err) {
        console.error('Error in log listener:', err);
      }
    });
  }

  debug(context: string, message: string, data?: any): void {
    this.log(LogLevel.DEBUG, context, message, data);
  }

  info(context: string, message: string, data?: any): void {
    this.log(LogLevel.INFO, context, message, data);
  }

  warn(context: string, message: string, data?: any, error?: Error): void {
    this.log(LogLevel.WARN, context, message, data, error);
  }

  error(context: string, message: string, data?: any, error?: Error): void {
    this.log(LogLevel.ERROR, context, message, data, error);
  }

  // Utility methods for common logging scenarios
  networkRequest(url: string, method: string, data?: any): void {
    this.debug('Network', `${method} ${url}`, data);
  }

  networkResponse(url: string, status: number, data?: any): void {
    if (status >= 400) {
      this.error('Network', `Response ${status} from ${url}`, data);
    } else {
      this.debug('Network', `Response ${status} from ${url}`, data);
    }
  }

  udpPacket(direction: 'sent' | 'received', address: string, type: string, data?: any): void {
    this.debug('UDP', `${direction.toUpperCase()} ${type} ${direction === 'sent' ? 'to' : 'from'} ${address}`, data);
  }

  stateChange(context: string, before: any, after: any): void {
    this.debug('State', `${context} state change`, { before, after });
  }

  userAction(action: string, data?: any): void {
    this.info('User', action, data);
  }

  // Get logs for display
  getLogs(level?: LogLevel, context?: string): LogEntry[] {
    return this.logs.filter(entry => {
      if (level !== undefined && entry.level < level) {
        return false;
      }
      if (context && entry.context !== context) {
        return false;
      }
      return true;
    });
  }

  // Export logs for debugging
  exportLogs(): string {
    return this.logs.map(entry => {
      const timestamp = entry.timestamp.toISOString();
      const level = LogLevel[entry.level];
      let line = `[${timestamp}] ${level} ${entry.context}: ${entry.message}`;
      
      if (entry.data) {
        line += `\nData: ${JSON.stringify(entry.data, null, 2)}`;
      }
      
      if (entry.error) {
        line += `\nError: ${entry.error.message}\nStack: ${entry.error.stack}`;
      }
      
      return line;
    }).join('\n\n');
  }

  // Clear logs
  clear(): void {
    this.logs = [];
    this.info('Logger', 'Logs cleared');
  }
}

// Global logger instance
export const logger = new Logger();

// Development helper to expose logger globally
if (__DEV__) {
  (global as any).logger = logger;
  (global as any).LogLevel = LogLevel;
  
  logger.info('Logger', 'Logger initialized in development mode');
  logger.info('Logger', 'Access logger globally with: global.logger');
  logger.info('Logger', 'Change log level with: global.logger.setLevel(global.LogLevel.INFO)');
}

export default logger;