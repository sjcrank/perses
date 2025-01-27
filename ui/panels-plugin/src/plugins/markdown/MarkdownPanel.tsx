// Copyright 2022 The Perses Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { useMemo } from 'react';
import * as DOMPurify from 'dompurify';
import { marked } from 'marked';
import { Box, Theme } from '@mui/material';
import { PanelProps } from '@perses-dev/plugin-system';
import { MarkdownPanelOptions } from './markdown-panel-model';

export type MarkdownPanelProps = PanelProps<MarkdownPanelOptions>;

function createMarkdownPanelStyles(theme: Theme) {
  return {
    // Make the content scrollable
    height: '100%',
    overflowY: 'auto',
    // Ignore first margin
    '& :first-child': {
      marginTop: 0,
    },
    // Styles for <code>
    '& code': { fontSize: '0.85em' },
    '& :not(pre) code': {
      padding: '0.2em 0.4em',
      backgroundColor: theme.palette.grey[100],
      borderRadius: '4px',
    },
    '& pre': {
      padding: '1.2em',
      backgroundColor: theme.palette.grey[100],
      borderRadius: '4px',
    },
    // Styles for <table>
    '& table, & th, & td': {
      padding: '0.6em',
      border: `1px solid ${theme.palette.grey[300]}`,
      borderCollapse: 'collapse',
    },
    // Styles for <li>
    '& li + li': {
      marginTop: '0.25em',
    },
    // Styles for <a>
    '& a': {
      color: theme.palette.primary.main,
    },
  };
}

// Convert markdown text to HTML, with support for original markdown and Github Flavored Markdown
function convertTextToHTML(text: string): string {
  return marked.parse(text, { gfm: true });
}

// Sanitize HTML (i.e. prevent XSS attacks by removing the vectors for these attacks)
function sanitizeHTML(html: string): string {
  return DOMPurify.sanitize(html);
}

export function MarkdownPanel(props: MarkdownPanelProps) {
  const {
    spec: { text },
  } = props;

  const html = useMemo(() => convertTextToHTML(text), [text]);
  const sanitizedHTML = useMemo(() => sanitizeHTML(html), [html]);

  return <Box sx={createMarkdownPanelStyles} dangerouslySetInnerHTML={{ __html: sanitizedHTML }} />;
}
