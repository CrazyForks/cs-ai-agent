import type { Editor } from "@tiptap/react"

export type UploadedEditorImage = {
  assetId: string
  provider: string
  storageKey: string
  filename?: string
}

export function removeEditorImageByTitle(editor: Editor, title: string) {
  const { state } = editor
  let targetPos: number | null = null
  state.doc.descendants((node, pos) => {
    if (node.type.name === "image" && node.attrs.title === title) {
      targetPos = pos
      return false
    }
    return true
  })
  if (targetPos === null) {
    return
  }
  editor.chain().focus().deleteRange({ from: targetPos, to: targetPos + 1 }).run()
}

export function markEditorImageUploadedByTitle(
  editor: Editor,
  title: string,
  uploaded: UploadedEditorImage
) {
  const { state, view } = editor
  let targetPos: number | null = null
  state.doc.descendants((node, pos) => {
    if (node.type.name === "image" && node.attrs.title === title) {
      targetPos = pos
      return false
    }
    return true
  })
  if (targetPos === null) {
    return
  }
  const attrs = view.state.doc.nodeAt(targetPos)?.attrs
  const transaction = view.state.tr.setNodeMarkup(targetPos, undefined, {
    ...attrs,
    alt: uploaded.filename || attrs?.alt || "image",
    dataAssetId: uploaded.assetId,
    dataProvider: uploaded.provider,
    dataStorageKey: uploaded.storageKey,
    title: "",
  })
  view.dispatch(transaction)
}

export function buildSendableEditorHTML(html: string) {
  if (typeof document === "undefined" || !html.includes("<img")) {
    return html
  }

  const template = document.createElement("template")
  template.innerHTML = html
  for (const image of Array.from(template.content.querySelectorAll("img"))) {
    if (
      image.getAttribute("data-asset-id") &&
      image.getAttribute("src")?.startsWith("blob:")
    ) {
      image.removeAttribute("src")
    }
  }
  return template.innerHTML
}

export function hasUploadingEditorImages(html: string) {
  if (typeof document === "undefined" || !html.includes("<img")) {
    return /<img\b[^>]*\btitle=(["'])uploading-[^"']+\1/i.test(html)
  }

  const template = document.createElement("template")
  template.innerHTML = html
  return Array.from(template.content.querySelectorAll("img")).some((image) =>
    image.getAttribute("title")?.startsWith("uploading-")
  )
}

export function revokeEditorObjectUrl(urls: Set<string>, objectUrl: string) {
  if (!urls.delete(objectUrl)) {
    return
  }
  URL.revokeObjectURL(objectUrl)
}

export function revokeEditorObjectUrls(urls: Set<string>) {
  for (const objectUrl of urls) {
    URL.revokeObjectURL(objectUrl)
  }
  urls.clear()
}
