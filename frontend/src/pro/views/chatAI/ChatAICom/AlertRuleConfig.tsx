import { Button, Input, message, Modal, Steps, Upload, UploadProps } from 'antd'
import { useState } from 'react'
import { IoMdAdd, IoMdAddCircleOutline } from 'react-icons/io'
import { MdOutlineUploadFile } from 'react-icons/md'

const UploadCom: React.FC = () => {
  const [files, setFiles] = useState([])
  const props: UploadProps = {
    customRequest: ({ file }) => {
      setFiles([file])
    },
    fileList: files,
    onChange(info) {
      if (info.file.status !== 'uploading') {
        console.log(info.file, info.fileList)
      }
      if (info.file.status === 'done') {
        message.success(`${info.file.name} file uploaded successfully`)
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} file upload failed.`)
      }
    },
  }
  return (
    <>
      <div className="flex ">
        <Upload {...props}>
          {files?.length === 0 && <Button icon={<MdOutlineUploadFile />}>上传知识文件</Button>}
        </Upload>
        <Button type="primary" className="mx-2">
          智能生成规则
        </Button>
      </div>

      {files?.length > 0 && (
        <div className="flex  mt-3 items-center">
          <Input.TextArea
            className="w-full mr-5"
            value={`如果某个节点发生异常，但是上游节点未发生接口异常告警，可以将这个节点及其分支全部排除`}
          ></Input.TextArea>
          <Button type="primary">确认规则</Button>
        </div>
      )}
    </>
  )
}
const RuleConfigItem = () => {
  return (
    <>
      <UploadCom />
    </>
  )
}

const AlertRuleConfig = () => {
  const [modalVisible, setModalVisible] = useState(false)
  const [items, setItems] = useState([
    {
      title: '规则 1',
      description: (
        <Input.TextArea
          autoSize
          disabled
          className="w-full mt-3"
          value={
            '告警异常类型 1接口层告警,2为容器异常,3为基础设施异常,4为网络程序异常,5为错误异常。没有告警事件的节点可以排除。'
          }
        ></Input.TextArea>
      ),
    },
    {
      title: '规则 2',
      description: (
        <Input.TextArea
          autoSize
          disabled
          className="w-full mt-3"
          value={`从应用接口层告警（如 URL 接口延时异常、错误异常）出发。
    沿着业务入口的拓扑结构，从上游向下游追踪（使用拓扑图数据，不要搞错上下游）。
    对于每个节点：
    如果存在接口层告警，继续向其依赖的下游节点追踪，如果没有下游依赖节点，停止追踪。
    如果没有接口层告警，停止追踪，返回上一级节点，当前节点可以排除。
    在追踪过程中，对于每个有告警的节点：
    检查是否存在其他类型的告警（如应用层、容器层、基础设施层告警）。
    如果存在，则该节点可能是根因，记录并停止追踪。
    如果不存在其他告警，继续分析下游节点。
    最下游节点定义：依赖链路上没有进一步下游依赖的节点。最下游节点如果有告警，是疑似根因的优先候选。`}
        ></Input.TextArea>
      ),
    },
    {
      title: '规则 3',
      description: (
        <Input.TextArea
          autoSize
          disabled
          className="w-full mt-3"
          value={`根因节点需要满足异常从该节点开始向上游传递递，同时该类型异常告警事件是持续发生的，而不是偶发的
    如果接口层类型的告警既有新增又有解决，该节点可认为是偶发。
    但是接口层类型告警存在存留告警，该节点应做为疑似根因节点。
    节点的异常告警是偶发，那么它不应该被认定为根因节点`}
        ></Input.TextArea>
      ),
    },
    {
      title: '规则 4',
      description: <RuleConfigItem />,
    },
    {
      title: '添加新规则',
      icon: (
        <Button
          type="primary"
          onClick={() => addMoreSteps()}
          icon={<IoMdAdd size={20} />}
          shape="circle"
        ></Button>
      ),
      //   description:
    },
  ])
  const addMoreSteps = () => {
    setItems((prev) => {
      const result = [...prev]
      result.splice(prev.length - 2, 0, {
        title: '规则 ' + (prev.length - 1),
        description: <RuleConfigItem />,
      })
      return result
    })
  }
  return (
    <>
      <Button onClick={() => setModalVisible(true)} type="primary">
        配置规则
      </Button>
      <Modal
        title={'推理规则学习'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        destroyOnClose
        centered
        okText={'保存'}
        cancelText="取消"
        maskClosable={false}
        width={1000}
        footer={(_, { OkBtn }) => <></>}
        bodyStyle={{ maxHeight: '80vh', overflowY: 'auto', overflowX: 'hidden' }}
      >
        <Steps direction="vertical" current={3} items={items} />
      </Modal>
    </>
  )
}
export default AlertRuleConfig
