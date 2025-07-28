/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import remarkBreaks from 'remark-breaks'
import './github-markdown.css'
import ReactMarkdown from 'react-markdown'
import { Hash, Quote, CheckSquare, ExternalLink } from 'lucide-react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'

const MarkdownRender = ({ content, theme }: { content: string; theme: 'light' | 'dark' }) => {
  const components = {
    // Headers with icons
    h1: ({ children }: any) => (
      <h1 className="text-4xl font-bold mb-6 flex items-center gap-3 group">
        <Hash className="w-8 h-8 text-blue-500 group-hover:text-blue-600 transition-colors" />
        {children}
      </h1>
    ),
    h2: ({ children }: any) => (
      <h2 className="text-3xl font-semibold mb-4 mt-8 flex items-center gap-2 group">
        <Hash className="w-6 h-6 text-blue-500 group-hover:text-blue-600 transition-colors" />
        {children}
      </h2>
    ),
    h3: ({ children }: any) => (
      <h3 className="text-2xl font-semibold mb-3 mt-6 flex items-center gap-2 group">
        <Hash className="w-5 h-5 text-blue-500 group-hover:text-blue-600 transition-colors" />
        {children}
      </h3>
    ),

    // Code blocks with syntax highlighting
    code: ({ node, inline, className, children, ...props }: any) => {
      const match = /language-(\w+)/.exec(className || '')
      const language = match ? match[1] : ''

      if (!inline && language) {
        return (
          <div className="my-6 rounded-xl overflow-hidden shadow-lg">
            <div className="bg-gray-800 dark:bg-gray-900 px-4 py-2 flex items-center gap-2">
              <div className="flex gap-1">
                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div className="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
              <span className="text-sm font-medium text-green-400 ml-2">{language}</span>
            </div>
            <SyntaxHighlighter
              style={theme === 'dark' ? oneDark : oneLight}
              language={language}
              PreTag="div"
              className="!m-0 !bg-transparent"
              {...props}
            >
              {String(children).replace(/\n$/, '')}
            </SyntaxHighlighter>
          </div>
        )
      }

      // Inline code
      return (
        <code
          className="px-2 py-1 bg-gray-100 dark:bg-gray-800 text-blue-600 dark:text-blue-400 rounded font-mono text-sm"
          {...props}
        >
          {children}
        </code>
      )
    },

    // Blockquotes with icon
    blockquote: ({ children }: any) => (
      <blockquote className="my-6 border-l-4 border-blue-500 bg-blue-50 dark:bg-blue-900/20 p-4 rounded-r-lg">
        <Quote className="w-6 h-6 text-blue-500 mb-2" />
        <div className="text-[var(--ant-color-text-secondary)] italic leading-relaxed">
          {children}
        </div>
      </blockquote>
    ),

    // Lists with custom styling
    ul: ({ children }: any) => <ul className="my-2 space-y-2">{children}</ul>,

    li: ({ children, ...props }: any) => {
      // Check if it's a task list item
      const isTaskList = props.className?.includes('task-list-item')

      if (isTaskList) {
        return (
          <li className="flex items-start gap-3 text-[var(--ant-color-text)] list-none">
            {children}
          </li>
        )
      }

      return (
        <li className="flex items-start gap-3  text-[var(--ant-color-text)]">
          {/* <List className="w-5 h-5 text-blue-500 mt-0.5 flex-shrink-0" /> */}
          <span>{children}</span>
        </li>
      )
    },

    // Task list checkboxes
    input: ({ type, checked, ...props }: any) => {
      if (type === 'checkbox') {
        return checked ? (
          <CheckSquare className="w-5 h-5 text-green-500 mt-0.5 flex-shrink-0" />
        ) : (
          <div className="w-5 h-5 border-2 border-gray-400 rounded mt-0.5 flex-shrink-0"></div>
        )
      }
      return <input type={type} checked={checked} {...props} />
    },

    // Links with external icon
    a: ({ href, children }: any) => (
      <a
        href={href}
        className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 underline transition-colors inline-flex items-center gap-1"
        target="_blank"
        rel="noopener noreferrer"
      >
        {children}
        <ExternalLink className="w-3 h-3" />
      </a>
    ),

    // Tables with beautiful styling
    table: ({ children }: any) => (
      <div className="my-6 overflow-x-auto">
        <table className="w-full border-collapse rounded-lg shadow-lg overflow-hidden">
          {children}
        </table>
      </div>
    ),

    thead: ({ children }: any) => <thead className="">{children}</thead>,

    th: ({ children }: any) => (
      <th className="px-4 py-3 text-left font-semibold border-b">{children}</th>
    ),

    tbody: ({ children }: any) => <tbody>{children}</tbody>,

    tr: ({ children }: any) => <tr className=" transition-colors">{children}</tr>,

    td: ({ children }: any) => (
      <td className="px-4 py-3  border-b border-gray-100 dark:border-gray-600">{children}</td>
    ),

    // Paragraphs
    p: ({ children }: any) => <p className="mb-2  leading-relaxed">{children}</p>,

    // Horizontal rules
    hr: () => <hr className="my-8 border-gray-300 dark:border-gray-600" />,

    // Strong and emphasis
    strong: ({ children }: any) => <strong className="font-bold">{children}</strong>,

    em: ({ children }: any) => (
      <em className="italic text-[var(--ant-color-text-secondary)]">{children}</em>
    ),

    // Strikethrough
    del: ({ children }: any) => (
      <del className="line-through text-[var(--ant-color-text-secondary)]">{children}</del>
    ),
  }
  return (
    <div
      className="markdown-body "
      style={{ fontSize: 14, background: 'transparent', minHeight: 0, fontWeight: 'normal' }}
    >
      <ReactMarkdown remarkPlugins={[remarkBreaks]} components={components}>
        {content}
      </ReactMarkdown>
    </div>
  )
}
export default MarkdownRender
