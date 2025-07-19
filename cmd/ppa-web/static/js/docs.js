// Document management and viewing functionality
class DocumentViewer {
    constructor() {
        this.currentDocument = null;
        this.searchResults = [];
        this.recentDocuments = JSON.parse(localStorage.getItem('recentDocuments') || '[]');
        
        this.bindEvents();
        this.loadDocumentList();
    }

    bindEvents() {
        // Search functionality
        const searchInput = document.getElementById('doc-search');
        const searchBtn = document.getElementById('search-btn');
        
        if (searchInput) {
            searchInput.addEventListener('input', this.debounce((e) => {
                if (e.target.value.length >= 2) {
                    this.performSearch(e.target.value);
                } else {
                    this.clearSearchResults();
                }
            }, 300));

            // Keyboard shortcuts
            searchInput.addEventListener('keydown', (e) => {
                if (e.ctrlKey && e.key === 'f') {
                    e.preventDefault();
                    searchInput.focus();
                }
            });
        }

        if (searchBtn) {
            searchBtn.addEventListener('click', () => {
                const query = searchInput?.value;
                if (query) {
                    this.performSearch(query);
                }
            });
        }

        // Document tree navigation
        document.addEventListener('click', (e) => {
            if (e.target.classList.contains('doc-item')) {
                e.preventDefault();
                const path = e.target.getAttribute('data-path');
                this.loadDocument(path);
            }
        });

        // Back button
        const backBtn = document.getElementById('back-btn');
        if (backBtn) {
            backBtn.addEventListener('click', () => {
                this.showDocumentList();
            });
        }

        // File type filters
        document.addEventListener('change', (e) => {
            if (e.target.classList.contains('file-type-filter')) {
                this.filterDocuments();
            }
        });
    }

    async loadDocumentList() {
        try {
            const response = await fetch('/api/docs/list');
            const documents = await response.json();
            
            this.renderDocumentTree(documents);
            this.renderRecentDocuments();
        } catch (error) {
            console.error('Failed to load document list:', error);
            this.showError('Failed to load documents');
        }
    }

    renderDocumentTree(documents) {
        const container = document.getElementById('document-tree');
        if (!container) return;

        // Group documents by directory
        const tree = this.buildDocumentTree(documents);
        
        container.innerHTML = this.renderTree(tree);
    }

    buildDocumentTree(documents) {
        const tree = {};
        
        documents.forEach(doc => {
            const parts = doc.relativePath.split('/');
            let current = tree;
            
            parts.forEach((part, index) => {
                if (index === parts.length - 1) {
                    // Leaf node (file)
                    current[part] = doc;
                } else {
                    // Directory node
                    if (!current[part]) {
                        current[part] = {};
                    }
                    current = current[part];
                }
            });
        });
        
        return tree;
    }

    renderTree(tree, path = '') {
        let html = '<ul class="document-tree">';
        
        // Sort entries: directories first, then files
        const entries = Object.entries(tree).sort((a, b) => {
            const aIsFile = typeof a[1].name === 'string';
            const bIsFile = typeof b[1].name === 'string';
            
            if (aIsFile && !bIsFile) return 1;
            if (!aIsFile && bIsFile) return -1;
            return a[0].localeCompare(b[0]);
        });
        
        entries.forEach(([name, item]) => {
            const currentPath = path ? `${path}/${name}` : name;
            
            if (typeof item.name === 'string') {
                // File
                html += `
                    <li class="tree-file">
                        <a href="#" class="doc-item d-flex align-items-center" data-path="${item.relativePath}">
                            <i class="bi bi-file-earmark-text me-2"></i>
                            <div>
                                <div class="fw-medium">${name}</div>
                                ${item.title ? `<small class="text-muted">${item.title}</small>` : ''}
                            </div>
                        </a>
                    </li>
                `;
            } else {
                // Directory
                html += `
                    <li class="tree-directory">
                        <div class="d-flex align-items-center mb-1">
                            <i class="bi bi-folder me-2"></i>
                            <strong>${name}</strong>
                        </div>
                        ${this.renderTree(item, currentPath)}
                    </li>
                `;
            }
        });
        
        html += '</ul>';
        return html;
    }

    renderRecentDocuments() {
        const container = document.getElementById('recent-documents');
        if (!container || this.recentDocuments.length === 0) return;

        container.innerHTML = `
            <h6>Recent Documents</h6>
            <div class="list-group list-group-flush">
                ${this.recentDocuments.slice(0, 5).map(doc => `
                    <a href="#" class="list-group-item list-group-item-action doc-item" data-path="${doc.path}">
                        <div class="d-flex w-100 justify-content-between">
                            <h6 class="mb-1">${doc.name}</h6>
                            <small>${this.formatDate(doc.lastViewed)}</small>
                        </div>
                        ${doc.title ? `<p class="mb-1">${doc.title}</p>` : ''}
                    </a>
                `).join('')}
            </div>
        `;
    }

    async loadDocument(path) {
        try {
            this.showLoading();
            
            const response = await fetch(`/api/docs/view?path=${encodeURIComponent(path)}`);
            const data = await response.json();
            
            this.currentDocument = data;
            this.renderDocument(data);
            this.addToRecentDocuments(data.document);
            
        } catch (error) {
            console.error('Failed to load document:', error);
            this.showError('Failed to load document');
        }
    }

    renderDocument(data) {
        const container = document.getElementById('document-viewer');
        if (!container) return;

        const doc = data.document;
        const content = data.html || data.content;
        
        let renderedContent;
        if (doc.type === '.md' && data.html) {
            // Markdown with HTML rendering
            renderedContent = `<div class="markdown-content">${content}</div>`;
        } else {
            // Plain text
            renderedContent = `<pre class="text-content">${this.escapeHtml(content)}</pre>`;
        }

        container.innerHTML = `
            <div class="document-header mb-4">
                <nav aria-label="breadcrumb">
                    <ol class="breadcrumb">
                        <li class="breadcrumb-item">
                            <a href="#" id="back-btn">
                                <i class="bi bi-arrow-left me-1"></i>
                                Documents
                            </a>
                        </li>
                        <li class="breadcrumb-item active">${doc.name}</li>
                    </ol>
                </nav>
                
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <h2>${doc.title || doc.name}</h2>
                        <p class="text-muted mb-0">
                            <i class="bi bi-calendar me-1"></i>
                            Modified ${this.formatDate(doc.modTime)}
                            <span class="ms-3">
                                <i class="bi bi-file-earmark me-1"></i>
                                ${this.formatFileSize(doc.size)}
                            </span>
                        </p>
                    </div>
                    <div class="btn-group">
                        <button class="btn btn-outline-primary btn-sm" onclick="documentViewer.searchInDocument()">
                            <i class="bi bi-search me-1"></i>
                            Search in Document
                        </button>
                        <button class="btn btn-outline-secondary btn-sm" onclick="documentViewer.generateTOC()">
                            <i class="bi bi-list-nested me-1"></i>
                            Table of Contents
                        </button>
                    </div>
                </div>
            </div>
            
            <div class="document-content">
                ${renderedContent}
            </div>
        `;

        // Syntax highlighting for code blocks
        if (doc.type === '.md' && window.hljs) {
            container.querySelectorAll('pre code').forEach(block => {
                hljs.highlightElement(block);
            });
        }

        // Show document view
        this.showDocumentView();
        
        // Scroll to top
        container.scrollTop = 0;
    }

    async performSearch(query) {
        try {
            const fileTypes = this.getSelectedFileTypes();
            const params = new URLSearchParams({ q: query });
            
            fileTypes.forEach(type => params.append('type', type));
            
            const response = await fetch(`/api/docs/search?${params}`);
            const results = await response.json();
            
            this.searchResults = results;
            this.renderSearchResults(results, query);
            
        } catch (error) {
            console.error('Search failed:', error);
            this.showError('Search failed');
        }
    }

    renderSearchResults(results, query) {
        const container = document.getElementById('search-results');
        if (!container) return;

        if (results.length === 0) {
            container.innerHTML = `
                <div class="alert alert-info">
                    <i class="bi bi-info-circle me-2"></i>
                    No results found for "${query}"
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="search-summary mb-3">
                <h6>Search Results (${results.length})</h6>
                <p class="text-muted mb-0">Found ${results.length} documents matching "${query}"</p>
            </div>
            
            <div class="search-results-list">
                ${results.map(result => this.renderSearchResult(result, query)).join('')}
            </div>
        `;
    }

    renderSearchResult(result, query) {
        const doc = result.document;
        const matches = result.matches.slice(0, 3); // Show up to 3 matches
        
        return `
            <div class="search-result mb-3 p-3 border rounded">
                <div class="d-flex justify-content-between align-items-start mb-2">
                    <div>
                        <h6 class="mb-1">
                            <a href="#" class="doc-item text-decoration-none" data-path="${doc.relativePath}">
                                ${doc.title || doc.name}
                            </a>
                        </h6>
                        <small class="text-muted">${doc.relativePath}</small>
                    </div>
                    <span class="badge bg-primary">${result.score} matches</span>
                </div>
                
                ${matches.length > 0 ? `
                    <div class="matches">
                        ${matches.map(match => `
                            <div class="match mb-2">
                                <small class="text-muted">Line ${match.lineNumber}:</small>
                                <div class="match-line">
                                    <code>${this.highlightSearchTerm(match.line, query)}</code>
                                </div>
                            </div>
                        `).join('')}
                        ${result.matches.length > 3 ? `
                            <small class="text-muted">... and ${result.matches.length - 3} more matches</small>
                        ` : ''}
                    </div>
                ` : ''}
            </div>
        `;
    }

    highlightSearchTerm(text, query) {
        if (!query) return this.escapeHtml(text);
        
        const escapedText = this.escapeHtml(text);
        const escapedQuery = this.escapeHtml(query);
        const regex = new RegExp(`(${escapedQuery})`, 'gi');
        
        return escapedText.replace(regex, '<mark>$1</mark>');
    }

    getSelectedFileTypes() {
        const checkboxes = document.querySelectorAll('.file-type-filter:checked');
        return Array.from(checkboxes).map(cb => cb.value);
    }

    addToRecentDocuments(doc) {
        // Remove if already exists
        this.recentDocuments = this.recentDocuments.filter(d => d.path !== doc.relativePath);
        
        // Add to beginning
        this.recentDocuments.unshift({
            path: doc.relativePath,
            name: doc.name,
            title: doc.title,
            lastViewed: new Date().toISOString()
        });
        
        // Keep only last 10
        this.recentDocuments = this.recentDocuments.slice(0, 10);
        
        // Save to localStorage
        localStorage.setItem('recentDocuments', JSON.stringify(this.recentDocuments));
        
        // Update UI
        this.renderRecentDocuments();
    }

    searchInDocument() {
        if (!this.currentDocument) return;
        
        const query = prompt('Search in document:');
        if (!query) return;
        
        const content = document.querySelector('.document-content');
        if (!content) return;
        
        // Simple text search and highlight
        this.highlightInDocument(content, query);
    }

    highlightInDocument(container, query) {
        // Remove existing highlights
        container.querySelectorAll('.search-highlight').forEach(el => {
            el.outerHTML = el.innerHTML;
        });
        
        // Add new highlights
        const walker = document.createTreeWalker(
            container,
            NodeFilter.SHOW_TEXT,
            null,
            false
        );
        
        const textNodes = [];
        let node;
        while (node = walker.nextNode()) {
            textNodes.push(node);
        }
        
        textNodes.forEach(textNode => {
            if (textNode.textContent.toLowerCase().includes(query.toLowerCase())) {
                const parent = textNode.parentNode;
                const html = textNode.textContent.replace(
                    new RegExp(`(${query})`, 'gi'),
                    '<span class="search-highlight bg-warning">$1</span>'
                );
                
                const wrapper = document.createElement('span');
                wrapper.innerHTML = html;
                parent.replaceChild(wrapper, textNode);
            }
        });
    }

    generateTOC() {
        if (!this.currentDocument) return;
        
        const content = document.querySelector('.document-content');
        if (!content) return;
        
        const headings = content.querySelectorAll('h1, h2, h3, h4, h5, h6');
        if (headings.length === 0) {
            alert('No headings found in document');
            return;
        }
        
        const tocHtml = Array.from(headings).map((heading, index) => {
            const level = parseInt(heading.tagName.charAt(1));
            const id = `toc-${index}`;
            heading.id = id;
            
            return `
                <div class="toc-item" style="margin-left: ${(level - 1) * 20}px">
                    <a href="#${id}" class="text-decoration-none">
                        ${heading.textContent}
                    </a>
                </div>
            `;
        }).join('');
        
        // Show TOC in modal or sidebar
        this.showTOC(tocHtml);
    }

    showTOC(tocHtml) {
        const modal = document.createElement('div');
        modal.className = 'modal fade';
        modal.innerHTML = `
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Table of Contents</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        ${tocHtml}
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        const bsModal = new bootstrap.Modal(modal);
        bsModal.show();
        
        modal.addEventListener('hidden.bs.modal', () => {
            document.body.removeChild(modal);
        });
    }

    clearSearchResults() {
        const container = document.getElementById('search-results');
        if (container) {
            container.innerHTML = '';
        }
    }

    showDocumentList() {
        document.getElementById('document-list-view').style.display = 'block';
        document.getElementById('document-viewer').style.display = 'none';
    }

    showDocumentView() {
        document.getElementById('document-list-view').style.display = 'none';
        document.getElementById('document-viewer').style.display = 'block';
    }

    showLoading() {
        const container = document.getElementById('document-viewer');
        if (container) {
            container.innerHTML = `
                <div class="text-center p-5">
                    <div class="spinner-border" role="status">
                        <span class="visually-hidden">Loading...</span>
                    </div>
                    <p class="mt-2">Loading document...</p>
                </div>
            `;
        }
    }

    showError(message) {
        const container = document.getElementById('document-viewer') || document.getElementById('search-results');
        if (container) {
            container.innerHTML = `
                <div class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    ${message}
                </div>
            `;
        }
    }

    // Utility functions
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// Initialize document viewer when DOM is loaded
let documentViewer;
document.addEventListener('DOMContentLoaded', () => {
    documentViewer = new DocumentViewer();
});

// Global keyboard shortcuts
document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'f') {
        e.preventDefault();
        const searchInput = document.getElementById('doc-search');
        if (searchInput) {
            searchInput.focus();
        }
    }
});
