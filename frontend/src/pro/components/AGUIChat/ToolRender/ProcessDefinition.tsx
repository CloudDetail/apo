import { Button, Input } from 'antd'

const ProcessDefinition = ({ args, status }: { args: any }) => {
  console.log(args)
  return (
    <>
      <Input.TextArea value={args} autoSize />
      <div className="text-right mt-2">
        <Button
          type="primary"
          onClick={() => {
            console.log(args)
          }}
        >
          5s 自动执行
        </Button>
      </div>
    </>
  )
}

export default ProcessDefinition
