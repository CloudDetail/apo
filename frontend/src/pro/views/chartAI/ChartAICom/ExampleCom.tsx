import { FaRegLightbulb } from 'react-icons/fa'
interface ExampleComProps {
  children: any
}

const ExampleCom: React.FC<ExampleComProps> = ({ children }) => {
  return (
    <div className="p-4 flex ">
      <div className="px-2">{children}</div>
    </div>
  )
}
export default ExampleCom
