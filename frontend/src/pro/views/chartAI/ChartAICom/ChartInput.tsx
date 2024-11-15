import { Button, ConfigProvider, Input } from 'antd'
import { FaCircleArrowUp } from 'react-icons/fa6'

const ChartInput = () => {
  return (
    <div className="w-full  min-h-[80px] rounded-lg  text-center left-0 flex items-center justify-center ">
      <div className="w-2/3 min-h-[40px] bg-[#2f2f2f] rounded-lg  flex items-center justify-center mb-3 overflow-hidden">
        <ConfigProvider
          theme={{
            components: {
              Input: {
                activeBorderColor: 'none',
                activeBg: '#2f2f2f',
                hoverBorderColor: '#2f2f2f',
                hoverBg: '#2f2f2f',
                activeShadow: '#2f2f2f',
              },
            },
          }}
        >
          <Input.TextArea
            className=" bg-[#2f2f2f] pr-10 "
            variant="filled"
            autoSize={{ minRows: 1, maxRows: 5 }}
            style={{ background: '#2f2f2f ' }}
            placeholder="对话大模型"
          ></Input.TextArea>
        </ConfigProvider>

        <Button
          type="text"
          shape="circle"
          className="mr-2  "
          icon={<FaCircleArrowUp size={24} />}
        ></Button>
      </div>
    </div>
  )
}
export default ChartInput
