import { useMemo } from 'preact/hooks';

interface MarkdownRendererProps {
    content: string;
    searchTerm?: string;
}

export function MarkdownRenderer({ content, searchTerm }: MarkdownRendererProps) {
    const htmlContent = useMemo(() => {
        return parseMarkdown(content, searchTerm);
    }, [content, searchTerm]);

    return (
        <div 
            className="markdown-content"
            dangerouslySetInnerHTML={{ __html: htmlContent }}
        />
    );
}

function parseMarkdown(markdown: string, searchTerm?: string): string {
    let html = markdown;

    // Escape HTML
    html = html.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

    // Headers
    html = html.replace(/^### (.*$)/gm, '<h3>$1</h3>');
    html = html.replace(/^## (.*$)/gm, '<h2>$1</h2>');
    html = html.replace(/^# (.*$)/gm, '<h1>$1</h1>');

    // Bold and italic
    html = html.replace(/\*\*\*(.*?)\*\*\*/g, '<strong><em>$1</em></strong>');
    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');

    // Code blocks
    html = html.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>');
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

    // Links
    html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>');

    // Lists
    html = html.replace(/^\* (.+$)/gm, '<li>$1</li>');
    html = html.replace(/^\d+\. (.+$)/gm, '<li>$1</li>');
    
    // Wrap consecutive <li> elements in <ul> or <ol>
    html = html.replace(/(<li>.*<\/li>)/gs, (match) => {
        if (match.includes('<li>')) {
            return '<ul>' + match + '</ul>';
        }
        return match;
    });

    // Line breaks
    html = html.replace(/\n\n/g, '</p><p>');
    html = '<p>' + html + '</p>';

    // Clean up empty paragraphs
    html = html.replace(/<p><\/p>/g, '');
    html = html.replace(/<p>\s*<h/g, '<h');
    html = html.replace(/<\/h([1-6])>\s*<\/p>/g, '</h$1>');
    html = html.replace(/<p>\s*<ul>/g, '<ul>');
    html = html.replace(/<\/ul>\s*<\/p>/g, '</ul>');
    html = html.replace(/<p>\s*<pre>/g, '<pre>');
    html = html.replace(/<\/pre>\s*<\/p>/g, '</pre>');

    // Highlight search terms
    if (searchTerm && searchTerm.trim()) {
        const regex = new RegExp(`(${escapeRegex(searchTerm)})`, 'gi');
        html = html.replace(regex, '<mark>$1</mark>');
    }

    return html;
}

function escapeRegex(string: string): string {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
