import { Button, Input } from 'antd'

const IntentRecognition = ({ args }: { args: any }) => {
  console.log(args)
  return (
    <>
      <Input.TextArea value={args} />
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

export default IntentRecognition
