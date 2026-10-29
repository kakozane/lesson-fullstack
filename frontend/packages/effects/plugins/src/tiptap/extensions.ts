import type { Editor as CoreEditor } from '@tiptap/core';
import type { Node as ProseMirrorNode } from '@tiptap/pm/model';
import type { EditorView } from '@tiptap/pm/view';
import type { Extensions } from '@tiptap/vue-3';

import type { ImageUploadOptions, VbenTiptapExtensionOptions } from './types';

import { $t } from '@vben/locales';

import { alert } from '@vben-core/popup-ui';

import Highlight from '@tiptap/extension-highlight';
import Image from '@tiptap/extension-image';
import Link from '@tiptap/extension-link';
import Placeholder from '@tiptap/extension-placeholder';
import TextAlign from '@tiptap/extension-text-align';
import { Color, TextStyle } from '@tiptap/extension-text-style';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import StarterKit from '@tiptap/starter-kit';

const DEFAULT_ACCEPT = 'image/*';
const IMAGE_MAX_WIDTH = 640;
const IMAGE_MIN_WIDTH = 80;

function parseImageWidth(element: HTMLElement): null | number {
  const width = element.getAttribute('width') ?? element.style.width;
  const parsedWidth = Number.parseInt(width, 10);

  return Number.isFinite(parsedWidth) && parsedWidth > 0 ? parsedWidth : null;
}

function renderImageWidth(attributes: Record<string, unknown>) {
  const width = attributes.width;

  if (width === null || width === undefined || width === '') {
    return {};
  }

  return { width: String(width) };
}

function syncImageAttributes(
  img: HTMLImageElement,
  attrs: Record<string, any>,
) {
  img.src = attrs.src;
  img.className = 'vben-tiptap__image';

  if (attrs.alt) {
    img.alt = attrs.alt;
  } else {
    img.removeAttribute('alt');
  }

  if (attrs.title) {
    img.title = attrs.title;
  } else {
    img.removeAttribute('title');
  }

  if (attrs.width) {
    img.style.width = `${attrs.width}px`;
    img.setAttribute('width', String(attrs.width));
  } else {
    img.style.removeProperty('width');
    img.removeAttribute('width');
  }
}

function createResizableImageNodeView(
  editor: CoreEditor,
  node: ProseMirrorNode,
  getPos: (() => number | undefined) | boolean,
) {
  const wrapper = document.createElement('div');
  wrapper.className = 'vben-tiptap-image-resize-wrapper';
  wrapper.contentEditable = 'false';

  const img = document.createElement('img');
  syncImageAttributes(img, node.attrs);
  wrapper.append(img);

  const resizeHandle = document.createElement('span');
  resizeHandle.className = 'vben-tiptap-image-resize-handle';
  wrapper.append(resizeHandle);

  let removeDragListeners: (() => void) | undefined;

  function stopResize() {
    removeDragListeners?.();
    removeDragListeners = undefined;
    wrapper.classList.remove('is-resizing');
  }

  resizeHandle.addEventListener('mousedown', (event) => {
    if (!editor.isEditable || typeof getPos !== 'function') {
      return;
    }

    event.preventDefault();
    event.stopPropagation();

    const startX = event.clientX;
    const startWidth =
      img.getBoundingClientRect().width || img.naturalWidth || IMAGE_MIN_WIDTH;
    const maxWidth = Math.min(
      wrapper.parentElement?.clientWidth ?? IMAGE_MAX_WIDTH,
      IMAGE_MAX_WIDTH,
    );
    let nextWidth = Math.round(startWidth);

    wrapper.classList.add('is-resizing');

    const handleMouseMove = (moveEvent: MouseEvent) => {
      moveEvent.preventDefault();
      nextWidth = Math.round(
        Math.min(
          Math.max(startWidth + moveEvent.clientX - startX, IMAGE_MIN_WIDTH),
          maxWidth,
        ),
      );
      img.style.width = `${nextWidth}px`;
    };

    const handleMouseUp = () => {
      stopResize();

      const pos = getPos();
      if (typeof pos !== 'number') {
        return;
      }

      const currentNode = editor.state.doc.nodeAt(pos);
      if (!currentNode) {
        return;
      }

      const transaction = editor.state.tr.setNodeMarkup(pos, undefined, {
        ...currentNode.attrs,
        width: nextWidth,
      });
      editor.view.dispatch(transaction);
      editor.view.focus();
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp, { once: true });
    removeDragListeners = () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  });

  return {
    dom: wrapper,
    destroy() {
      stopResize();
    },
    update(updatedNode: ProseMirrorNode) {
      if (updatedNode.type.name !== node.type.name) {
        return false;
      }

      syncImageAttributes(img, updatedNode.attrs);
      return true;
    },
  } as any;
}

function validateFile(
  file: File,
  options: ImageUploadOptions,
): string | undefined {
  if (options.maxSize !== undefined && file.size > options.maxSize) {
    return $t('ui.tiptap.upload.fileTooLarge');
  }

  const accept = options.accept ?? DEFAULT_ACCEPT;
  if (accept && accept !== '*/*' && accept !== 'image/*') {
    const acceptedTypes = accept.split(',').map((t) => t.trim());
    const isAccepted = acceptedTypes.some((type) => {
      if (type.endsWith('/*')) {
        return file.type.startsWith(type.slice(0, -1));
      }
      return file.type === type;
    });
    if (!isAccepted) {
      return $t('ui.tiptap.upload.fileTypeNotAllowed');
    }
  }

  return undefined;
}

function handleUploadError(error: unknown, options: ImageUploadOptions): void {
  if (options.onUploadError) {
    options.onUploadError(error);
  } else {
    const message = error instanceof Error ? error.message : String(error);
    alert(message, $t('ui.tiptap.upload.uploadFailed')).catch(() => {});
  }
}

function findPlaceholderPos(doc: ProseMirrorNode, blobUrl: string): number {
  let found = -1;
  doc.descendants((node: ProseMirrorNode, offset: number) => {
    if (found !== -1) return false;
    if (
      node.type.name === 'image' &&
      node.attrs.src === blobUrl &&
      node.attrs['data-uploading'] === 'true'
    ) {
      found = offset;
      return false;
    }
    return true;
  });
  return found;
}

interface UploadContext {
  blobUrl: string;
  pos: number;
}

function createUploadProcess(
  editor: CoreEditor,
  file: File,
  options: ImageUploadOptions,
  blobUrlTracker?: Set<string>,
  pos?: number,
): UploadContext {
  const blobUrl = URL.createObjectURL(file);
  blobUrlTracker?.add(blobUrl);
  const insertPos = pos ?? editor.state.selection.from;

  // Insert placeholder image with blob URL
  editor
    .chain()
    .insertContentAt(insertPos, {
      attrs: {
        'data-upload-progress': 0,
        'data-uploading': 'true',
        src: blobUrl,
      },
      type: 'image',
    })
    .run();

  const nodePos = findPlaceholderPos(editor.state.doc, blobUrl);

  const uploadContext: UploadContext = { blobUrl, pos: nodePos };

  options
    .upload(file, (percent: number) => {
      if (editor.isDestroyed) return;

      const currentPos = findPlaceholderPos(editor.state.doc, blobUrl);
      if (currentPos === -1) return;

      const node = editor.state.doc.nodeAt(currentPos);
      if (!node) return;

      const transaction = editor.state.tr.setNodeMarkup(currentPos, undefined, {
        ...node.attrs,
        'data-upload-progress': percent,
      });
      editor.view.dispatch(transaction);
    })
    .then((url: string) => {
      if (editor.isDestroyed) {
        URL.revokeObjectURL(blobUrl);
        return;
      }

      const currentPos = findPlaceholderPos(editor.state.doc, blobUrl);

      if (currentPos === -1) {
        blobUrlTracker?.delete(blobUrl);
        URL.revokeObjectURL(blobUrl);
        return;
      }

      const node = editor.state.doc.nodeAt(currentPos);
      if (!node) {
        blobUrlTracker?.delete(blobUrl);
        URL.revokeObjectURL(blobUrl);
        return;
      }

      const transaction = editor.state.tr.setNodeMarkup(currentPos, undefined, {
        ...node.attrs,
        'data-upload-progress': null,
        'data-uploading': null,
        src: url,
      });
      editor.view.dispatch(transaction);
      blobUrlTracker?.delete(blobUrl);
      URL.revokeObjectURL(blobUrl);
    })
    .catch((error: unknown) => {
      if (editor.isDestroyed) {
        URL.revokeObjectURL(blobUrl);
        return;
      }

      const currentPos = findPlaceholderPos(editor.state.doc, blobUrl);

      if (currentPos !== -1) {
        const transaction = editor.state.tr.delete(
          currentPos,
          currentPos + (editor.state.doc.nodeAt(currentPos)?.nodeSize ?? 1),
        );
        editor.view.dispatch(transaction);
      }

      URL.revokeObjectURL(blobUrl);
      blobUrlTracker?.delete(blobUrl);
      handleUploadError(error, options);
    });

  return uploadContext;
}

function createCustomImage(
  imageUpload: ImageUploadOptions,
  blobUrlTracker?: Set<string>,
) {
  return Image.extend({
    addAttributes() {
      return {
        ...this.parent?.(),
        'data-upload-progress': {
          default: null,
          parseHTML: (element) => element.dataset.uploadProgress,
          renderHTML: () => {
            return {};
          },
        },
        'data-uploading': {
          default: null,
          parseHTML: (element) => element.dataset.uploading,
          renderHTML: () => {
            return {};
          },
        },
        width: {
          default: null,
          parseHTML: parseImageWidth,
          renderHTML: renderImageWidth,
        },
      };
    },

    addNodeView() {
      return ({ editor, getPos, node }) => {
        const isUploading = node.attrs['data-uploading'] === 'true';

        if (!isUploading) {
          return createResizableImageNodeView(editor, node, getPos);
        }

        const wrapper = document.createElement('div');
        wrapper.className = 'vben-tiptap-upload-wrapper';

        const img = document.createElement('img');
        img.src = node.attrs.src;
        img.className = 'vben-tiptap__image';
        wrapper.append(img);

        const spinner = document.createElement('div');
        spinner.className = 'vben-tiptap-upload-spinner';
        wrapper.append(spinner);

        const progressBar = document.createElement('div');
        progressBar.className = 'vben-tiptap-upload-progress';
        const progressFill = document.createElement('div');
        progressFill.className = 'vben-tiptap-upload-progress-fill';
        progressBar.append(progressFill);
        wrapper.append(progressBar);

        const progress = node.attrs['data-upload-progress'];
        if (progress !== null && progress !== undefined && progress > 0) {
          spinner.style.display = 'none';
          progressBar.style.display = '';
          progressFill.style.width = `${progress}%`;
        } else {
          spinner.style.display = '';
          progressBar.style.display = 'none';
        }

        return {
          dom: wrapper,
          update(updatedNode: ProseMirrorNode) {
            if (updatedNode.attrs['data-uploading'] !== 'true') {
              return false;
            }

            if (updatedNode.attrs.src !== img.src) {
              img.src = updatedNode.attrs.src;
            }

            const newProgress = updatedNode.attrs['data-upload-progress'];
            if (
              newProgress !== null &&
              newProgress !== undefined &&
              newProgress > 0
            ) {
              spinner.style.display = 'none';
              progressBar.style.display = '';
              progressFill.style.width = `${newProgress}%`;
            } else {
              spinner.style.display = '';
              progressBar.style.display = 'none';
            }

            return true;
          },
        } as any;
      };
    },

    addCommands() {
      return {
        ...this.parent?.(),
        uploadImage:
          () =>
          ({ editor: cmdEditor }: { editor: CoreEditor }) => {
            const input = document.createElement('input');
            input.type = 'file';
            input.accept = imageUpload.accept ?? DEFAULT_ACCEPT;
            input.style.display = 'none';

            input.addEventListener('change', () => {
              const file = input.files?.[0];
              if (!file) return;

              const error = validateFile(file, imageUpload);
              if (error) {
                handleUploadError(new Error(error), imageUpload);
                return;
              }

              createUploadProcess(cmdEditor, file, imageUpload, blobUrlTracker);
              input.remove();
            });

            document.body.append(input);
            input.click();
            return true;
          },
      };
    },

    addProseMirrorPlugins() {
      const editor = this.editor;

      return [
        new Plugin({
          key: new PluginKey('imageUploadDrop'),
          props: {
            handleDrop: (view: EditorView, event: DragEvent) => {
              if (!event.dataTransfer?.files.length) return false;

              const imageFiles = [...event.dataTransfer.files].filter((f) =>
                f.type.startsWith('image/'),
              );
              if (imageFiles.length === 0) return false;

              event.preventDefault();

              // Only support single image upload
              const file = imageFiles[0];
              if (!file) return false;
              if (imageFiles.length > 1) {
                handleUploadError(
                  new Error($t('ui.tiptap.upload.onlySingleImage')),
                  imageUpload,
                );
              }

              const error = validateFile(file, imageUpload);
              if (error) {
                handleUploadError(new Error(error), imageUpload);
                return true;
              }

              const coordinates = view.posAtCoords({
                left: event.clientX,
                top: event.clientY,
              });

              const pos = coordinates?.pos ?? view.state.selection.from;

              createUploadProcess(
                editor,
                file,
                imageUpload,
                blobUrlTracker,
                pos,
              );
              return true;
            },
          },
        }),
        new Plugin({
          key: new PluginKey('imageUploadPaste'),
          props: {
            handlePaste: (_view: EditorView, event: ClipboardEvent) => {
              const items = event.clipboardData?.items;
              if (!items) return false;

              const imageFiles: File[] = [];
              for (const item of items) {
                if (item.type.startsWith('image/')) {
                  const file = item.getAsFile();
                  if (file) imageFiles.push(file);
                }
              }

              if (imageFiles.length === 0) return false;

              event.preventDefault();

              const imageFile = imageFiles[0];
              if (!imageFile) return false;
              if (imageFiles.length > 1) {
                handleUploadError(
                  new Error($t('ui.tiptap.upload.onlySingleImage')),
                  imageUpload,
                );
              }

              const error = validateFile(imageFile, imageUpload);
              if (error) {
                handleUploadError(new Error(error), imageUpload);
                return true;
              }

              createUploadProcess(
                editor,
                imageFile,
                imageUpload,
                blobUrlTracker,
              );
              return true;
            },
          },
        }),
      ];
    },
  });
}

function createResizableImage() {
  return Image.extend({
    addAttributes() {
      return {
        ...this.parent?.(),
        width: {
          default: null,
          parseHTML: parseImageWidth,
          renderHTML: renderImageWidth,
        },
      };
    },

    addNodeView() {
      return ({ editor, getPos, node }) => {
        return createResizableImageNodeView(editor, node, getPos);
      };
    },
  });
}

export function createDefaultTiptapExtensions(
  options: VbenTiptapExtensionOptions = {},
): Extensions {
  return [
    StarterKit.configure({
      heading: {
        levels: [1, 2, 3, 4],
      },
      link: false,
    }),
    TextAlign.configure({
      types: ['heading', 'paragraph'],
    }),
    TextStyle,
    Color.configure({
      types: ['textStyle'],
    }),
    Highlight.configure({
      multicolor: true,
    }),
    Link.configure({
      autolink: true,
      defaultProtocol: 'https',
      enableClickSelection: true,
      openOnClick: false,
      protocols: ['mailto', { optionalSlashes: true, scheme: 'tel' }],
    }),
    options.imageUpload
      ? createCustomImage(
          options.imageUpload,
          options._blobUrlTracker,
        ).configure({
          allowBase64: true,
          HTMLAttributes: {
            class: 'vben-tiptap__image',
          },
        })
      : createResizableImage().configure({
          allowBase64: true,
          HTMLAttributes: {
            class: 'vben-tiptap__image',
          },
        }),
    Placeholder.configure({
      placeholder: options.placeholder ?? $t('ui.tiptap.placeholder'),
    }),
  ];
}
