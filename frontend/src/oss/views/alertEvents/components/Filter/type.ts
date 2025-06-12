export interface FilterItem {
    key: string
    name: string
    selected?: string[]
    matchExpr?: string
    oldKey?: string
    labelKeys?: string[]
    wildcard?: boolean
    isLabelKey?: boolean
    selectedOptions?: { label: React.ReactNode; value: string }[]
}

export interface FilterProps {
    keys: { key: string; name: string; wildcard?: boolean }[]
    labelKeys: string[]
    filters: FilterItem[]
    setFilters: any
}
export interface AddFilterProps {
    filters: FilterItem[]
    keys: { key: string; name: string; wildcard?: boolean }[]
    onAddFilter: (filter: FilterItem) => void
}
export interface LabelProps {
    itemKey: string
    display?: string
    value: string
}
export interface FilterRenderProps {
    item: FilterItem
    addFilter: (filter: FilterItem) => void
    filters: FilterItem[]
}
export interface FilterTagItemProps {
    filters: FilterItem[]
    onDeleteFilter: (filter: FilterItem) => void
    onChangeFilter: (filter: FilterItem) => void
    item: FilterItem
}