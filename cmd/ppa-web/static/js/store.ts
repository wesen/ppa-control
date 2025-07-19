import { create } from 'zustand';
import { subscribeWithSelector, devtools } from 'zustand/middleware';
import type { AppState, PCAPFile, AnalysisResult, AnalysisStatus, Document, SearchResult, APIError } from './types';

interface AppActions {
    // PCAP File actions
    addPcapFile: (file: PCAPFile) => void;
    updatePcapFile: (id: string, updates: Partial<PCAPFile>) => void;
    removePcapFile: (id: string) => void;
    selectPcapFile: (file: PCAPFile | undefined) => void;
    
    // Analysis actions
    addAnalysisResult: (result: AnalysisResult) => void;
    setCurrentAnalysis: (analysis: AnalysisResult | undefined) => void;
    updateAnalysisStatus: (pcapFileId: string, status: AnalysisStatus) => void;
    
    // Document actions
    setDocuments: (documents: Document[]) => void;
    addDocument: (document: Document) => void;
    
    // Search actions
    setSearchQuery: (query: string) => void;
    setSearchResults: (results: SearchResult[]) => void;
    
    // UI actions
    setActiveView: (view: AppState['activeView']) => void;
    selectPacket: (packet: AppState['selectedPacket']) => void;
    setSidebarVisible: (visible: boolean) => void;
    setLoading: (loading: boolean) => void;
    setError: (error: APIError | undefined) => void;
    
    // Async actions
    uploadPcapFile: (file: File) => Promise<void>;
    startAnalysis: (pcapFileId: string) => Promise<void>;
    searchDocuments: (query: string) => Promise<void>;
    loadAnalysisResult: (resultId: string) => Promise<void>;
}

export const useAppStore = create<AppState & AppActions>()(
    devtools(
        subscribeWithSelector((set, get) => ({
        // Initial state
        pcapFiles: [],
        selectedPcapFile: undefined,
        analysisResults: [],
        currentAnalysis: undefined,
        analysisStatus: {},
        documents: [],
        searchResults: [],
        searchQuery: '',
        activeView: 'dashboard',
        selectedPacket: undefined,
        sidebarVisible: true,
        loading: false,
        error: undefined,
        
        // PCAP File actions
        addPcapFile: (file) => set((state) => ({
            pcapFiles: [...state.pcapFiles, file]
        })),
        
        updatePcapFile: (id, updates) => set((state) => ({
            pcapFiles: state.pcapFiles.map(f => f.id === id ? { ...f, ...updates } : f),
            selectedPcapFile: state.selectedPcapFile?.id === id 
                ? { ...state.selectedPcapFile, ...updates } 
                : state.selectedPcapFile
        })),
        
        removePcapFile: (id) => set((state) => ({
            pcapFiles: state.pcapFiles.filter(f => f.id !== id),
            selectedPcapFile: state.selectedPcapFile?.id === id ? undefined : state.selectedPcapFile
        })),
        
        selectPcapFile: (file) => set({ selectedPcapFile: file }),
        
        // Analysis actions
        addAnalysisResult: (result) => set((state) => ({
            analysisResults: [...state.analysisResults, result]
        })),
        
        setCurrentAnalysis: (analysis) => set({ currentAnalysis: analysis }),
        
        updateAnalysisStatus: (pcapFileId, status) => set((state) => ({
            analysisStatus: { ...state.analysisStatus, [pcapFileId]: status }
        })),
        
        // Document actions
        setDocuments: (documents) => set({ documents }),
        
        addDocument: (document) => set((state) => ({
            documents: [...state.documents, document]
        })),
        
        // Search actions
        setSearchQuery: (query) => set({ searchQuery: query }),
        
        setSearchResults: (results) => set({ searchResults: results }),
        
        // UI actions
        setActiveView: (view) => set({ activeView: view }),
        
        selectPacket: (packet) => set({ selectedPacket: packet }),
        
        setSidebarVisible: (visible) => set({ sidebarVisible: visible }),
        
        setLoading: (loading) => set({ loading }),
        
        setError: (error) => set({ error }),
        
        // Async actions
        uploadPcapFile: async (file) => {
            const { setLoading, setError, addPcapFile } = get();
            
            try {
                setLoading(true);
                setError(undefined);
                
                const formData = new FormData();
                formData.append('file', file);
                
                const response = await fetch('/api/pcap/upload', {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error(`Upload failed: ${response.statusText}`);
                }
                
                const result = await response.json();
                
                const pcapFile: PCAPFile = {
                    id: result.id,
                    name: file.name,
                    size: file.size,
                    uploadDate: new Date(),
                    status: 'uploaded'
                };
                
                addPcapFile(pcapFile);
            } catch (error) {
                setError({
                    message: error instanceof Error ? error.message : 'Upload failed',
                    code: 'UPLOAD_ERROR'
                });
            } finally {
                setLoading(false);
            }
        },
        
        startAnalysis: async (pcapFileId) => {
            const { setLoading, setError, updateAnalysisStatus, updatePcapFile } = get();
            
            try {
                setLoading(true);
                setError(undefined);
                
                // Update status to analyzing
                updateAnalysisStatus(pcapFileId, {
                    pcapFileId,
                    status: 'analyzing',
                    progress: 0,
                    message: 'Starting analysis...'
                });
                
                updatePcapFile(pcapFileId, { status: 'analyzing', analysisProgress: 0 });
                
                const response = await fetch(`/api/pcap/${pcapFileId}/analyze`, {
                    method: 'POST'
                });
                
                if (!response.ok) {
                    throw new Error(`Analysis failed: ${response.statusText}`);
                }
                
                // Start polling for status updates
                const pollStatus = async () => {
                    try {
                        const statusResponse = await fetch(`/api/pcap/${pcapFileId}/status`);
                        if (statusResponse.ok) {
                            const status = await statusResponse.json();
                            updateAnalysisStatus(pcapFileId, status);
                            updatePcapFile(pcapFileId, { 
                                status: status.status === 'completed' ? 'analyzed' : 'analyzing',
                                analysisProgress: status.progress 
                            });
                            
                            if (status.status === 'analyzing') {
                                setTimeout(pollStatus, 1000);
                            }
                        }
                    } catch (error) {
                        console.error('Status polling error:', error);
                    }
                };
                
                setTimeout(pollStatus, 1000);
                
            } catch (error) {
                setError({
                    message: error instanceof Error ? error.message : 'Analysis failed',
                    code: 'ANALYSIS_ERROR'
                });
                updatePcapFile(pcapFileId, { status: 'error' });
            } finally {
                setLoading(false);
            }
        },
        
        searchDocuments: async (query) => {
            const { setLoading, setError, setSearchQuery, setSearchResults } = get();
            
            try {
                setLoading(true);
                setError(undefined);
                setSearchQuery(query);
                
                if (!query.trim()) {
                    setSearchResults([]);
                    return;
                }
                
                const response = await fetch(`/api/docs/search?q=${encodeURIComponent(query)}`);
                
                if (!response.ok) {
                    throw new Error(`Search failed: ${response.statusText}`);
                }
                
                const results = await response.json();
                setSearchResults(results);
                
            } catch (error) {
                setError({
                    message: error instanceof Error ? error.message : 'Search failed',
                    code: 'SEARCH_ERROR'
                });
                setSearchResults([]);
            } finally {
                setLoading(false);
            }
        },
        
        loadAnalysisResult: async (resultId) => {
            const { setLoading, setError, setCurrentAnalysis } = get();
            
            try {
                setLoading(true);
                setError(undefined);
                
                const response = await fetch(`/api/analysis/${resultId}`);
                
                if (!response.ok) {
                    throw new Error(`Failed to load analysis: ${response.statusText}`);
                }
                
                const result = await response.json();
                
                // Convert date strings to Date objects
                const analysisResult: AnalysisResult = {
                    ...result,
                    createdAt: new Date(result.createdAt),
                    timeRange: {
                        start: new Date(result.timeRange.start),
                        end: new Date(result.timeRange.end)
                    },
                    packets: result.packets.map((p: any) => ({
                        ...p,
                        timestamp: new Date(p.timestamp),
                        payload: new Uint8Array(p.payload)
                    }))
                };
                
                setCurrentAnalysis(analysisResult);
                
            } catch (error) {
                setError({
                    message: error instanceof Error ? error.message : 'Failed to load analysis',
                    code: 'LOAD_ERROR'
                });
            } finally {
                setLoading(false);
            }
        }
        })),
        { name: 'ppa-app-store' }
    )
);

// Selector hooks for common use cases
export const useSelectedPcapFile = () => useAppStore(state => state.selectedPcapFile);
export const useCurrentAnalysis = () => useAppStore(state => state.currentAnalysis);
export const useActiveView = () => useAppStore(state => state.activeView);
export const useLoading = () => useAppStore(state => state.loading);
export const useError = () => useAppStore(state => state.error);
