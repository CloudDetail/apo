import { Typography, Image, Table } from 'antd'
import Paragraph from 'antd/es/typography/Paragraph'
import Title from 'antd/es/typography/Title'
import Text from 'antd/es/typography/Text'
import CopyPre from './CopyPre'
const columns1 = [
  { dataIndex: 'title', title: '字段名', width: 150 },
  { dataIndex: 'meaning', title: '含义', width: 150 },
  { dataIndex: 'type', title: '字段类型', width: 150 },
  { dataIndex: 'description', title: '说明' },
]
const data1 = [
  {
    title: 'sourceId',
    meaning: '告警源编号',
    type: 'string',
    description: '使用在APO上创建数据信源时,提供的sourceId',
  },
]
const data2 = [
  {
    title: 'name',
    meaning: '告警事件名称',
    type: 'string',
    description: '告警事件的名称, 比如Zabbix的触发器名',
  },
  {
    title: 'status',
    meaning: '告警事件状态',
    type: 'string',
    description: '表示告警当前是触发还是解决，枚举值: firing: 正在发生, resolved: 已经结束',
  },
  {
    title: 'severity',
    meaning: '告警事件严重程度',
    type: 'string',
    description:
      '告警的严重程度，枚举值: critical: 严重, error: 错误, warning: 警告, info: 信息, unknown: 未知',
  },
  {
    title: 'detail',
    meaning: '告警事件内容',
    type: 'string',
    description: '告警的详细内容, 文本格式',
  },
  {
    title: 'alertId',
    meaning: '告警事件编号',
    type: 'string',
    description:
      '告警的编号,相同的告警指同一个告警随时间变化产生了多个事件,这些事件都是用相同的编号;如果原始数据携带了编号信息,则直接使用;否则使用告警事件上下文中的所有标签进行计算',
  },
  {
    title: 'tags',
    meaning: '告警事件标签',
    type: 'map[string]string',
    description: '告警自身携带的信息,输入后不做更改',
  },
  {
    title: 'startTime',
    meaning: '告警事件发生时间',
    type: 'int64',
    description:
      '告警事件从告警源生成的时间,相同告警事件重复告警时,会使用第一次发生告警事件的时间，格式: 2025-01-21T15:04:05+00:00或以数字传入毫秒级时间戳(例如1737514800000)',
  },
  {
    title: 'updateTime',
    meaning: '告警事件更新时间',
    type: 'int64',
    description: '告警再次发生的时间, 格式同startTime',
  },
  {
    title: 'endTime',
    meaning: '告警事件结束时间',
    type: 'int64',
    description: '通常只有解决的事件会包含告警事件的结束时间, 格式同startTime',
  },
]
const code1 = `{
    "name": "Zabbix-Trigger-1",
    "status": "trigger",
    "detail": "Zabbix-Trigger-1",
    "alertId": "1234567890",
    "tags": {
        "tag1": "value1",
        "tag2": "value2"
    },
    "startTime": "2025-01-21T15:04:05+00:00",
    "updateTime": 1737514800000,
    "endTime": 1737514800000,
    "severity": "error",
    "status": "firing"
}`
const code2 = `{
    "code": "B1319",
    "message": "处理告警事件失败"
}
`
const JsonInfo = () => {
  return (
    <>
      <Typography>
        <Title level={4}>标准告警源</Title>
        <Text>对于还未进行适配的告警源,可以尝试通过标准告警事件输入接口进行集成.</Text>
        <Title level={5}>接口信息</Title>
        <Typography className="mb-2">
          <Text strong>请求方式</Text>
          <Typography>POST</Typography>
        </Typography>
        <Typography className="mb-2">
          <Text strong>请求头</Text>
          <Typography>Content-Type: application/json</Typography>
        </Typography>
        <Typography className="mb-2">
          <Text strong>请求参数</Text>
          <Table columns={columns1} dataSource={data1} pagination={false}></Table>
        </Typography>
        <Typography className="mb-2">
          <Text strong>请求Body</Text>
          <Table columns={columns1} dataSource={data2} pagination={false}></Table>
        </Typography>
        <Title level={5}>请求示例</Title>
        <Typography className="mb-2">
          <CopyPre code={code1}></CopyPre>
        </Typography>
        <Title level={5}>响应</Title>
        <Typography className="mb-2">
          <Text strong>成功响应</Text>
          <Typography>200 "ok"</Typography>
        </Typography>
        <Typography className="mb-2">
          <Text strong>失败响应</Text>
          <CopyPre code={code2}></CopyPre>
        </Typography>
      </Typography>
    </>
  )
}
export default JsonInfo
