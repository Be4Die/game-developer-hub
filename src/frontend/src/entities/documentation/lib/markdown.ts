/**
 * Безопасный и производительный рендерер Markdown в HTML для документации GDH.
 * Поддерживает:
 * - Заголовки h2, h3, h4
 * - Стилизованные блоки кода с указанием языка (```lang ... ```)
 * - Markdown-таблицы (| th | th |)
 * - GitHub-style алерты (> [!NOTE], > [!WARNING], > [!TIP], > [!IMPORTANT])
 * - Изображения с подписями (![alt](url))
 * - Списки (маркированные - / * и нумерованные 1.)
 * - Ссылки, инлайн-код, жирный и курсивный текст
 */

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

function inlineFormat(text: string): string {
  return text
    // Инлайн-код
    .replace(/`([^`]+)`/g, (_, code) => `<code class="inline-code">${escapeHtml(code)}</code>`)
    // Жирный шрифт
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    // Курсив
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    // Ссылки
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, txt, url) => {
      const isExt = url.startsWith('http://') || url.startsWith('https://') || url.startsWith('//');
      const target = isExt ? ' target="_blank" rel="noopener noreferrer"' : '';
      return `<a href="${escapeHtml(url)}" class="doc-link"${target}>${txt}</a>`;
    });
}

export function renderDocMarkdown(text?: string | null): string {
  if (!text) return '';

  const lines = text.trim().split('\n');
  const result: string[] = [];

  let inCode = false;
  let codeLang = '';
  let codeLines: string[] = [];

  let inTable = false;
  let tableHeader: string[] = [];
  let tableRows: string[][] = [];

  let inList = false;
  let listType: 'ul' | 'ol' = 'ul';
  let listItems: string[] = [];

  let inQuote = false;
  let quoteType: 'note' | 'warning' | 'tip' | 'important' = 'note';
  let quoteLines: string[] = [];

  const flushList = () => {
    if (!inList) return;
    const tag = listType;
    const lis = listItems.map((it) => `<li>${inlineFormat(it)}</li>`).join('');
    result.push(`<${tag} class="doc-list doc-${tag}">${lis}</${tag}>`);
    inList = false;
    listItems = [];
  };

  const flushQuote = () => {
    if (!inQuote) return;
    const content = quoteLines.map((l) => inlineFormat(l)).join('<br/>');
    let calloutClass = 'callout-info';
    let title = 'Примечание';

    if (quoteType === 'warning') {
      calloutClass = 'callout-warning';
      title = 'Внимание';
    } else if (quoteType === 'tip') {
      calloutClass = 'callout-tip';
      title = 'Полезный совет';
    } else if (quoteType === 'important') {
      calloutClass = 'callout-important';
      title = 'Важно';
    }

    result.push(
      `<div class="doc-callout ${calloutClass}">` +
        `<div class="callout-header"><span class="callout-title">${title}</span></div>` +
        `<div class="callout-body">${content}</div>` +
      `</div>`
    );
    inQuote = false;
    quoteLines = [];
  };

  const flushTable = () => {
    if (!inTable) return;
    let html = `<div class="doc-table-wrapper"><table class="doc-table">`;
    if (tableHeader.length > 0) {
      html += `<thead><tr>` + tableHeader.map((h) => `<th>${inlineFormat(h)}</th>`).join('') + `</tr></thead>`;
    }
    html += `<tbody>`;
    for (const row of tableRows) {
      html += `<tr>` + row.map((c) => `<td>${inlineFormat(c)}</td>`).join('') + `</tr>`;
    }
    html += `</tbody></table></div>`;
    result.push(html);
    inTable = false;
    tableHeader = [];
    tableRows = [];
  };

  for (let i = 0; i < lines.length; i++) {
    const rawLine = lines[i];
    const line = rawLine.trim();

    // Блоки кода (```lang)
    if (line.startsWith('```')) {
      flushList();
      flushQuote();
      flushTable();
      if (!inCode) {
        inCode = true;
        codeLang = line.slice(3).trim();
        codeLines = [];
      } else {
        inCode = false;
        result.push(
          `<div class="doc-code-block">` +
            `<div class="code-block-header"><span class="code-language-tag">${escapeHtml(codeLang || 'text')}</span></div>` +
            `<pre class="code-block-pre"><code>${escapeHtml(codeLines.join('\n'))}</code></pre>` +
          `</div>`
        );
        codeLang = '';
        codeLines = [];
      }
      continue;
    }

    if (inCode) {
      codeLines.push(rawLine);
      continue;
    }

    // Таблицы Markdown
    if (line.startsWith('|') && line.endsWith('|')) {
      flushList();
      flushQuote();
      const cells = line.split('|').slice(1, -1).map((c) => c.trim());
      if (cells.every((c) => /^:?-+:?$/.test(c))) {
        // Разделительная строка колонок |---|---|
        continue;
      }
      if (!inTable) {
        inTable = true;
        tableHeader = cells;
      } else {
        tableRows.push(cells);
      }
      continue;
    } else {
      flushTable();
    }

    // Цитаты и алерты (> [!NOTE] / > [!WARNING])
    if (line.startsWith('>')) {
      flushList();
      const qText = line.replace(/^>\s?/, '').trim();
      if (!inQuote) {
        inQuote = true;
        if (/^\[!(WARNING|CAUTION)\]/i.test(qText)) {
          quoteType = 'warning';
          const cleanText = qText.replace(/^\[!(WARNING|CAUTION)\]\s*/i, '');
          quoteLines = cleanText ? [cleanText] : [];
        } else if (/^\[!TIP\]/i.test(qText)) {
          quoteType = 'tip';
          const cleanText = qText.replace(/^\[!TIP\]\s*/i, '');
          quoteLines = cleanText ? [cleanText] : [];
        } else if (/^\[!IMPORTANT\]/i.test(qText)) {
          quoteType = 'important';
          const cleanText = qText.replace(/^\[!IMPORTANT\]\s*/i, '');
          quoteLines = cleanText ? [cleanText] : [];
        } else if (/^\[!NOTE\]/i.test(qText)) {
          quoteType = 'note';
          const cleanText = qText.replace(/^\[!NOTE\]\s*/i, '');
          quoteLines = cleanText ? [cleanText] : [];
        } else {
          quoteType = 'note';
          quoteLines = [qText];
        }
      } else {
        quoteLines.push(qText);
      }
      continue;
    } else {
      flushQuote();
    }

    // Изображения (![alt](url))
    const imgMatch = line.match(/^!\[(.*?)\]\((.*?)\)$/);
    if (imgMatch) {
      flushList();
      const alt = imgMatch[1];
      const src = imgMatch[2];
      result.push(
        `<figure class="doc-media-figure">` +
          `<img src="${escapeHtml(src)}" alt="${escapeHtml(alt)}" class="doc-media-img" loading="lazy" />` +
          (alt ? `<figcaption class="doc-media-caption">${inlineFormat(alt)}</figcaption>` : '') +
        `</figure>`
      );
      continue;
    }

    // Заголовки
    if (/^####\s+/.test(line)) {
      flushList();
      result.push(`<h4 class="doc-h4">${inlineFormat(line.replace(/^####\s+/, ''))}</h4>`);
      continue;
    }
    if (/^###\s+/.test(line)) {
      flushList();
      result.push(`<h3 class="doc-h3">${inlineFormat(line.replace(/^###\s+/, ''))}</h3>`);
      continue;
    }
    if (/^##\s+/.test(line)) {
      flushList();
      result.push(`<h2 class="doc-h2">${inlineFormat(line.replace(/^##\s+/, ''))}</h2>`);
      continue;
    }

    // Маркированные списки (- или *)
    const ulMatch = line.match(/^[-*]\s+(.*)/);
    if (ulMatch) {
      if (!inList || listType !== 'ul') {
        flushList();
        inList = true;
        listType = 'ul';
      }
      listItems.push(ulMatch[1]);
      continue;
    }

    // Нумерованные списки (1. )
    const olMatch = line.match(/^\d+\.\s+(.*)/);
    if (olMatch) {
      if (!inList || listType !== 'ol') {
        flushList();
        inList = true;
        listType = 'ol';
      }
      listItems.push(olMatch[1]);
      continue;
    }

    flushList();

    // Пустая строка
    if (line === '') {
      continue;
    }

    // Обычный абзац
    result.push(`<p class="doc-p">${inlineFormat(line)}</p>`);
  }

  flushList();
  flushQuote();
  flushTable();

  return result.join('\n');
}
