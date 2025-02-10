import { Button, Input } from 'antd'
import { RiArrowGoBackFill } from 'react-icons/ri'
interface ToolboxProps {
  searchValue: string
  setSearchValue: any
  isFocused: boolean
  resetView: any
}
const Toolbox = ({ searchValue, setSearchValue, isFocused, resetView }: ToolboxProps) => {
  return (
    <div className="flex justify-between mb-3">
      <Input.Search
        placeholder="input search text"
        value={searchValue}
        onChange={(e) => setSearchValue(e.target.value)}
        style={{ width: 200 }}
      />
      <Button disabled={!isFocused} onClick={resetView}>
        <RiArrowGoBackFill />
      </Button>
    </div>
  )
}
export default Toolbox
