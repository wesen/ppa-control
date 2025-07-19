// Type definitions for PPA packet analysis

export interface PCAPFile {
    id: string;
    name: string;
    size: number;
    uploadDate: Date;
    status: 'uploaded' | 'analyzing' | 'analyzed' | 'error';
    analysisProgress?: number;
    errorMessage?: string;
}

export interface Packet {
    id: string;
    timestamp: Date;
    source: string;
    destination: string;
    messageType: PacketType;
    size: number;
    payload: Uint8Array;
    metadata?: Record<string, any>;
}

export type PacketType = 'request' | 'response' | 'error' | 'other';

export interface AnalysisResult {
    id: string;
    pcapFileId: string;
    createdAt: Date;
    totalPackets: number;
    timeRange: {
        start: Date;
        end: Date;
    };
    messageTypeDistribution: Record<PacketType, number>;
    packets: Packet[];
    summary: string;
    metadata: Record<string, any>;
}

export interface AnalysisStatus {
    pcapFileId: string;
    status: 'idle' | 'analyzing' | 'completed' | 'error';
    progress: number;
    message: string;
    estimatedTimeRemaining?: number;
}

export interface SearchResult {
    documentId: string;
    title: string;
    content: string;
    matches: Array<{
        line: number;
        text: string;
        highlight: string;
    }>;
}

export interface Document {
    id: string;
    title: string;
    content: string;
    type: 'markdown' | 'text' | 'analysis';
    createdAt: Date;
    updatedAt: Date;
    tags: string[];
}

export interface APIError {
    message: string;
    code?: string;
    details?: Record<string, any>;
}

export interface AppState {
    // PCAP Files
    pcapFiles: PCAPFile[];
    selectedPcapFile?: PCAPFile;
    
    // Analysis
    analysisResults: AnalysisResult[];
    currentAnalysis?: AnalysisResult;
    analysisStatus: Record<string, AnalysisStatus>;
    
    // Documents and Search
    documents: Document[];
    searchResults: SearchResult[];
    searchQuery: string;
    
    // UI State
    activeView: 'dashboard' | 'analysis' | 'packets' | 'documents';
    selectedPacket?: Packet;
    sidebarVisible: boolean;
    loading: boolean;
    error?: APIError;
}
