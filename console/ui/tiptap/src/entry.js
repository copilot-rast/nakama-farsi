import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Link from '@tiptap/extension-link';
import Placeholder from '@tiptap/extension-placeholder';
import Table from '@tiptap/extension-table';
import TableRow from '@tiptap/extension-table-row';
import TableHeader from '@tiptap/extension-table-header';
import TableCell from '@tiptap/extension-table-cell';
import TaskList from '@tiptap/extension-task-list';
import TaskItem from '@tiptap/extension-task-item';
import TextAlign from '@tiptap/extension-text-align';
import Underline from '@tiptap/extension-underline';

function create(opts) {
  return new Editor({
    element: opts.element,
    content: opts.content || '',
    editorProps: { attributes: { class: 'eco-rt-body', dir: opts.dir || 'auto', role: 'textbox', 'aria-multiline': 'true', 'aria-label': opts.label || '' } },
    extensions: [
      StarterKit.configure({ heading: { levels: [1, 2, 3, 4] } }),
      Underline,
      Link.configure({ openOnClick: false, autolink: true, protocols: ['mailto', 'tel'], HTMLAttributes: { rel: 'noopener noreferrer nofollow', target: '_blank' } }),
      Placeholder.configure({ placeholder: opts.placeholder || '' }),
      Table.configure({ resizable: false }),
      TableRow,
      TableHeader,
      TableCell,
      TaskList,
      TaskItem.configure({ nested: true }),
      TextAlign.configure({ types: ['heading', 'paragraph'], alignments: ['left', 'center', 'right', 'justify'] })
    ],
    onUpdate: opts.onUpdate
  });
}

window.NakamaTiptap = { create: create, Editor: Editor, version: '2.27.2' };
