import CheckboxFilter from './CheckboxFilter'
import InputFilter from './InputFilter'
import LabelKeyFilter from './LabelKeyFilter'
import { FilterRenderProps } from './type'

const FilterRenderer = ({ item, addFilter, filters }: FilterRenderProps) => {
  if (item?.labelKeys) {
    return <LabelKeyFilter item={item} addFilter={addFilter} filters={filters} />
  }
  return item.wildcard || item?.matchExpr ? (
    <InputFilter item={item} addFilter={addFilter} filters={filters} />
  ) : (
    <CheckboxFilter item={item} addFilter={addFilter} filters={filters} />
  )
}
export default FilterRenderer
