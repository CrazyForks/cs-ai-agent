export type SortableKnowledgeDirectory = {
  id: number
  parentId: number
  children?: SortableKnowledgeDirectory[]
}

export type MoveDirectoryResult<T extends SortableKnowledgeDirectory> = {
  items: T[]
  changed: boolean
  parentId: number
  orderedIds: number[]
}

function arrayMove<T>(items: T[], fromIndex: number, toIndex: number) {
  const next = [...items]
  const [item] = next.splice(fromIndex, 1)
  next.splice(toIndex, 0, item)
  return next
}

export function findDirectoryParentId<T extends SortableKnowledgeDirectory>(
  items: T[],
  id: number,
): number | null {
  for (const item of items) {
    if (item.id === id) {
      return item.parentId
    }
    const childParentId = findDirectoryParentId(item.children as T[] | undefined ?? [], id)
    if (childParentId !== null) {
      return childParentId
    }
  }
  return null
}

export function moveDirectoryWithinParent<T extends SortableKnowledgeDirectory>(
  items: T[],
  parentId: number,
  activeId: number,
  overId: number,
): MoveDirectoryResult<T> {
  if (activeId === overId) {
    return { items, changed: false, parentId, orderedIds: [] }
  }

  if (parentId === 0) {
    return moveSiblingList(items, parentId, activeId, overId)
  }

  let changed = false
  let orderedIds: number[] = []
  const nextItems = items.map((item) => {
    if (item.id === parentId) {
      const moved = moveSiblingList((item.children as T[] | undefined) ?? [], parentId, activeId, overId)
      changed = moved.changed
      orderedIds = moved.orderedIds
      return { ...item, children: moved.items }
    }
    if (item.children?.length) {
      const moved = moveDirectoryWithinParent(item.children as T[], parentId, activeId, overId)
      if (moved.changed) {
        changed = true
        orderedIds = moved.orderedIds
        return { ...item, children: moved.items }
      }
    }
    return item
  })

  return {
    items: changed ? nextItems : items,
    changed,
    parentId,
    orderedIds,
  }
}

function moveSiblingList<T extends SortableKnowledgeDirectory>(
  items: T[],
  parentId: number,
  activeId: number,
  overId: number,
): MoveDirectoryResult<T> {
  const oldIndex = items.findIndex((item) => item.id === activeId)
  const newIndex = items.findIndex((item) => item.id === overId)
  if (oldIndex < 0 || newIndex < 0) {
    return { items, changed: false, parentId, orderedIds: [] }
  }

  const nextItems = arrayMove(items, oldIndex, newIndex)
  return {
    items: nextItems,
    changed: true,
    parentId,
    orderedIds: nextItems.map((item) => item.id),
  }
}
